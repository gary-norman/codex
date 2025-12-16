# ErrorMsgs to Structured Logging Migration Guide

## Overview

Migrate from global `ErrorMsgs` templates to structured logging with proper error wrapping.

## Core Principles

1. **Separation of Concerns**:
   - **Errors** = control flow (return from functions)
   - **Logging** = observability (what happened)

2. **Error Wrapping**:
   - Use `fmt.Errorf("context: %w", err)` to preserve error chains
   - Allows tracing root cause through call stack

3. **Logging Levels**:
   - `LogError()` - Actual errors that need attention
   - `LogWarn()` - Recoverable issues, partial failures
   - `LogInfo()` - Normal operations, success states

## Migration Patterns

### Pattern 1: Database Query Errors

**Before:**
```go
stmt, insertErr := m.DB.Prepare("INSERT INTO ...")
if insertErr != nil {
    log.Printf(ErrorMsgs.Query, username, insertErr)
}
defer func(stmt *sql.Stmt) {
    closErr := stmt.Close()
    if closErr != nil {
        log.Printf(ErrorMsgs.Close, "stmt", "insert", closErr)
    }
}(stmt)
_, err := stmt.Exec(id, username, email, ...)
if err != nil {
    return err
}
return nil
```

**After:**
```go
stmt, err := m.DB.Prepare("INSERT INTO ...")
if err != nil {
    return fmt.Errorf("failed to prepare insert for user %s: %w", username, err)
}
defer func() {
    if closeErr := stmt.Close(); closeErr != nil {
        LogWarn("Failed to close prepared statement: %v", closeErr)
    }
}()

if _, err := stmt.Exec(id, username, email, ...); err != nil {
    return fmt.Errorf("failed to insert user %s: %w", username, err)
}

LogInfo("User created: %s", username)
return nil
```

**What changed:**
- ✅ Return wrapped errors instead of just logging
- ✅ Caller can decide how to handle
- ✅ Error chain preserved with `%w`
- ✅ Success logged with `LogInfo()`
- ✅ Non-critical errors (close) logged with `LogWarn()`

### Pattern 2: Row Scanning Errors

**Before:**
```go
rows, err := m.DB.Query("SELECT * FROM Users")
if err != nil {
    log.Printf(ErrorMsgs.Query, "Users", err)
    return nil, err
}
defer func(rows *sql.Rows) {
    closeErr := rows.Close()
    if closeErr != nil {
        log.Printf(ErrorMsgs.Close, rows, "All", closeErr)
    }
}(rows)

for rows.Next() {
    var user models.User
    if err := rows.Scan(&user.ID, &user.Username, ...); err != nil {
        log.Printf(ErrorMsgs.Query, "scan user", err)
        continue
    }
    users = append(users, &user)
}
return users, nil
```

**After:**
```go
rows, err := m.DB.Query("SELECT * FROM Users")
if err != nil {
    return nil, fmt.Errorf("failed to query users: %w", err)
}
defer func() {
    if closeErr := rows.Close(); closeErr != nil {
        LogWarn("Failed to close rows: %v", closeErr)
    }
}()

var users []*models.User
for rows.Next() {
    var user models.User
    if err := rows.Scan(&user.ID, &user.Username, ...); err != nil {
        LogWarn("Failed to scan user row: %v", err)
        continue  // Skip malformed rows but continue processing
    }
    users = append(users, &user)
}

if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("row iteration error: %w", err)
}

LogInfo("Retrieved %d users", len(users))
return users, nil
```

**What changed:**
- ✅ Check `rows.Err()` after iteration
- ✅ Scan errors logged but processing continues
- ✅ Success state logged with count
- ✅ Clear error wrapping at each layer

### Pattern 3: Update/Edit Operations

**Before:**
```go
stmt, prepErr := m.DB.Prepare("UPDATE Users SET ...")
if prepErr != nil {
    log.Printf(ErrorMsgs.Query, "Users", prepErr)
}
defer func(stmt *sql.Stmt) {
    closErr := stmt.Close()
    if closErr != nil {
        log.Printf(ErrorMsgs.Close, "stmt", "Edit", closErr)
    }
}(stmt)

_, err := stmt.Exec(user.Username, user.Email, ..., user.ID)
if err != nil {
    log.Printf(ErrorMsgs.Update, user.Username, "Edit", err)
    return err
}
return nil
```

