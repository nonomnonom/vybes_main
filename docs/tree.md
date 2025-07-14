# Project Directory Structure

This document provides an overview of the project's directory structure and the purpose of each main folder.

```
.
├── cmd/                  # Entry points for the application binaries.
│   └── api/              # Main application for the Vybes API.
│       └── main.go       # Initializes and starts all services.
├── docs/                 # Project documentation.
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
│   ├── media/            # Media processing utilities (e.g., FFmpeg wrapper).
│   ├── storage/          # Object storage client wrapper (MinIO).
│   └── utils/            # General utility functions (e.g., crypto, password hashing).
├── .env                  # Environment variables (not committed to git).
├── .gitignore            # Files and directories to ignore in git.
├── go.mod                # Go module definitions and dependencies.
└── go.sum                # Checksums for Go module dependencies.
```

### Directory Breakdown

*   **/cmd**: Contains the `main` packages for the executables in the project. The `api` subdirectory holds the primary entry point for our web server.
*   **/docs**: All project-related documentation, including architecture diagrams, API specifications, and setup guides.
*   **/internal**: This is where the core application logic resides. According to Go conventions, code in `internal` is not importable by other projects, ensuring our core logic remains private.
    *   **config**: Handles loading configuration from environment variables (`.env` file).
    *   **domain**: Defines the core data structures (structs) that represent the business entities, like `User`, `Post`, `Notification`, etc.
    *   **handler**: Contains the HTTP handlers. These are responsible for parsing incoming requests, calling the appropriate services, and formatting the responses. We use the Gin framework here.
    *   **middleware**: Provides reusable HTTP middleware functions that can be chained in request processing, such as for authentication or rate limiting.
    *   **repository**: Implements the data persistence layer. It defines interfaces for data access and provides MongoDB-specific implementations, abstracting the database from the business logic.
    *   **service**: The heart of the application. This layer contains all the business logic, orchestrating calls to repositories and other services to fulfill use cases (e.g., registering a user, creating a post, generating a feed).
*   **/pkg**: Contains code that is safe to be imported and used by external applications. This is for reusable components.
    *   **cache**: A wrapper for our Redis client.
    *   **evm**: Utilities for handling Ethereum-related operations like transaction and message signing.
    *   **media**: Wrappers for media processing tools like FFmpeg.
    *   **storage**: A client for interacting with our S3-compatible object storage (MinIO).
    *   **utils**: A collection of helper functions for tasks like password hashing, OTP generation, and encryption.