# Logging Migration Guide

This document provides patterns for converting old logging to the new colorized logging system.

## Quick Reference

### 1. Import Changes

**Remove:**
```go
"log"
```

**Keep/Add (if not present):**
```go
"github.com/gary-norman/forum/internal/models"
```

---

## Conversion Patterns

### Pattern 1: Basic Error Logging

**Before:**
```go
log.Printf(ErrorMsgs.Encode, "search results", err)
log.Printf("Failed to parse form: %v", err)
```

**After:**
```go
models.LogErrorWithContext(r.Context(), "Failed to encode search results", err)
models.LogErrorWithContext(r.Context(), "Failed to parse form", err)
```

**Key points:**
- Use `LogErrorWithContext` for errors
- Pass `r.Context()` as first parameter
- Error object goes as second parameter (after message)
- Additional formatting args go after error

---

### Pattern 2: Warning Logging

**Before:**
```go
log.Printf("Search completed with errors: %v", err)
log.Printf("User not found: %s", username)
```

**After:**
```go
models.LogWarnWithContext(r.Context(), "Search completed with errors: %v", err)
models.LogWarnWithContext(r.Context(), "User not found: %s", username)
```

**When to use warnings:**
- Partial failures (operation continues)
- User not found / not authenticated
- Invalid input that doesn't break the request

---

### Pattern 3: Info Logging

**Before:**
```go
log.Printf("User %s logged in successfully", username)
fmt.Printf(Colors.Green+"Success!"+Colors.Reset)
```

**After:**
```go
models.LogInfoWithContext(r.Context(), "User %s logged in successfully", username)
models.LogInfoWithContext(r.Context(), "Success!")
```

**When to use info:**
- Successful operations
- User actions (login, logout, etc.)
- Normal flow events

---

### Pattern 4: Errors with Additional Context

**Before:**
```go
log.Printf(ErrorMsgs.NotFound, login, "login > GetUserFromLogin", getUserErr)
```

**After:**
```go
models.LogWarnWithContext(r.Context(), "User not found: %s", getUserErr, login)
```

**Pattern:**
```go
models.LogErrorWithContext(r.Context(), "message with %s placeholder", errorObject, contextValue)
```

---

### Pattern 5: Non-Request Context Logging

For helper functions without `*http.Request`:

**Before:**
```go
log.Printf("Error converting postID: %v", postID)
```

**After:**
```go
models.LogError("Failed to convert postID: %s", conversionErr, postIDStr)
models.LogWarn("Invalid postID format: %s", postIDStr)
models.LogInfo("Processing file: %s", filename)
```

**Note:** Use non-context versions (`LogError`, `LogWarn`, `LogInfo`) when you don't have `r.Context()`

---

### Pattern 6: Remove Redundant Timing Logs

**Remove entirely:**
```go
start := time.Now()
// ... handler code ...
log.Printf("[GET] /search - 200 (%dms)", time.Since(start).Milliseconds())
```

**Why:** The `LoggingEnhanced` middleware already logs all request timing automatically!

---

### Pattern 7: Conversion Error Logging

**Before:**
```go
postID, postConvErr := strconv.ParseInt(postIDStr, 10, 64)
if postConvErr != nil {
    log.Printf("Error converting postID: %v", postID)
}
```

**After:**
```go
postID, postConvErr := strconv.ParseInt(postIDStr, 10, 64)
if postConvErr != nil {
    models.LogWarnWithContext(r.Context(), "Failed to convert postID: %s", postConvErr, postIDStr)
}
```

**Why warning, not error:** Conversion failures are often expected input validation, not system errors.

---

### Pattern 8: fmt.Printf with Colors

**Before:**
```go
fmt.Printf(Colors.Green+"Passwords for %v match\n"+Colors.Reset, user.Username)
fmt.Printf(Colors.Red+"Error: %v\n"+Colors.Reset, err)
```

**After:**
```go
models.LogInfoWithContext(r.Context(), "Passwords for %s match", user.Username)
models.LogErrorWithContext(r.Context(), "Password verification failed", err)
```

**Why:** The new logging system handles colors automatically!

---

## Complete Examples

### Example 1: Search Handler (Before & After)

**Before:**
```go
func (s *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    result, err := ConcurrentSearch(r.Context(), s.App)
    if err != nil {
        log.Printf("Search completed with errors: %v", err)
    }

    currentUser, ok := mw.GetUserFromContext(r.Context())
    if !ok {
        log.Printf(ErrorMsgs.KeyValuePair, "User is not logged in. CurrentUser", currentUser)
    }

    if err := json.NewEncoder(w).Encode(searchResults); err != nil {
        log.Printf(ErrorMsgs.Encode, "search results", err)
        http.Error(w, "Error encoding search results", http.StatusInternalServerError)
        return
    }

    log.Printf("[GET] /search - 200 (%dms)", time.Since(start).Milliseconds())
}
```