**After:**
```go
stmt, err := m.DB.Prepare("UPDATE Users SET ...")
if err != nil {
    return fmt.Errorf("failed to prepare user update: %w", err)
}
defer func() {
    if closeErr := stmt.Close(); closeErr != nil {
        LogWarn("Failed to close prepared statement: %v", closeErr)
    }
}()

result, err := stmt.Exec(user.Username, user.Email, ..., user.ID)
if err != nil {
    return fmt.Errorf("failed to update user %s: %w", user.Username, err)
}

rowsAffected, _ := result.RowsAffected()
if rowsAffected == 0 {
    LogWarn("User update affected 0 rows: %s", user.Username)
}

LogInfo("User updated: %s", user.Username)
return nil
```

**What changed:**
- ✅ Check `RowsAffected()` for unexpected results
- ✅ Warn if update didn't affect any rows
- ✅ Log success with user identifier

### Pattern 4: Context-Aware Logging (HTTP Handlers)

**Before:**
```go
func (h *HomeHandler) GetHome(w http.ResponseWriter, r *http.Request) {
    posts, err := h.App.Posts.All()
    if err != nil {
        log.Printf(ErrorMsgs.Query, "posts", err)
        return
    }
    // ...
}
```

**After:**
```go
func (h *HomeHandler) GetHome(w http.ResponseWriter, r *http.Request) {
    posts, err := h.App.Posts.All()
    if err != nil {
        models.LogErrorWithContext(r.Context(), "Failed to fetch posts", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

**What changed:**
- ✅ Use `LogErrorWithContext()` to include request ID
- ✅ Return HTTP error to client
- ✅ Request is traceable through logs

## Quick Reference

### When to use each function:

| Old ErrorMsgs | New Function | When to Use |
|---------------|--------------|-------------|
| `ErrorMsgs.Query` | `fmt.Errorf("...: %w", err)` | Return wrapped error |
| `ErrorMsgs.Close` | `LogWarn("Failed to close: %v", err)` | Non-critical cleanup |
| `ErrorMsgs.Insert` | `fmt.Errorf("...: %w", err)` | Return wrapped error |
| `ErrorMsgs.Update` | `fmt.Errorf("...: %w", err)` | Return wrapped error |
| `ErrorMsgs.NotFound` | `fmt.Errorf("user not found: %w", err)` | Return wrapped error |
| `ErrorMsgs.KeyValuePair` | `LogInfo("%s: %v", key, value)` | Debug logging |

### Error Wrapping Decision Tree:

```
Is this an error condition?
├─ YES → Return fmt.Errorf("context: %w", err)
│        └─ Let caller decide: log, retry, or fail
│
└─ NO → Is this important to track?
        ├─ YES → Is it successful?
        │        ├─ YES → LogInfo("Operation succeeded")
        │        └─ NO → LogWarn("Partial failure")
        │
        └─ NO → Don't log (avoid noise)
```

## Common Mistakes to Avoid

❌ **Logging AND returning the same error**
```go
if err != nil {
    LogError("Database error", err)  // Don't do this
    return err                        // Caller will log again
}
```

✅ **Return wrapped errors, let caller log**
```go
if err != nil {
    return fmt.Errorf("failed to query: %w", err)  // Caller decides
}
```

❌ **Swallowing errors silently**
```go
if err != nil {
    // Nothing - error disappears!
}
```

✅ **Log if you can't return**
```go
if err != nil {
    LogWarn("Non-critical error in cleanup: %v", err)
}
```

❌ **Using generic error messages**
```go
return fmt.Errorf("database error: %w", err)  // Too vague
```

✅ **Adding context at each layer**
```go
return fmt.Errorf("failed to create user %s: %w", username, err)
```

## Benefits

1. **Testable**: Can check error messages in unit tests
2. **Traceable**: Error chains show full context
3. **Flexible**: Caller decides how to handle
4. **Request-aware**: Context propagates request IDs
5. **Type-safe**: No more `%v` everywhere
6. **Performance**: Less string formatting
7. **Maintainable**: Clear separation of concerns
