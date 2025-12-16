# ErrorMsgs Migration Practice Guide

## Overview

This guide will help you complete the remaining ErrorMsgs migration instances as hands-on practice. I've added `TODO(human)` comments throughout the codebase to guide you through each migration.

**Your Goal**: Migrate 24 ErrorMsgs instances across 5 files from the old template-based logging system to the new structured logging system.

## Files to Practice On

| File | Instances | Difficulty |
|------|-----------|------------|
| `internal/sqlite/cookies-sql.go` | 9 | Medium |
| `internal/sqlite/moderators-sql.go` | 6 | Easy |
| `internal/sqlite/rules-sql.go` | 5 | Easy |
| `internal/sqlite/memberships-sql.go` | 2 | Easy |
| `internal/sqlite/channels-sql.go` | 2 | Easy |

## How to Find Your Work

Search for `TODO(human)` in any of the files above. Each TODO comment includes:
- **Pattern**: The migration pattern to follow
- **Example**: Specific code to write
- **Reference**: Line numbers in completed files showing similar migrations

## Quick Reference Patterns

### Pattern 1: Database Query Errors (Return Wrapped Error)
```go
// ❌ OLD (ErrorMsgs template)
if queryErr != nil {
    return nil, fmt.Errorf(ErrorMsgs.Query, "Posts", queryErr)
}

// ✅ NEW (Wrapped error)
if queryErr != nil {
    return nil, fmt.Errorf("failed to query all posts: %w", queryErr)
}
```
**Key**: Use `%w` to wrap the error, add descriptive context, remove the log call entirely.

### Pattern 2: Defer Cleanup Errors (Log Warning)
```go
// ❌ OLD (ErrorMsgs template)
defer func() {
    if closeErr := rows.Close(); closeErr != nil {
        log.Printf(ErrorMsgs.Close, "rows", "All")
    }
}()

// ✅ NEW (Structured logging)
defer func() {
    if closeErr := rows.Close(); closeErr != nil {
        models.LogWarn("Failed to close rows: %v", closeErr)
    }
}()
```
**Key**: Use `models.LogWarn` for non-critical cleanup errors in defer blocks.

### Pattern 3: Operational Errors (Log with Context)
```go
// ❌ OLD (ErrorMsgs template)
if err != nil {
    log.Printf(ErrorMsgs.Cookies, "query", err)
}

// ✅ NEW (Structured logging)
if err != nil {
    models.LogError("Failed to query cookie expiration", err, "Username:", user.Username)
}
```
**Key**: Use `models.LogError` with descriptive message and relevant context values.

### Pattern 4: Debug/Info Logging
```go
// ❌ OLD (ErrorMsgs template)
log.Printf(ErrorMsgs.KeyValuePair, "Cookie SessionToken", st.Value)

// ✅ NEW (Structured logging)
models.LogInfo("Cookie SessionToken: %s", st.Value)
```
**Key**: Use `models.LogInfo` for informational debug logs.

## Step-by-Step Workflow

### Step 1: Start with an Easy File
Begin with `memberships-sql.go` (only 2 instances):

1. Open the file
2. Search for `TODO(human)`
3. Read the TODO comment guidance
4. Replace the ErrorMsgs line with the new pattern
5. Remove the TODO comment when done

### Step 2: Verify Your Changes
After each file, verify the build:
```bash
make build
```

### Step 3: Reference Completed Work
If stuck, check these completed migrations:
- **Database layer patterns**: `internal/sqlite/posts-sql.go`, `internal/sqlite/comments-sql.go`, `internal/sqlite/users-sql.go`
- **Defer cleanup**: Look for any `defer` block in the files above
- **Error wrapping**: See any function returning errors in the completed files

### Step 4: Remove Unused Import
After migrating all ErrorMsgs in a file, remove the `log` import if it's no longer used:
```go
import (
    "database/sql"
    "fmt"
    "log"  // ← Remove this if no longer used

    "github.com/gary-norman/forum/internal/models"
)
```

## Common Mistakes to Avoid

