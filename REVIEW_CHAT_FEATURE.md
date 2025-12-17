# Chat Creation Feature - Code Review Document

## Overview

This document reviews the **chat creation UI and infrastructure** added to the `ws/main` branch. This feature allows users to create one-on-one ("buddy") chats directly from the sidebar.

**Branch**: `ws/main`
**Commits**: 5 commits since last merge (93108f4, 56a0bc4, a6d0d44, 3ece419, 4ce3164)
**Status**: Ready for review and merge to `main`

---

## What Was Built

### User-Facing Features

1. **"Start New Chat" UI** in sidebar (`assets/templates/chat-controls.tmpl`)
   - Shows all users in the system (except current user)
   - Click any username to create/open a chat with that user
   - Page reloads after successful chat creation
   - Chat appears in the user's chat list

2. **Idempotent Chat Creation**
   - If chat already exists between two users, returns existing chat
   - If no chat exists, creates new buddy chat
   - Automatically attaches both users to the chat

3. **Error Handling**
   - Inline notifications for failures (uses existing `showInlineNotification` system)
   - Proper HTTP status codes (200 for success/existing, 201 for new, 500 for errors)

### Technical Implementation

#### 1. Frontend Changes

**New File: `assets/js/chat.js`** (122 lines)
- `setupChatFormHandlers()`: Handles sending messages via WebSocket
- `setupStartChatHandlers()`: Handles "Start New Chat" button clicks
- Integrates with existing `showInlineNotification` system
- Exports functions for use in `main.js`

**Modified: `assets/js/main.js`**
- Imports and initializes chat handlers on page load

**Modified: `assets/templates/chat-controls.tmpl`**
- Added new "Start New Chat" section
- Iterates over all users to create clickable buttons
- Uses `data-user-id` and `data-username` attributes for JavaScript

**Modified: `assets/templates/index.html`**
- Removed test WebSocket UI div (lines 461-468)

#### 2. Backend Changes

**New File: `internal/http/handlers/chat-handlers.go`** (92 lines)
- `ChatHandler` struct with `App *app.App` dependency
- `CreateChat(w, r)` endpoint implementation:
  1. Authenticates user via context
  2. Parses buddy ID from JSON request body
  3. Checks if chat already exists via `GetBuddyChatID()`
  4. Creates new chat if doesn't exist
  5. Attaches both users to the chat
  6. Returns JSON response with chat ID and `exists` flag

**Modified: `internal/http/routes/routes.go`**
- Added route: `POST /api/chats/create` with `WithUser` middleware

**Modified: `internal/http/routes/registry.go`**
- Added `Chat *h.ChatHandler` field to `RouteHandler` struct
- Added `NewChatHandler()` constructor function
- Wired `ChatHandler` into dependency injection system

#### 3. Database Layer Changes

**Modified: `internal/sqlite/chats-sql.go`**

**Function signature change (BREAKING)**:
```go
// OLD
func (c *ChatModel) CreateChat(ctx, chatType, name string, groupID, buddyID models.UUIDField)

// NEW
func (c *ChatModel) CreateChat(ctx, chatType, name string, groupID, buddyID models.NullableUUIDField)
```

**Why this changed**: The database CHECK constraint requires:
- Buddy chats: `BuddyID IS NOT NULL AND GroupID IS NULL`
- Group chats: `GroupID IS NOT NULL AND BuddyID IS NULL`

Passing `models.ZeroUUIDField()` inserts `00000000-0000-0000-0000-000000000000` (NOT NULL).
Passing `models.NullableUUIDField{Valid: false}` inserts SQL `NULL`.

**New method: `GetBuddyChatID()`** (lines 325-343)
```go
func (c *ChatModel) GetBuddyChatID(ctx, user1ID, user2ID models.UUIDField) (models.UUIDField, error)
```
- Queries for existing buddy chat between two users
- Returns chat ID if exists, error if not found
- Used to prevent duplicate chats

**Transaction removal from `GetUserChats()`**:
- Removed unnecessary transaction wrapper (causing "nested transaction" errors)
- Now queries directly via `c.DB.QueryContext()`
- Read-only operations don't need transactions

#### 4. Model Changes

**Modified: `internal/models/uuidfield-models.go`**

**Fixed `NullableUUIDField.Scan()` method** (lines 113-125):
```go
// OLD - Broken (used uuid.ParseBytes which expects string format)
case []byte:
    parsed, err := uuid.ParseBytes(v)
    if err != nil {
        return err
    }
    u.UUID = UUIDField{UUID: parsed}

// NEW - Fixed (delegates to UUIDField.Scan for BLOB handling)
if value == nil {
    u.Valid = false
    return nil
}
err := u.UUID.Scan(value)
if err != nil {
    return err
}
u.Valid = true
```