**After:**
```go
func (s *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
    result, err := ConcurrentSearch(r.Context(), s.App)
    if err != nil {
        models.LogWarnWithContext(r.Context(), "Search completed with errors: %v", err)
    }

    currentUser, ok := mw.GetUserFromContext(r.Context())
    if !ok {
        models.LogInfoWithContext(r.Context(), "Anonymous user accessing search")
    } else {
        models.LogInfoWithContext(r.Context(), "User %s accessing search", currentUser.ID)
    }

    if err := json.NewEncoder(w).Encode(searchResults); err != nil {
        models.LogErrorWithContext(r.Context(), "Failed to encode search results", err)
        http.Error(w, "Error encoding search results", http.StatusInternalServerError)
        return
    }

    // Removed: timing log (middleware handles this)
}
```

---

### Example 2: Auth Handler Login (Before & After)

**Before:**
```go
user, getUserErr := h.App.Users.GetUserFromLogin(login, "login")
if getUserErr != nil {
    log.Printf(ErrorMsgs.NotFound, login, "login > GetUserFromLogin", getUserErr)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]any{
        "code": http.StatusUnauthorized,
        "message": "user not found",
    })
    return
}

if models.CheckPasswordHash(password, user.HashedPassword) {
    fmt.Printf(Colors.Green+"Passwords for %v match\n"+Colors.Reset, user.Username)
    // ... success handling ...
}
```

**After:**
```go
user, getUserErr := h.App.Users.GetUserFromLogin(login, "login")
if getUserErr != nil {
    models.LogWarnWithContext(r.Context(), "User not found: %s", getUserErr, login)
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]any{
        "code": http.StatusUnauthorized,
        "message": "user not found",
    })
    return
}

if models.CheckPasswordHash(password, user.HashedPassword) {
    models.LogInfoWithContext(r.Context(), "Successful login for user: %s", user.Username)
    // ... success handling ...
}
```

---

## Decision Tree

```
Does the log indicate an error condition?
├─ YES → Use LogErrorWithContext(r.Context(), msg, err, ...args)
│
└─ NO → Is it a warning/partial failure?
    ├─ YES → Use LogWarnWithContext(r.Context(), msg, ...args)
    │
    └─ NO → Is it informational?
        ├─ YES → Use LogInfoWithContext(r.Context(), msg, ...args)
        │
        └─ Is it timing/request info?
            └─ REMOVE IT (middleware logs this)
```

---

## Checklist for Each Handler File

- [ ] Remove `"log"` import
- [ ] Add `"github.com/gary-norman/forum/internal/models"` import (if not present)
- [ ] Replace all `log.Printf()` with appropriate `models.Log*WithContext()`
- [ ] Replace all `log.Println()` with appropriate `models.Log*WithContext()`
- [ ] Replace colored `fmt.Printf()` with `models.Log*WithContext()`
- [ ] Remove timing logs (`start := time.Now()` ... `time.Since(start)`)
- [ ] Verify all error logs have error object as 2nd parameter
- [ ] Run `make build` to verify
- [ ] Commit changes

---

## Common Mistakes to Avoid

❌ **Wrong order of parameters:**
```go
models.LogErrorWithContext(r.Context(), err, "Failed to save")  // ERROR!
```

✅ **Correct order:**
```go
models.LogErrorWithContext(r.Context(), "Failed to save", err)  // CORRECT
```

---

❌ **Missing context:**
```go
models.LogErrorWithContext("Failed to save", err)  // ERROR!
```

✅ **Include context:**
```go
models.LogErrorWithContext(r.Context(), "Failed to save", err)  // CORRECT
```

---

❌ **Including color codes:**
```go
models.LogInfo(Colors.Green + "Success!" + Colors.Reset)  // NO!
```

✅ **Plain message:**
```go
models.LogInfo("Success!")  // Colors added automatically
```

---

## Remaining Files to Update

1. **post-handlers.go** (~9 log statements)
2. **reaction-handlers.go** (~11 log statements)
3. **user-handlers.go** (~14 log statements)
4. **channel-handlers.go** (~24 log statements)
5. **home-handlers.go** (~29 log statements)

**Total:** ~87 log statements remaining

---

## Benefits of New System

✅ **Request Tracing:** All logs include request ID automatically
✅ **Color Coded:** Green (info), Orange (warn), Red (error)
✅ **Consistent Format:** [timestamp] [icon] [message]
✅ **Context Aware:** Request context flows through all logs
✅ **Less Code:** No manual color formatting needed
✅ **Better DX:** Logs are easier to scan visually

---

*Generated: 2025-12-14*
*Migration Started: 5/10 files complete*
