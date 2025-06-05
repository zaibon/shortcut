# Shortcut - Your Friendly URL Shortener

Shortcut is a simple yet powerful URL shortener service written in Go. It allows you to create short, manageable links from long URLs, making them easier to share and track.

## ✨ Features

* **Shorten URLs:** Create concise links for easy sharing.
* **OAuth Integration:** Supports Google and GitHub for user authentication.
* **Stripe Integration:** Manage subscriptions and payments.
* **Database Support:** Uses PostgreSQL for data storage.
* **Admin Dashboard:** Manage users, links, and view statistics.
* **Sentry Integration:** For error tracking and monitoring.
* **Easy to Deploy:** Includes a `Dockerfile` and `fly.toml` for quick deployment.

## 🚀 Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

* [Go](https://golang.org/doc/install) (version 1.23 or higher)
* [Docker](https://docs.docker.com/get-docker/) (optional, for containerized deployment)
* [Just](https://github.com/casey/just) (optional, for using the `justfile` commands)
* A PostgreSQL database instance.

### Installation & Setup

1. **Clone the repository:**

   ```sh
   git clone https://github.com/zaibon/shortcut.git
   cd shortcut
   ```

2. **Configuration:**
   The application is configured via environment variables. You can create a `.env` file in the root directory or set them directly. Refer to `cmd/server.go` for a list of available environment variables (e.g., `SHORTCUT_DB`, `SHORTCUT_DOMAIN`, `SHORTCUT_GOOGLE_OAUTH_CLIENT_ID`, etc.).

   A `justfile` is provided for common development tasks. You can enable a development environment configuration using:

   ```sh
   just enable-env dev
   ```

   This will create a symlink from `.env-dev` to `.env`. You'll need to populate `.env-dev` with your development settings.

3. **Database Migrations:**
   Before running the application, you need to apply the database migrations.

   ```sh
   # Make sure your SHORTCUT_DB environment variable is set
   just db-migrate up
   ```

4. **Build the application:**

   ```sh
   just build
   ```

   Or for development with live reloading (requires `air`):

   ```sh
   just dev
   ```

5. **Run the application:**

   ```sh
   ./bin/shortcut server
   ```

   By default, the server will run on `http://localhost:8080`.

### Development

The `justfile` contains several helpful commands for development:

* `just watch`: Generates `templ` templates and watches for changes.
* `just dev`: Runs the application in development mode with live reloading (uses `air` and `templ generate --watch`).
* `just fmt`: Formats the Go code.
* `just lint`: Lints the Go code.
* `just generate`: Generates Go code (e.g., from SQLC, templ).
* `just build`: Builds the application binary.
* `just test`: Runs the tests.
* `just coverage`: Runs tests and generates a coverage report.

## 🐳 Docker

You can build and run Shortcut using Docker:

1. **Build the Docker image:**

   ```sh
   just package
   # or
   docker build -t zaibon/shortcut:latest .
   ```

2. **Run the Docker container:**
   Make sure to pass the necessary environment variables for configuration.

   ```sh
   docker run -p 8080:8080 \
     -e SHORTCUT_DB="your_db_connection_string" \
     -e SHORTCUT_DOMAIN="your.domain.com" \
     # Add other environment variables as needed
     zaibon/shortcut:latest server
   ```

## ☁️ Deployment

This project includes a `fly.toml` file, making it easy to deploy to [Fly.io](https://fly.io/).

```sh
fly deploy
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m '''Add some AmazingFeature'''`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.

*(Create a `LICENSE` file in the root of your project, e.g., `LICENSE.md`)*

## 🙏 Acknowledgements

* [Go](https://golang.org/)
* [htmx](https://htmx.org/)
* [Chi Router](https://github.com/go-chi/chi)
* [SQLC](https://sqlc.dev/)
* [Templ](https://templ.guide/)
* [Goose](https://github.com/pressly/goose)
* [Urfave CLI](https://cli.urfave.org/)
* And all other amazing open-source libraries used in this project!