**Why this matters**: SQLite stores UUIDs as 16-byte BLOBs. `UUIDField.Scan()` correctly copies these bytes via `copy(u.UUID[:], v)`. The old code tried to parse them as string UUIDs, causing "invalid UUID length: 16" errors.

#### 5. WebSocket Infrastructure (Existing, Not Changed This Session)

These files support the chat system but were implemented in previous commits:
- `internal/http/websocket/client.go`: WebSocket client management with user identification
- `internal/http/websocket/manager.go`: Message broadcasting to chat participants
- `internal/http/websocket/otp.go`: One-time passwords for WebSocket authentication
- `assets/js/websocket.js`: Client-side WebSocket connection and message handling

---

## Database Schema (Existing)

The Chats table (from `migrations/004_chats.sql`):

```sql
CREATE TABLE IF NOT EXISTS Chats (
    ID BLOB PRIMARY KEY,
    Type TEXT NOT NULL CHECK (Type IN ('buddy', 'group')),
    Name TEXT,
    GroupID BLOB,
    BuddyID BLOB,
    Created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    LastActive DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (BuddyID) REFERENCES Users(ID) ON DELETE SET NULL,
    CHECK (
        (Type = 'buddy' AND BuddyID IS NOT NULL AND GroupID IS NULL) OR
        (Type = 'group' AND GroupID IS NOT NULL AND BuddyID IS NULL)
    )
);

CREATE TABLE IF NOT EXISTS ChatUsers (
    ChatID BLOB NOT NULL,
    UserID BLOB NOT NULL,
    PRIMARY KEY (ChatID, UserID),
    FOREIGN KEY (ChatID) REFERENCES Chats(ID) ON DELETE CASCADE,
    FOREIGN KEY (UserID) REFERENCES Users(ID) ON DELETE CASCADE
);
```

**Key constraint**: The CHECK constraint enforces mutual exclusivity between buddy and group chats.

---

## API Endpoints

### POST /api/chats/create

**Authentication**: Requires authenticated user (via `WithUser` middleware)

**Request Body**:
```json
{
  "buddy_id": "uuid-string"
}
```

**Success Response (Existing Chat)**:
```json
{
  "chat_id": "uuid-string",
  "exists": true
}
```
Status: 200 OK

**Success Response (New Chat)**:
```json
{
  "chat_id": "uuid-string",
  "exists": false
}
```
Status: 201 Created

**Error Responses**:
- 401 Unauthorized: User not authenticated
- 400 Bad Request: Invalid buddy_id format
- 500 Internal Server Error: Database error

---

## How to Test

### Prerequisites
1. Two browser windows/profiles logged in as different users
2. Server running on port 8888

### Test Steps

1. **Basic Chat Creation**
   - User A: Click on User B's name in "Start New Chat" section
   - Expected: Page reloads, chat appears in sidebar
   - User B: Refresh page
   - Expected: Same chat appears in their sidebar

2. **Idempotency Test**
   - User A: Click User B's name again
   - Expected: Page reloads, same chat (not duplicate)
   - Check database: `SELECT COUNT(*) FROM Chats WHERE Type = 'buddy';`
   - Expected: Only 1 chat between the users

3. **Multiple Chats**
   - User A: Click User C's name
   - Expected: New separate chat created
   - User A should now have 2 chats in sidebar

4. **Error Handling**
   - Stop server
   - User A: Try to create chat
   - Expected: Inline error notification in chat title area
   - No browser alert/console error spam

### Database Verification

```bash
# Check chats were created properly
sqlite3 /var/lib/db-codex/dev_forum_database.db "
SELECT
    hex(ID) as ChatID,
    Type,
    hex(BuddyID) as BuddyID,
    GroupID
FROM Chats
WHERE Type = 'buddy';
"

# Check user attachments
sqlite3 /var/lib/db-codex/dev_forum_database.db "
SELECT
    hex(ChatID),
    hex(UserID)
FROM ChatUsers;
"
```

Expected:
- BuddyID should be 16-byte hex (NOT NULL)
- GroupID should be empty/NULL
- Each chat should have 2 entries in ChatUsers

---

## Breaking Changes

### ⚠️ Database Method Signature Change

**Impact**: Any code calling `CreateChat` must be updated.

