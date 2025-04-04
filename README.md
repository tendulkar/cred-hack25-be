# Scalable Go Backend

A modular, production-ready backend service built with Go, featuring central logging, user authentication, and PostgreSQL integration.

## Features

- **Modular Architecture**: Clean architecture principles with clear separation of concerns
- **Central Logging System**: All modules import and use the central logging package
- **User Authentication**: Complete JWT-based authentication system optimized for SaaS applications
- **Database Integration**: PostgreSQL with prepared statements for optimal performance
- **API Documentation**: Comprehensive documentation of all available endpoints
- **Error Handling**: Consistent error handling across all modules
- **Configuration Management**: Environment-based configuration using environment variables
- **Security Best Practices**: Password hashing, JWT security, and SQL injection protection

## Project Structure

```
backend/
│
├── cmd/                      # Application entry points
│   └── api/                  # Main API server
│
├── internal/                 # Private application code
│   ├── config/               # Configuration management
│   ├── handlers/             # HTTP request handlers
│   ├── middleware/           # HTTP middleware
│   ├── models/               # Data models
│   ├── repository/           # Database access layer
│   └── service/              # Business logic
│
├── pkg/                      # Public libraries
│   ├── auth/                 # Authentication utilities
│   ├── database/             # Database connections
│   └── logger/               # Centralized logging
│
└── docs/                     # Documentation
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 13+

### Configuration

Create a `.env` file in the root directory with the following variables:

```
# Application
APP_ENV=development
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hack25
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_TOKEN_TTL=15     # minutes
JWT_REFRESH_TOKEN_TTL=10080 # minutes (7 days)

# Logging
LOG_LEVEL=info
LOG_FILE=logs/app.log       # Optional, logs to stdout if empty
```

### Running the Application

Build and run the application:

```bash
go build -o app ./cmd/api
./app
```

Or simply run:

```bash
go run ./cmd/api/main.go
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Login and get access token
- `POST /api/v1/auth/refresh` - Refresh access token

### User Management

- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile

### Admin

- `GET /api/v1/admin/users` - List all users (admin only)

## Architectural Design

### Central Logging

The central logging module (`pkg/logger`) provides:

- Structured JSON logging
- Multiple log levels (debug, info, warn, error, fatal)
- File and console output
- Caller information (file and line)
- Context-enriched logging with fields

### Database Layer

The database layer uses standard SQL with prepared statements for:

- Improved security with protection against SQL injection
- Better performance through statement caching
- Explicit control over transactions
- Simplified testing

### Authentication System

The JWT-based authentication system provides:

- Secure token generation and validation
- Access and refresh token mechanism
- Role-based authorization
- Token expiration and renewal

## Development

### Adding a New Model

1. Define the model in `internal/models/`
2. Add the schema to `pkg/database/postgres.go`
3. Create a repository in `internal/repository/`
4. Create a service in `internal/service/`
5. Create handlers in `internal/handlers/`
6. Add routes in `cmd/api/main.go`

### Testing

Run tests:

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.
