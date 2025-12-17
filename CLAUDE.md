# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Codex** is a web forum application built with Go that allows users to communicate, create posts, react to content, and organize discussions in channels. The application features user authentication, post/comment systems, reactions (likes/dislikes), channel management, moderation, and image uploads.

## Build & Run Commands

### Local Development (Native)
```bash
# Build the application
make build

# Run the server (requires prior build)
make run

# Build and run in one step
make build && make run

# Run database migrations
bin/codex migrate

# Seed the database with initial data
bin/codex seed
```

### Docker Development
```bash
# Interactive menu for common tasks
make menu

# Configure Docker settings (image name, container name, port)
make configure

# Reset Docker configuration and choose db (dev/prod)
make reset-config

# Build Docker image
make build-image

# Run Docker container
make run-container
```

### Script Management
```bash
# Make scripts executable with correct permissions
make install-scripts

# Backup scripts before updates
make backup-scripts

# Verify script checksums to detect changes
make verify-scripts
```

## Architecture

### Directory Structure

- **`cmd/server/`** - Application entry point
  - `main.go` - Server initialization, CLI commands (migrate/seed), graceful shutdown
  - `migrations.go` - Database migration runner
  - `seed.go` - Database seeding logic

- **`internal/app/`** - Application initialization and configuration
  - Central `App` struct that holds all database models and configuration
  - Loads `.env` file for DB_ENV and DB_PATH settings
  - Initializes SQLite connection and all model instances

- **`internal/db/`** - Database connection and initialization

- **`internal/dao/`** - Generic Data Access Object with type-safe CRUD operations
  - Uses Go generics `DAO[T models.DBModel]`
  - Provides `All()`, `GetByID()`, `Insert()`, `Update()`, `Delete()`

- **`internal/sqlite/`** - SQLite-specific implementations for each model
  - `*-sql.go` files contain model-specific queries
  - Models: Users, Posts, Comments, Reactions, Channels, Memberships, Cookies, Flags, Loyalty, Saved, Mods, Rules, Images, MutedChannels, Chats
  - `chats-sql.go` - Chat and message storage with buddy/group chat support

- **`internal/models/`** - Data models and business logic
  - `*-models.go` files define structs with `db` tags for column mapping
  - `uuidfield-models.go` - Custom UUIDField type with database driver interfaces
  - Password hashing, UUID generation, validation logic

- **`internal/dbutils/`** - Database utilities
  - `uuid.go` - UUID type for database operations

- **`internal/service/`** - Business logic layer (currently minimal)
  - User creation with password hashing
  - Follows repository pattern for testability

- **`internal/http/`**
  - `handlers/` - HTTP request handlers for auth, posts, comments, channels, users, search, reactions, moderation, chats
    - `chat-handlers.go` - Chat creation and management endpoints
  - `middleware/` - Authentication, context injection, logging, timeout protection
    - `timeout.go` - Request timeout middleware with context cancellation
  - `routes/` - Router setup and handler dependency injection
    - `registry.go` - Dependency injection for handlers (important!)
    - `routes.go` - Route definitions using Go 1.22+ enhanced servemux
  - `websocket/` - WebSocket server for real-time messaging
    - Connection management with user identification via OTP
    - Chat message broadcasting to participants

- **`internal/workers/`** - Background job processing
  - `image_worker.go` - Concurrent image processing worker pool
    - Worker pool pattern with configurable goroutines and buffered job queue
    - Graceful shutdown with context-based timeout
    - Atomic state management for race-free operations
    - Database integration for image metadata persistence
  - `image_worker_test.go` - Unit tests (9 tests covering pool lifecycle, concurrency, edge cases)
  - `image_worker_integration_test.go` - Integration tests with real image files
  - `image_worker_database_test.go` - Database integration tests with in-memory SQLite

- **`internal/view/`** - HTML template rendering
  - `render.go` - Template initialization and helper functions
  - Custom template functions: `compareAsInts`, `increment`, `dict`, `reactionStatus`, etc.
  - Templates located in `assets/templates/` (`.html` and `.tmpl` files)

