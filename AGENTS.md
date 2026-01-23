# Agent Guidelines for github.com/zaibon/shortcut

This document provides instructions and guidelines for AI agents modifying this codebase.

## 1. Build, Lint, and Test

This project uses `just` as a command runner. Always prefer `just` commands over raw shell commands when available.

### Common Commands
- **Build:** `just build` (compiles to `bin/shortcut`)
- **Lint:** `just lint` (runs `golangci-lint` after formatting)
- **Format:** `just fmt` (runs `gci` to organize imports and format code)
- **Test (All):** `just test` (runs all tests with `go test -v ./...`)
- **Generate Code:** `just generate` (runs `templ` generation, `sqlc`, etc.)
- **Database Migrations:** `just db-migrate up` (or `down`, `status`)
- **Run Locally:** `just run` or `just dev` (for hot reload with `air`)

### Running Specific Tests
The `justfile` runs all tests. To run a single test or package, use standard `go test`:

```bash
# Run a specific test function
go test -v ./path/to/package -run TestName

# Run tests in a specific package
go test -v ./path/to/package
```

**Agent Rule:** Always run `just lint` and `just test` before declaring a task complete. If you modify templates (`.templ` files), run `just generate` first.

## 2. Project Structure & Architecture

The project follows a Domain-Driven Design (DDD) inspired layout:

- **`cmd/`**: Entry points (e.g., `server.go`).
- **`domain/`**: Core business logic interfaces and types.
- **`services/`**: Implementation of business logic (implements `domain` interfaces).
- **`handlers/`**: HTTP handlers (Controllers).
- **`middleware/`**: HTTP middleware (Auth, Sentry, Logging).
- **`db/`**: Database access (migrations, `sqlc` queries).
- **`templates/`**: `templ` UI components.
- **`static/`**: Static assets.

## 3. Code Style & Conventions

### Go Guidelines (Go 1.24+)
- **Formatting:** Code must be formatted with standard `gofmt`. Imports are sorted using `gci` with specific groups:
    1. Standard Library
    2. Third-party packages
    3. Local packages (`github.com/zaibon/shortcut/...`)
- **Error Handling:**
    - Use explicit error checking: `if err != nil { return nil, err }`.
    - Wrap errors when adding context, but follow existing patterns (mostly `%v` or `%w`).
    - Do not swallow errors.
- **Naming:**
    - **Exported:** PascalCase (e.g., `UserFromContext`).
    - **Internal:** camelCase (e.g., `contextUser`).
    - **Acronyms:** Keep consistent case (e.g., `ServeHTTP`, `ID`, `URL`, `JSON` - not `Url` or `Json`).
- **Context:**
    - Use `context.Context` as the first argument for functions performing I/O or long-running tasks.
    - Use `middleware.UserFromContext(ctx)` to retrieve the authenticated user.

### Tech Stack Specifics
- **Web Framework:** `chi` router.
- **Templates:** `templ`. **Important:** `templ` files generate Go code (`_templ.go`). Do not edit the generated files directly. Edit `.templ` files and run `just generate`.
- **Frontend:** `htmx` is used for dynamic behavior.
- **Database:** PostgreSQL with `pgx`.
    - Use `sqlc` for type-safe SQL queries. Edit queries in `db/queries/*.sql` and run `just generate` (via `go generate`).
- **Logging:** Use `github.com/zaibon/shortcut/log`.

## 4. Dependencies
- **Management:** `go.mod`.
- **Verification:** Run `go mod tidy` if adding/removing imports.

## 5. Deployment
- **Docker:** `Dockerfile` is available for containerization.

## 6. Development Workflow for Agents
1. **Explore:** Use `ls -R`, `read`, or `grep` to understand the relevant files.
2. **Plan:** Determine if changes affect DB schemas (`goose`), SQL queries (`sqlc`), or Templates (`templ`).
3. **Edit:** Apply changes.
4. **Generate:** If you touched `.sql` or `.templ` files, run `just generate`.
5. **Verify:**
    - Run `just fmt`.
    - Run `just lint`.
    - Run `just test` (or specific tests).
6. **Commit:** Ensure no generated files are missing if they are tracked (check `.gitignore` if unsure, generally `_templ.go` files are often committed in this repo style, check existing files).
