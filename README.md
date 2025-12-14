<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->

<a id="readme-top"></a>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->

<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->

[//]: # ([![Forks][forks-shield]][forks-url])
[![Contributors][contributors-shield]][contributors-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/gary.norman/forum">
    <img src="assets/icons/codex-logo-green-trimmed.svg" alt="Logo" width="350">
  </a>
  <p align="center">
    A forum for finding all the help you need with code.
    <br />
    <a href="https://github.com/gary-norman/forum/wiki"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://codex.loreworld.live">View Demo</a>
    ·
    <a href="https://github.com/gary-norman/forum/issues/new?labels=type%3A+bug">Report Bug</a>
    ·
    <a href="https://github.com/gary-norman/forum/issues/new?labels=type%3A+feature">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
        <li><a href="#architecture">Architecture</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->

[![Codex screenshot][product-screenshot]](https://codex.loreworld.live)

## About The Project

**Codex** is a feature-rich web forum application built with Go that provides a modern, community-driven platform for discussions and knowledge sharing. Designed with developers in mind, Codex offers a comprehensive set of features for creating, organizing, and moderating online communities.

### Key Features

#### User Management
- **Authentication & Authorization** - Secure user registration and login with session-based authentication
- **User Profiles** - Customizable profiles with avatars and user information
- **Role-Based Access** - Moderator and admin roles with granular permissions

#### Content & Communication
- **Posts & Comments** - Create threaded discussions with rich content
- **Channels** - Organize discussions into topic-based channels
- **Channel Membership** - Join/leave channels to customize your feed
- **Reactions** - Like and dislike posts and comments to surface quality content
- **Bookmarks** - Save posts for later reference

#### Community Features
- **Channel Rules** - Set community guidelines per channel
- **Mute Channels** - Hide channels you're not interested in
- **Content Filtering** - Filter posts by channel, reactions, and user-created content

#### Media & Design
- **Image Uploads** - Attach images to posts, channels, and user profiles
- **Concurrent Image Processing** - Worker pool for background image processing
- **Responsive Design** - 🚧 Work in progress for mobile optimization
- **Modular CSS Architecture** - Organized stylesheets for maintainability

#### Engagement & Rewards
- **Loyalty System** - Track user engagement and participation
- **Search Functionality** - Find posts, channels, and users quickly

### Built With

* **Backend:** Go 1.22+
* **Database:** SQLite3
* **Templating:** Go HTML templates
* **Containerization:** Docker
* **Build System:** Make

### Architecture

#### UUID System
Codex uses a robust UUID implementation for user identification throughout the application:

- **UUIDField Type** - Custom type wrapping `github.com/google/uuid` with database driver interfaces
- **Automatic Conversion** - Implements `driver.Valuer` and `sql.Scanner` for seamless SQLite integration
- **BLOB Storage** - UUIDs stored as 16-byte BLOBs in SQLite for efficiency
- **JSON Marshaling** - Automatic string conversion for API responses
- **Type Safety** - Compile-time guarantees for user ID operations

**Key Files:**
- `internal/models/uuidfield-models.go` - UUIDField type definition
- `internal/dbutils/uuid.go` - Database utilities

#### Concurrent Image Processing
Production-ready worker pool for background image processing:

- **Worker Pool Pattern** - Configurable number of goroutines with buffered job queue
- **Graceful Shutdown** - Context-based timeout with proper cleanup
- **Atomic State Management** - Race-free shutdown tracking using `sync/atomic`
- **Database Integration** - Stores image metadata after successful processing
- **Error Handling** - Comprehensive validation and colored logging

**Key Files:**
- `internal/workers/image_worker.go` - Worker pool implementation
- `internal/workers/image_worker_test.go` - Unit tests (9 tests)
- `internal/workers/image_worker_integration_test.go` - Integration tests
- `internal/workers/image_worker_database_test.go` - Database integration tests

**Features:**
- Non-blocking job submission with queue-full detection
- Image validation (JPEG, PNG, GIF)
- Directory-based organization (post-images, user-images, channel-images)
- Metadata persistence to database with path tracking

#### Database Patterns

**Generic DAO Layer** (🚧 Planned):
- Type-safe CRUD operations using Go generics: `DAO[T models.DBModel]`
- Automatic struct-to-SQL mapping via reflection
- Currently using model-specific implementations in `internal/sqlite/`

**Migration System:**
- Sequential SQL migrations in `migrations/` directory
- Automatic tracking of applied migrations
- Run with: `bin/codex migrate`

**Database Seeding:**
- Populate initial data for development
- Run with: `bin/codex seed`

**Recent Migrations:**
- `005_add_image_path.sql` - Adds Path column to Images table for worker pool integration

#### CSS Architecture

**Modular System:**
Codex uses a modular CSS architecture for maintainability and performance:

- **Import-Based Structure** - `main.css` imports 27+ specialized modules
- **Logical Organization** - Separated by concern (typography, layout, buttons, forms, etc.)
- **OKLCH Color System** - Modern color space with Catppuccin Mocha palette
- **CSS Variables** - Centralized theming in `variables.css`

**Key Modules:**
- `colors-oklch.css` - OKLCH color definitions with light/dark mode support
- `variables.css` - CSS custom properties for spacing, shadows, z-index
- `layout.css`, `typography.css`, `buttons.css`, `forms.css` - Core UI modules
- `popovers.css`, `feeds.css`, `cards.css` - Feature-specific styles

**Benefits:**
- Easy navigation and maintenance
- Reduced merge conflicts
- Better caching (specific module updates)
- Clear separation of concerns

#### Concurrency & Fault Tolerance

**Request Tracing:**
- UUID-based request IDs for distributed tracing
- Context propagation across middleware and handlers
- X-Request-ID response headers for client-side debugging
- Automatic correlation of logs across concurrent operations

**Circuit Breaker Pattern:**
- Database protection with automatic failure detection
- Configurable failure threshold (5 failures, 5s timeout)
- Half-open state for recovery testing
- Shared circuit state across concurrent goroutines
- Located in `internal/patterns/circuitbreaker.go`

**Async Logging System:**
- Worker pool for non-blocking log writes
- Buffered queue (1000 entries) for high throughput
- Graceful shutdown with queue draining
- Database persistence for request logs

**Graceful Shutdown:**
- Ordered shutdown sequence (HTTP → Logger → Database)
- Context-based timeout for cleanup operations
- Wait for in-flight requests before closing connections

**Colorized Logging:**
- Context-aware logging with request ID injection
- Color-coded output (🟢 info, 🟠 warn, 🔴 error)
- Emoji icons for visual scanning
- Automatic timestamp and formatting

**Key Files:**
- `internal/http/middleware/tracing.go` - Request ID middleware
- `internal/patterns/circuitbreaker.go` - Circuit breaker implementation
- `internal/workers/logger_pool.go` - Async logging worker pool
- `internal/models/logs-models.go` - Colorized logging functions
- `cmd/server/main.go` - Graceful shutdown implementation

<!-- Do a search and replace with your text editor for the following: `gary.norman`, `forum`, `twitter_handle`, `linkedin_username`, `email_client`, `email`, `project_title`, `project_description` -->

<p align="right">(<a href="#readme-top">back to top</a>)</p>




[//]: # (* [![Next][Next.js]][Next-url])

[//]: # (* [![React][React.js]][React-url])

[//]: # (* [![Vue][Vue.js]][Vue-url])

[//]: # (* [![Angular][Angular.io]][Angular-url])

[//]: # (* [![Svelte][Svelte.dev]][Svelte-url])

[//]: # (* [![Laravel][Laravel.com]][Laravel-url])

[//]: # (* [![Bootstrap][Bootstrap.com]][Bootstrap-url])

[//]: # (* [![JQuery][JQuery.com]][JQuery-url])

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* Go 1.22 or higher
* SQLite3

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/gary-norman/forum.git
   cd forum
   ```

### Usage

#### Option 1: Using Make (Recommended)

The project includes an interactive menu system for easy command execution:

```sh
make menu
```

This will display an interactive menu with the following options:
- **build** - Build the web server application
- **run** - Run the web server application
- **build-run** - Build and run the application in sequence
- **Docker** - Docker management submenu (configure, build image, run container)
- **Scripts** - Script management submenu (install, verify, backup scripts)

You can navigate using arrow keys or type the option number.

#### Option 2: Using Make Directly

```sh
# Build the application
make build

# Run the application
make run

# Build and run
make build-run

# Docker commands
make configure        # Configure Docker options
make build-image      # Build Docker image
make run-container    # Run Docker container

# Script management
make install-scripts  # Install/update scripts
make verify-scripts   # Verify script checksums
make backup-scripts   # Backup scripts
```

#### Option 3: Direct Terminal Commands

If you prefer not to use Make:

```sh
# Build the application
go build -o bin/codex github.com/gary-norman/forum/cmd/server

# Run the application
./bin/codex

# Build and run
go build -o bin/codex github.com/gary-norman/forum/cmd/server && ./bin/codex
```

The server will start on `http://localhost:8888` by default.

#### Database Management

```sh
# Run database migrations
bin/codex migrate

# Seed the database with initial data
bin/codex seed
```

**Note:** The application will automatically run migrations on first startup, but you can run them manually using the commands above.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->

<!-- Use this space to show useful examples of how a project can be used. Additional screenshots, code examples and demos work well in this space. You may also link to more resources. -->

_For more examples, please refer to the [Documentation](https://example.com)_

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

### Completed ✓
- [x] Core forum functionality (posts, comments, reactions)
- [x] User authentication and authorization
- [x] Channel-based organization
- [x] Image upload support
- [x] **Concurrent image processing with worker pool**
- [x] **UUID-based user identification system**
- [x] **Modular CSS architecture**
- [x] **Database migration system**
- [x] **Request tracing with context propagation**
- [x] **Circuit breaker pattern for fault tolerance**
- [x] **Async logging with worker pool**
- [x] **Graceful shutdown with cleanup**
- [x] **Colorized logging system**
- [x] Search functionality
- [x] Docker deployment
- [x] Interactive build menu

### In Progress 🚧
- [ ] Responsive design (mobile optimization)
- [ ] Error handling improvements (400/500 pages)
- [ ] Enhanced UI/UX refinements
- [ ] Image optimization (resizing, thumbnails)
- [ ] Logging migration (5/10 handlers complete - see LOGGING_MIGRATION.md)

### Planned 📋

#### Architecture Improvements
- [ ] Generic DAO layer with Go generics
- [ ] Type-safe database queries with reflection
- [ ] Complete logging migration for all handlers

#### Advanced Features
- [ ] Bookmark system
- [ ] Real-time notifications
- [ ] Advanced filtering options
- [ ] User reputation system
- [ ] Markdown support for posts

#### Security Enhancements
- [ ] Rate limiting
- [ ] Enhanced session management
- [ ] Two-factor authentication
- [ ] Content sanitization improvements

#### Moderation Tools
- [ ] Content flags and reporting system
- [ ] Channel moderation roles
- [ ] Admin dashboard
- [ ] Automated content filtering
- [ ] User reporting system
- [ ] Moderator activity logs

#### Authentication
- [ ] OAuth integration (GitHub, Google)
- [ ] Email verification
- [ ] Password recovery

See the [open issues](https://github.com/gary-norman/forum/issues) for a full list of proposed features (and known issues).

View progress by [milestone](https://github.com/gary-norman/forum/milestones).

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTRIBUTING -->

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<a href="https://github.com/github.com/gary-norman/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=github.com/gary-norman" />
</a>

Made with [contrib.rocks](https://contrib.rocks).


<!-- LICENSE -->

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- CONTACT -->

Gary Norman - [@gary](https://twitter.com/twitter_handle_gary) - email@gary.com

Kamil Ornal - [@kamil](https://twitter.com/twitter_handle_kamil) - email@kamil.com

<!-- Project Link: [https://github.com/gary.norman/forum](https://github.com/gary.norman/forum) -->

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->

<!-- 
* []()
* []()
* []()
-->

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/gary-norman/forum.svg?style=for-the-badge
[contributors-url]: https://github.com/gary-norman/forum/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/gary-norman/forum.svg?style=for-the-badge
[forks-url]: https://github.com/gary-norman/forum/network/members
[stars-shield]: https://img.shields.io/github/stars/gary-norman/forum.svg?style=for-the-badge
[stars-url]: https://github.com/gary-norman/forum/stargazers
[issues-shield]: https://img.shields.io/github/issues/gary-norman/forum.svg?style=for-the-badge
[issues-url]: https://github.com/gary-norman/forum/issues
[license-shield]: https://img.shields.io/github/license/gary-norman/forum.svg?style=for-the-badge
[license-url]: https://github.com/gary-norman/forum/blob/main/LICENSE
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/in/gary-norman/
[product-screenshot]: /assets/images/screenshot-3d.png
[Next.js]: https://img.shields.io/badge/next.js-000000?style=for-the-badge&logo=nextdotjs&logoColor=white
[Next-url]: https://nextjs.org/
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://reactjs.org/
[Vue.js]: https://img.shields.io/badge/Vue.js-35495E?style=for-the-badge&logo=vuedotjs&logoColor=4FC08D
[Vue-url]: https://vuejs.org/
[Angular.io]: https://img.shields.io/badge/Angular-DD0031?style=for-the-badge&logo=angular&logoColor=white
[Angular-url]: https://angular.io/
[Svelte.dev]: https://img.shields.io/badge/Svelte-4A4A55?style=for-the-badge&logo=svelte&logoColor=FF3E00
[Svelte-url]: https://svelte.dev/
[Laravel.com]: https://img.shields.io/badge/Laravel-FF2D20?style=for-the-badge&logo=laravel&logoColor=white
[Laravel-url]: https://laravel.com
[Bootstrap.com]: https://img.shields.io/badge/Bootstrap-563D7C?style=for-the-badge&logo=bootstrap&logoColor=white
[Bootstrap-url]: https://getbootstrap.com
[JQuery.com]: https://img.shields.io/badge/jQuery-0769AD?style=for-the-badge&logo=jquery&logoColor=white
[JQuery-url]: https://jquery.com 
