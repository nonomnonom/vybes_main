# Project Directory Structure

This document provides an overview of the project's directory structure and the purpose of each main file and folder.

```
.
├── cmd/                  # Entry points for the application binaries.
│   └── api/              # Main application for the Vybes API.
│       └── main.go       # Initializes and starts all services.
├── docs/                 # Project documentation.
│   ├── api.md            # Detailed API endpoint documentation.
│   ├── architecture.md   # Architectural diagrams and explanations.
│   └── tree.md           # This file, showing the directory structure.
├── internal/             # Private application code, not for export.
│   ├── config/           # Configuration loading and management (from .env).
│   ├── domain/           # Core data structures and models (e.g., User, Post).
│   ├── handler/          # HTTP handlers (Gin) that process requests.
│   ├── middleware/       # HTTP middleware (e.g., Auth, Rate Limiting).
│   ├── repository/       # Data access layer (interfaces and MongoDB implementations).
│   └── service/          # Business logic layer.
├── pkg/                  # Public library code, can be used by other projects.
│   ├── cache/            # Redis cache client wrapper.
│   ├── evm/              # Ethereum Virtual Machine utilities (e.g., signing).
│   ├── storage/          # Object storage client wrapper (MinIO).
│   └── utils/            # General utility functions (e.g., crypto, password hashing).
├── .env                  # Environment variables (not committed to git).
├── .gitignore            # Files and directories to ignore in git.
├── Dockerfile            # Instructions to build the application's Docker image.
├── docker-compose.yml    # Defines and runs the multi-service local environment.
├── go.mod                # Go module definitions and dependencies.
├── go.sum                # Checksums for Go module dependencies.
└── Makefile              # Automates common development tasks (build, run, test, docker).
```

### File & Directory Breakdown

*   **/cmd**: Contains the `main` packages for the executables in the project.
*   **/docs**: All project-related documentation.
*   **/internal**: Core application logic, not importable by other projects.
    *   **config**: Handles loading configuration from `.env` files.
    *   **domain**: Defines core data structures (structs).
    *   **handler**: HTTP handlers (Gin) that process requests and call services.
    *   **middleware**: Reusable HTTP middleware functions (e.g., authentication).
    *   **repository**: Implements the data persistence layer, abstracting the database.
    *   **service**: The heart of the application, containing all business logic.
*   **/pkg**: Reusable library code, safe to be used by external applications.
    *   **cache**: A wrapper for our Redis client.
    *   **evm**: Utilities for handling Ethereum-related operations.
    *   **storage**: A client for interacting with MinIO object storage.
    *   **utils**: Helper functions for tasks like password hashing and encryption.
*   **Dockerfile**: Instructions for building a production-ready, multi-stage Docker image for the Go application.
*   **docker-compose.yml**: A configuration file for Docker Compose to easily set up and run the entire local development environment, including the API, database, cache, and other services.
*   **Makefile**: A command center with shortcuts for common development tasks like `make build`, `make test`, and `make docker-up`.