**Old signature**:
```go
CreateChat(ctx context.Context, chatType, name string, groupID, buddyID models.UUIDField)
```

**New signature**:
```go
CreateChat(ctx context.Context, chatType, name string, groupID, buddyID models.NullableUUIDField)
```

**Migration example**:
```go
// OLD CODE (will not compile)
chatID, err := app.Chats.CreateChat(ctx, "buddy", "",
    models.ZeroUUIDField(),  // groupID
    buddyUUID)               // buddyID

// NEW CODE (correct)
chatID, err := app.Chats.CreateChat(ctx, "buddy", "",
    models.NullableUUIDField{Valid: false},              // groupID = NULL
    models.NullableUUIDField{UUID: buddyUUID, Valid: true}) // buddyID = value
```

**For group chats**:
```go
chatID, err := app.Chats.CreateChat(ctx, "group", "Group Name",
    models.NullableUUIDField{UUID: groupUUID, Valid: true},  // groupID = value
    models.NullableUUIDField{Valid: false})                  // buddyID = NULL
```

---

## Files Changed Summary

### New Files
- `assets/js/chat.js` (122 lines)
- `internal/http/handlers/chat-handlers.go` (92 lines)

### Modified Files
- `assets/js/main.js` (+1 import)
- `assets/templates/chat-controls.tmpl` (+14 lines)
- `assets/templates/index.html` (-7 lines, removed test UI)
- `internal/http/routes/routes.go` (+1 route)
- `internal/http/routes/registry.go` (+8 lines, wired ChatHandler)
- `internal/sqlite/chats-sql.go` (signature change, transaction removal, +1 method)
- `internal/models/uuidfield-models.go` (fixed Scan method)
- `CLAUDE.md` (+35 lines documentation)

### WebSocket Files (From Previous Work)
- `assets/js/websocket.js`
- `assets/templates/chat-popover.tmpl`
- `internal/http/handlers/auth-handlers.go`
- `internal/http/middleware/logging_enhanced.go`
- `internal/http/websocket/*`

---

## How to Create Pull Request

### Step 1: Verify Current State

```bash
# Ensure you're on ws/main
git branch --show-current
# Should output: ws/main

# Verify all commits are pushed
git log --oneline -5
# Should see:
# 93108f4 feat: implement WebSocket real-time messaging infrastructure
# 56a0bc4 docs: update architecture documentation for chat system
# a6d0d44 refactor: remove test WebSocket UI from index page
# 3ece419 fix: update CreateChat to use NullableUUIDField for SQL NULL support
# 4ce3164 feat: add chat creation UI and API endpoint
```

### Step 2: Create Pull Request

**Via GitHub Web Interface**:

1. Navigate to: https://learn.01founders.co/git/gnorman/forum
2. You should see a banner: "ws/main had recent pushes"
3. Click "Compare & pull request" (or "New pull request")
4. Set **base**: `main` ← **compare**: `ws/main`
5. **Title**: "Chat Creation Feature - User-to-User Messaging UI"
6. **Description**: (paste below)

```markdown
## Overview
Adds UI for creating one-on-one chats between users.

## Features
- "Start New Chat" section in sidebar with all users
- One-click chat creation (idempotent - won't create duplicates)
- Proper NULL handling for database CHECK constraints
- Inline error notifications

## Technical Changes
- **New**: ChatHandler with POST /api/chats/create endpoint
- **New**: chat.js for frontend chat management
- **Fixed**: NullableUUIDField.Scan() to handle 16-byte BLOBs correctly
- **Fixed**: Removed nested transaction from GetUserChats
- **Breaking**: CreateChat now requires NullableUUIDField (see REVIEW_CHAT_FEATURE.md)

## Testing
- Tested with multiple users in separate browsers
- Database constraints verified (buddy chats have NULL GroupID)
- Idempotency confirmed (clicking same user twice doesn't create duplicates)

## Documentation
See REVIEW_CHAT_FEATURE.md for full technical details and testing instructions.
```

7. **Reviewers**: Add team member(s)
8. Click "Create pull request"

**Via GitHub CLI** (alternative):

```bash
gh pr create \
  --base main \
  --head ws/main \
  --title "Chat Creation Feature - User-to-User Messaging UI" \
  --body-file REVIEW_CHAT_FEATURE.md
```

### Step 3: Important - Merge Strategy

**⚠️ CRITICAL**: To keep the codebase EXACTLY as it is on `ws/main`:

**DO NOT use "Squash and merge"** - This will rewrite history
**DO NOT use "Rebase and merge"** - This will rewrite commits