- **`migrations/`** - SQL schema files
  - `001_schema.sql` - Main database schema
  - `002_triggers.sql` - Database triggers
  - `003_indexes.sql` - Performance indexes
  - `004_chats.sql` - Chat/messaging schema
  - `005_add_image_path.sql` - Adds Path column to Images table for worker pool integration

- **`assets/`** - Static files (CSS, icons, fonts, templates)
  - `css/` - Modular CSS architecture
    - `main.css` - Import hub for 27+ CSS modules
    - `colors-oklch.css` - OKLCH color definitions with Catppuccin Mocha palette
    - `variables.css` - CSS custom properties (spacing, shadows, z-index)
    - Feature modules: `typography.css`, `layout.css`, `buttons.css`, `forms.css`, etc.

- **`scripts/`** - Shell scripts for configuration and deployment
  - `menu.sh` - Interactive menu for Docker operations
  - `configure.sh` - Docker configuration
  - `reset-config.sh` - Reset Docker settings

### Key Architectural Patterns

**UUID System**
Codex uses a custom UUIDField type for user identification throughout the application:
- **Type Safety**: `UUIDField` wraps `github.com/google/uuid` with compile-time type safety
- **Database Integration**: Implements `driver.Valuer` and `sql.Scanner` for automatic SQLite conversion
- **Storage**: UUIDs stored as 16-byte BLOBs in SQLite for efficiency (not strings)
- **JSON Support**: Implements `MarshalJSON`/`UnmarshalJSON` for API responses
- **No Manual Conversion**: The `Value()` method automatically converts to `[]byte` for database operations
- **Nullable Support**: `NullableUUIDField` provides SQL NULL handling with `Valid` flag

**Location**: `internal/models/uuidfield-models.go`

**Example Usage**:
```go
// Creating a new UUID
userID := models.NewUUIDField()

// Storing in database (automatic conversion to []byte)
db.Exec("INSERT INTO Users (ID) VALUES (?)", userID)

// The UUIDField.Value() method handles conversion automatically

// Nullable UUID for optional foreign keys
groupID := models.NullableUUIDField{Valid: false} // SQL NULL
buddyID := models.NullableUUIDField{UUID: userID, Valid: true} // Has value
```

**Concurrent Image Processing Worker Pool**
Production-ready background job processor for image uploads:
- **Worker Pool Pattern**: Fixed number of goroutines process jobs from buffered channel
- **Non-Blocking Submission**: `Submit()` returns error immediately if queue is full
- **Graceful Shutdown**: Context-based timeout with `sync.WaitGroup` for cleanup
- **Atomic State**: Uses `atomic.Bool` for race-free shutdown tracking
- **Database Integration**: Stores image metadata (path, authorID, postID) after successful processing
- **Error Handling**: Comprehensive validation with colored logging (teal=processing, red=error, green=success, blue=database)

**Location**: `internal/workers/image_worker.go`

**Example Usage**:
```go
// Initialize pool (5 workers, 100 job queue, database connection)
pool := workers.NewImageWorkerPool(5, 100, db)
pool.Start()

// Submit job (non-blocking)
job := workers.ImageJob{
    ID:       "unique-job-id",
    FilePath: "/tmp/upload.jpg",
    UserID:   userID,    // UUIDField
    PostID:   postID,    // int64
}
if err := pool.Submit(job); err != nil {
    // Queue full - return 503 Service Unavailable
}

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
pool.Shutdown(ctx)
```

**Real-Time Chat System**
WebSocket-based chat with buddy and group messaging:
- **Chat Types**: Buddy chats (1-on-1) and group chats with CHECK constraints ensuring data integrity
- **User Identification**: WebSocket connections authenticated via one-time passwords (OTP)
- **Idempotent Creation**: `GetBuddyChatID()` prevents duplicate chats between users
- **Database Design**: Uses `NullableUUIDField` to enforce mutual exclusivity (buddy chats have NULL GroupID, group chats have NULL BuddyID)
- **Frontend Integration**: `chat.js` handles form submission via WebSocket, `setupStartChatHandlers()` creates new chats
- **API Endpoint**: `POST /api/chats/create` creates or retrieves existing chat, returns chat ID

