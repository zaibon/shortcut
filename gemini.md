# Shortcut - Project Context

## Overview
This project is a URL shortener application written in Go. It features user authentication, URL management, click analytics, and subscription plans (Stripe integration).

## Technology Stack

### Backend
- **Language:** Go (1.23+)
- **Web Framework:** [Chi](https://github.com/go-chi/chi) (v5)
- **Database:** PostgreSQL
- **Data Access:** [sqlc](https://sqlc.dev/) (Generates Go from SQL)
- **Migrations:** [goose](https://github.com/pressly/goose)

### Frontend
- **Templating:** [templ](https://templ.guide/) (Type-safe Go-to-HTML)
- **Styling:** Tailwind CSS
- **Interactivity:** [htmx](https://htmx.org/) (Server-driven UI updates)
- **Client Logic:** [Alpine.js](https://alpinejs.dev/) (Lightweight state)

### Infrastructure & Tooling
- **Task Runner:** `just` (See `justfile`)
- **Live Reload:** `air`
- **Deployment:** Fly.io (`fly.toml`)
- **Container:** Docker

## Directory Structure & Architecture

- **`cmd/`**: Application entry points.
  - `server.go`: Main HTTP server initialization and routing.
  - `migration.go`: Migration CLI command.
- **`db/`**: Database related code.
  - `queries/`: **Source of Truth** for database interactions. Raw SQL files.
  - `datastore/`: **Auto-generated** Go code by `sqlc`. **DO NOT EDIT MANUALLY**.
  - `migrations/`: SQL migration files managed by `goose`.
- **`domain/`**: Core business entities and interface definitions. Used to decouple layers.
- **`handlers/`**: HTTP Controllers. Responsible for parsing requests, validating input, calling `services`, and rendering `templ` views.
- **`services/`**: Business Logic Layer. Implements complex logic, interacts with `datastore`, and manages external APIs (e.g., Stripe).
- **`templates/`**: UI View Layer. Contains `*.templ` files.
  - `components/`: Reusable UI parts.
  - `layout.templ`: Master page layout.

## Development Workflows

### 1. Database Changes
1.  Create a new migration: `just db-create-migration <name> sql`
2.  Edit the new file in `db/migrations/`.
3.  Add or update queries in `db/queries/<model>.sql`.
4.  Run code generation: `just generate` (updates `db/datastore`).
5.  Apply migrations: `just db-migrate up`.

### 2. Frontend Changes
1.  Edit `templates/**/*.templ` files.
2.  Run `just generate` to recompile templates to Go code.
3.  (Optional) During development, `just dev` runs `templ generate --watch`.

### 3. Backend Logic
1.  Define interfaces/structs in `domain/` if necessary.
2.  Implement logic in `services/`.
3.  Wire up the service in `cmd/server.go` (dependency injection).
4.  Create/Update `handlers/` to use the service.

## Key Commands (`justfile`)

- **`just dev`**: Start the dev environment (Air + Templ watch).
- **`just generate`**: Run all code generators (`sqlc`, `templ`, `go generate`).
- **`just test`**: Run all tests.
- **`just db-migrate up`**: Apply DB migrations.
- **`just db-create-migration <name> <type>`**: Create a new migration file.
- **`just build`**: Build the binary.

## Coding Conventions
- **Strict Typing:** Use `domain` types over primitives where possible.
- **No ORM:** Do not use GORM or similar. Use raw SQL in `db/queries` and let `sqlc` handle the Go code.
- **Server-Side Rendering:** Prefer `templ` + `htmx` over sending JSON to a JS frontend, unless building a specific public API.
