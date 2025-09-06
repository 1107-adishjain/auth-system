# Production-Ready Authentication System in Golang

This project is a secure, production-grade authentication backend built with Golang and containerized using Docker. It implements modern security best practices, including JWT-based authentication with rotating refresh tokens, secure cookie handling, and a robust, scalable architecture using PostgreSQL and Redis.

The system is designed to be a reliable and extensible foundation for any modern web application requiring secure user management.

## Core Features

- **User Registration:** Securely register new users with email normalization and `bcrypt` password hashing.
- **User Login:** Authenticate users and issue short-lived access tokens and long-lived, rotating refresh tokens.
- **Token Refresh:** Seamlessly refresh expired access tokens using a valid refresh token, which is then rotated for enhanced security (one-time use).
- **Secure Logout:** Invalidate user sessions by revoking refresh tokens from the Redis store.
- **Protected Routes:** Includes middleware to easily protect API endpoints, allowing access only to authenticated users.
- **Containerized Environment:** Fully containerized with Docker and Docker Compose for consistent, one-command setup and deployment.

## Tech Stack & Architecture

This project leverages a modern, scalable tech stack suitable for high-performance applications:

- **Language:** **Go (Golang)**
- **Framework:** **Gin** (A high-performance HTTP web framework)
- **Database:** **PostgreSQL** (A powerful, open-source object-relational database)
- **ORM:** **GORM** (A developer-friendly ORM for Go)
- **Cache / Session Store:** **Redis** (An in-memory data store used for managing refresh tokens)
- **Authentication:** **JWT** (JSON Web Tokens) for access and refresh tokens.
- **Migrations:** **golang-migrate** for managing database schema changes.
- **Containerization:** **Docker & Docker Compose** for building and running the application stack.
- **Configuration:** **godotenv** for local environment variable management, adhering to the 12-factor app methodology.

## Security Best Practices Implemented

Security is the primary focus of this authentication system. The following best practices have been implemented:

- **Password Hashing:** User passwords are never stored in plain text. They are securely hashed using the industry-standard **`bcrypt`** algorithm.
- **Rotating Refresh Tokens:** To mitigate token theft, refresh tokens are single-use. When a refresh token is used, it is immediately invalidated, and a new one is issued.
- **Secure `HttpOnly` Cookies:** JWTs are stored in `HttpOnly`, `Secure`, and `SameSite=Strict` cookies, preventing access from client-side JavaScript (mitigating XSS attacks) and ensuring they are only sent over HTTPS.
- **Short-Lived Access Tokens:** Access tokens have a short TTL (15 minutes), minimizing the window of opportunity for misuse if one is compromised.
- **Centralized Token Revocation:** Refresh tokens are stored in Redis, allowing for instant, server-side revocation upon logout or if a security event is detected.
- **Parameter-ized Queries:** GORM abstracts database interactions, preventing SQL injection vulnerabilities by using parameterized queries.
- **Separation of Concerns:** A clean, multi-layered architecture (handler, service, repository) ensures that code is organized, maintainable, and secure.
- **Environment-Based Secret Management:** All sensitive information (database passwords, JWT secrets) is managed via environment variables and is never hardcoded into the source code.

## Project Structure

The project follows a clean, industry-standard layout to promote separation of concerns and maintainability.

```
/auth-system
├── cmd/api/            # Main application entrypoint
├── config/             # Environment variable loading
├── db/migration/       # Database migration files
├── internal/
│   ├── auth/           # Authentication logic (handlers, services, etc.)
│   ├── middleware/     # Gin middleware (e.g., for auth)
│   ├── models/         # GORM database models
│   └── server/         # Gin server setup and configuration
├── .env                # Local development environment variables (NOT in Git)
├── docker-compose.yml  # Docker service orchestration
└── Dockerfile          # Multi-stage Docker build for the Go app
```

## Getting Started

### Prerequisites
- Docker
- Docker Compose

### Running the Application

This project is fully containerized, allowing for a simple, one-command setup.

1.  **Clone the Repository**
    ```
    git clone <your-repository-url>
    cd auth-system
    ```

2.  **Create Your Environment File**
    Create a `.env` file in the root of the project by copying the provided `.env.example` file (or create one from scratch). This file contains all the necessary environment variables, including your database credentials and `JWT_SECRET`.

    ```
    # Create the file
    touch .env
    ```

    Paste the following into your `.env` file and **change the `JWT_SECRET` to your own strong, random string**:
    ```
    DB_HOST=postgres
    DB_PORT=5432
    DB_USER=user
    DB_PASSWORD=password
    DB_NAME=auth_db
    REDIS_ADDR=redis:6379
    SERVER_PORT=8080
    JWT_SECRET="your-own-very-long-and-secure-random-string-here"
    ACCESS_TOKEN_TTL=15
    REFRESH_TOKEN_TTL=10080
    ```

3.  **Build and Run with Docker Compose**
    Run the following command from the project root. This will build the Go application, start the PostgreSQL and Redis containers, run the database migrations, and start the API server.

    ```
    docker-compose up --build
    ```

The API will be available at `http://localhost:8080`.

## API Endpoints Reference

All endpoints are prefixed with `/api/v1`.

| Feature      | Method | Endpoint              | Request Body                                        | Response                                               |
|--------------|--------|-----------------------|-----------------------------------------------------|--------------------------------------------------------|
| **Register**   | `POST` | `/auth/register`        | `{ "email": "user@example.com", "password": "..." }` | `{ "id": 1, "email": "...", "created_at": "..." }`       |
| **Login**      | `POST` | `/auth/login`           | `{ "email": "user@example.com", "password": "..." }` | `{ "message": "Logged in successfully" }` (sets cookies) |
| **Refresh**    | `POST` | `/auth/refresh`         | (No body, requires `refresh_token` cookie)          | `{ "message": "Tokens refreshed successfully" }`       |
| **Logout**     | `POST` | `/auth/logout`          | (No body, requires `refresh_token` cookie)          | `{ "message": "Logged out successfully" }`             |
| **Health Check** | `GET`  | `/health`               | (No body)                                           | `{ "status": "ok" }`                                   |
| **Protected**  | `GET`  | `/protected`            | (No body, requires `access_token` cookie)           | `{ "message": "This is a protected route", "user_id": "..." }` |
```