**Location**: `internal/http/handlers/chat-handlers.go`, `internal/sqlite/chats-sql.go`, `assets/js/chat.js`

**Example Flow**:
```go
// Create buddy chat (groupID = NULL, buddyID = user UUID)
chatID, err := h.App.Chats.CreateChat(ctx, "buddy", "",
    models.NullableUUIDField{Valid: false},              // groupID = NULL
    models.NullableUUIDField{UUID: buddyUUID, Valid: true}) // buddyID = value

// Attach users to chat
h.App.Chats.AttachUserToChat(ctx, chatID, currentUserID)
h.App.Chats.AttachUserToChat(ctx, chatID, buddyID)
```

**Dependency Injection via Handler Registry**
The application uses manual dependency injection in `internal/http/routes/registry.go`. Handlers are created in a specific order to resolve dependencies:
1. Create flat handlers (Session, Reaction, Auth)
2. Create handlers with single-level deps (Comment, Channel)
3. Create complex handlers (User, Post, Home)
4. Wire everything into RouteHandler struct

**Database Models**
All models implement the `DBModel` interface with a `TableName()` method. Struct fields use `db` tags for SQLite column mapping. The generic DAO uses reflection to map between structs and database rows.

**Template Rendering**
Templates are parsed once at startup. The `TempHelper` struct provides access to the App instance for template functions like `reactionStatus`. Templates are composed of partials (`.tmpl` files) and assembled dynamically.

**Authentication Flow**
Session tokens stored in cookies → `auth` middleware extracts user → context injection via `WithUser` middleware → handlers access user from `r.Context()`

**Routes (Go 1.22+ Enhanced Servemux)**
Uses method-based routing: `POST /register`, `GET /post/{postId}`, etc.
Middleware wrapping: `mw.WithUser(http.HandlerFunc(r.Home.GetHome), r.App)`

**CSS Architecture**
Modular CSS system for maintainability and performance:
- **Import-Based**: `main.css` imports 27+ specialized modules
- **Logical Organization**: Separated by concern (colors, typography, layout, buttons, forms, popovers, etc.)
- **OKLCH Color Space**: Modern color definitions with Catppuccin Mocha palette
- **CSS Variables**: Centralized theming for spacing, shadows, z-index, transitions
- **Benefits**: Easier maintenance, reduced merge conflicts, better caching, clear separation of concerns

**Concurrency Patterns**
- **Context Propagation**: Database queries accept `context.Context` for cancellation
- **Timeout Middleware**: HTTP requests protected with configurable timeouts
- **Graceful Shutdown**: Server shutdown waits for in-flight requests with timeout
- **Worker Pools**: Background jobs processed concurrently without blocking HTTP handlers

## Database

- **Type:** SQLite3
- **Driver:** `github.com/mattn/go-sqlite3`
- **Configuration:** `.env` file (DB_ENV, DB_PATH)
- **Schema Location:** `migrations/001_schema.sql`
- **Migrations:** Run with `bin/codex migrate` CLI command
- **Seeding:** Run with `bin/codex seed` CLI command

## Environment Configuration

Create a `.env` file with:
```
DB_ENV=dev
DB_PATH=./identifier.sqlite
```

Use `make configure` or `make reset-config` to set up Docker environment variables.

## Important Notes

- The server runs on port `8888` by default
- A pprof server runs on `localhost:6060` for profiling
- Graceful shutdown on SIGTERM/SIGINT with 10-second timeout
- Image uploads stored in `db/userdata/images/{channel-images,user-images,post-images}/`
- Custom color scheme using Catppuccin Mocha palette (see `internal/colors/`)
- Script backups automatically timestamped in `scripts/backups/`
