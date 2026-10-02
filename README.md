# Web Chat

A fast, lightweight, and full-stack **Web Chat application** built with **Go (Golang)**. This project features server-side rendering for its frontend, uses **SQLC** for type-safe database queries, and is fully containerized.

## 🚀 Features
* **Real-time Messaging**: Powered by full-duplex **Gorilla WebSocket** connections for instant chat updates.
* **Modern Frontend (HTMX)**: Dynamic UI interactions using **HTMX** combined with Go's `html/template` engine, eliminating the need for heavy JavaScript frameworks.
* **User Authentication**: Built-in secure user registration, login, and session logout functionality.
* **Protected Routes**: Secure group routing via `authMiddleware.Middleware` to safeguard private channels.
* **Type-Safe DB Queries**: Integrated with SQLC to compile raw SQL from `sql/queries/` into clean Go source code.
* **Database Migrations**: Automatic database schema management for persistent data.
* **Dockerized Setup**: Ready for local development and deployment via Docker Compose.


## 📁 Project Structure
* `cmd/app/` — Main entry point that boots up the web chat server.
* `internal/` — Core business logic, HTTP route handlers, and SQLC-generated repository files.
* `migrations/` — SQL database migration scripts.
* `sql/queries/` — Raw SQL query templates used by SQLC.
* `templates/` — HTML template files for rendering the web interface.
* `sqlc.yaml` — Configuration file for the SQLC code generator.

## 🛠️ Prerequisites
Before running the project, make sure you have the following installed:
* [Go](https://go.dev) (1.22+ recommended)
* [Docker](https://docker.com) & [Docker Compose](https://docker.com)
* [SQLC](https://sqlc.dev) (Optional, only for regenerating DB code)

## ⚡ Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com
cd web-chat
```

### 2. Run with Docker Compose (Recommended)
To start the chat application along with its pre-configured database environment, execute:
```bash
docker-compose up --build
```

### 3. Run Locally
To run the project directly on your system:
```bash
# Install dependencies
go mod download

# Run the application
go run cmd/app/main.go
```

## 🛠️ Database Development (SQLC)
If you add or update SQL queries inside `sql/queries/`, regenerate the type-safe Go database layer using:
```bash
sqlc generate
```

## 🔌 API Endpoints

The project uses `go-chi` for routing. Standard web actions are public, while real-time connection lines are secured.

### 🌐 Web & Authentication Routes

| Method | Endpoint | Handler | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | `chatHandler.RenderIndex` | Renders the main chat interface (HTMX template) |
| `POST` | `/register` | `authHandler.Register` | Registers a new user account |
| `POST` | `/login` | `authHandler.Login` | Logs the user in and establishes a session |
| `GET` | `/logout` | `authHandler.Logout` | Logs out the user and clears the session |
| `POST` | `/chat/clear`| `chatHandler.ClearChat` | Clears the chat message history |

### 🔒 Protected Routes (Auth Required)

| Method | Endpoint | Handler | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/ws` | `wsHandler.ServeWS` | Upgrades the HTTP connection to a secure **Gorilla WebSocket** |

## 📝 License
This project is open-source and available under the MIT License.