**✅ USE "Create a merge commit"** or **"Merge pull request"**

This preserves all 5 commits exactly as they are, maintaining the commit history.

**In GitHub UI**:
1. When ready to merge, click the dropdown next to "Merge pull request"
2. Select "Create a merge commit" (NOT "Squash and merge")
3. Click "Confirm merge"

**Via CLI**:
```bash
# After PR is approved
gh pr merge <PR-NUMBER> --merge
```

### Step 4: After Merge

```bash
# Switch to main and pull the merge
git checkout main
git pull origin main

# Verify the commits are there
git log --oneline -10
# Should see all 5 commits plus a merge commit

# Delete local ws/main branch (optional)
git branch -d ws/main

# Delete remote ws/main branch (optional)
git push origin --delete ws/main
```

---

## Common Issues & Solutions

### Issue: "CHECK constraint failed" when creating chat

**Symptom**: 500 error with message about buddy/group constraint

**Cause**: Passing `UUIDField` instead of `NullableUUIDField` to `CreateChat`

**Solution**: Update code to use `NullableUUIDField{Valid: false}` for NULL values

### Issue: "cannot start a transaction within a transaction"

**Symptom**: Chats don't appear in sidebar after creation

**Cause**: GetUserChats was wrapped in unnecessary transaction

**Solution**: Already fixed in commit 3ece419 (transaction removed)

### Issue: "invalid UUID length: 16"

**Symptom**: Error when fetching user chats

**Cause**: NullableUUIDField.Scan() using uuid.ParseBytes (expects string, got bytes)

**Solution**: Already fixed in commit 3ece419 (delegates to UUIDField.Scan)

### Issue: Chat created but doesn't appear

**Symptom**: POST returns 201 but chat not in sidebar

**Cause**: Page may have cached chat list

**Solution**: Hard refresh browser (Cmd+Shift+R / Ctrl+Shift+F5)

---

## Performance Considerations

### Database Queries
- `GetBuddyChatID`: Uses INNER JOIN with 2 user lookups - indexed by ChatID and UserID
- `GetUserChats`: Single JOIN query, ordered by LastActive - scales O(n) with user's chat count
- `CreateChat` + `AttachUserToChat`: 3 INSERT operations (1 chat + 2 user attachments) - wrapped in HTTP request, not transaction

### Frontend
- Sidebar shows ALL users - may need pagination if user count > 100
- Page reload after chat creation - could be optimized to update DOM instead
- WebSocket connection persists after chat creation (no reconnection needed)

---

## Security Notes

1. **Authentication**: Chat creation requires authenticated user (WithUser middleware)
2. **Authorization**: Any user can create chat with any other user (no permission check)
3. **Input Validation**: Buddy ID validated as UUID format before database query
4. **SQL Injection**: Protected by parameterized queries (uses `?` placeholders)
5. **XSS**: User names rendered in templates are auto-escaped by Go's html/template

**Potential Enhancement**: Add privacy settings (e.g., "allow messages from: everyone | following | nobody")

---

## Future Improvements

1. **No Page Reload**: Update chat list dynamically via JavaScript instead of `window.location.reload()`
2. **User Search**: Add search bar in "Start New Chat" section for large user lists
3. **User Status**: Show online/offline indicators next to usernames
4. **Unread Counts**: Show unread message count per chat
5. **Group Chat UI**: Add interface for creating group chats (backend already supports it)
6. **Delete Chat**: Allow users to leave/delete chats
7. **Notifications**: Desktop/push notifications for new messages

---

## Questions for Reviewer

1. **UX**: Should chat creation update the UI without page reload?
2. **Privacy**: Should users be able to block chat requests?
3. **Naming**: Is "Start New Chat" clear, or should it be "Message a User"?
4. **Pagination**: At what user count should we paginate the user list?
5. **Group Chats**: Priority for group chat creation UI?

---

## Approval Checklist

- [ ] Code compiles and runs without errors
- [ ] Can create chat between two users
- [ ] Chat appears in both users' sidebars
- [ ] Clicking same user twice doesn't create duplicate
- [ ] Error messages display inline (not browser alerts)
- [ ] Database constraints verified (NULL handling correct)
- [ ] No console errors in browser DevTools
- [ ] Documentation (CLAUDE.md) updated
- [ ] Commit messages are clear and conventional
- [ ] No merge conflicts with main branch

---

**Reviewer**: After testing, please merge using "Create a merge commit" strategy to preserve all commit history exactly as is.