### ❌ Mistake 1: Using Wrong Error Verb in fmt.Errorf
```go
// WRONG - Uses %v instead of %w
return fmt.Errorf("failed to query: %v", err)

// CORRECT - Uses %w for error wrapping
return fmt.Errorf("failed to query: %w", err)
```

### ❌ Mistake 2: Logging AND Returning in Database Layer
```go
// WRONG - Logs and returns
if err != nil {
    models.LogError("Failed to insert", err)
    return err
}

// CORRECT - Only returns (caller decides whether to log)
if err != nil {
    return fmt.Errorf("failed to insert post: %w", err)
}
```

### ❌ Mistake 3: Passing Wrong Arguments to LogWarn
```go
// WRONG - Passing rows object (wrong type)
models.LogWarn("Failed to close rows: %v", rows, closeErr)

// CORRECT - Only pass the error
models.LogWarn("Failed to close rows: %v", closeErr)
```

## Detailed File Guides

### cookies-sql.go (9 instances)
This file handles cookie authentication. Key challenges:
- Multiple debug log statements (lines 131-138) - convert to `models.LogInfo`
- Database queries - use `models.LogError`
- One nil check - log before returning error

**Practice Focus**: Converting debug logging and handling authentication-specific context.

### moderators-sql.go (6 instances)
This file manages channel moderators. Key challenges:
- Fix existing `fmt.Errorf` calls that use ErrorMsgs templates
- Defer cleanup in multiple functions
- Parse functions that currently log + return (should only return)

**Practice Focus**: Fixing incorrect error wrapping patterns.

### rules-sql.go (5 instances)
This file manages channel rules. Key challenges:
- Multiple defer blocks (rows and prepared statements)
- Error wrapping in query functions
- `rows.Err()` error handling

**Practice Focus**: Defer cleanup and error iteration patterns.

### memberships-sql.go (2 instances)
Simplest file - just defer cleanup. Great starting point!

**Practice Focus**: Basic defer pattern practice.

### channels-sql.go (2 instances)
- One defer cleanup
- One insert error (log + return → just return)

**Practice Focus**: Removing redundant logging in database operations.

## How to Verify You're Done

1. **Search for remaining ErrorMsgs**:
   ```bash
   grep -r "ErrorMsgs" internal/sqlite/cookies-sql.go internal/sqlite/moderators-sql.go internal/sqlite/rules-sql.go internal/sqlite/memberships-sql.go internal/sqlite/channels-sql.go
   ```
   Should return no results (except in comments/TODOs).

2. **Search for TODO(human)**:
   ```bash
   grep -r "TODO(human)" internal/sqlite/
   ```
   Should return no results when you're finished.

3. **Build verification**:
   ```bash
   make build
   ```
   Should complete successfully with no errors.

4. **Check imports**:
   Ensure you removed unused `log` imports from files where you migrated all log.Printf calls.

## Learning Objectives

By completing this practice, you'll gain experience with:
- ✅ **Error Wrapping**: Using `%w` for proper error chain preservation
- ✅ **Separation of Concerns**: Database layer returns errors, handlers log them
- ✅ **Logging Levels**: When to use LogInfo vs LogWarn vs LogError
- ✅ **Context Propagation**: Adding meaningful context to logs for debugging
- ✅ **Defer Cleanup Patterns**: Handling non-critical cleanup errors
- ✅ **Code Archaeology**: Reading existing patterns and applying them consistently

## Getting Help

If you get stuck:
1. Read the ERROR_SYSTEM_MIGRATION.md guide (comprehensive migration patterns)
2. Look at the TODO(human) comment for that specific instance
3. Check the referenced line numbers in completed files
4. Search for similar patterns in completed migrations
5. Verify the build after each change to catch errors early

## Final Notes

- Take your time - understanding the patterns is more important than speed
- Each TODO comment is designed to guide you through the specific case
- The build will tell you immediately if you made a syntax error
- These patterns will appear throughout the codebase, so learning them well will help in future work

**Estimated Time**: 30-45 minutes for all 24 instances

Good luck! 🚀
