# Vybes Backend Architecture

This document outlines the architecture of the Vybes backend application. We follow the principles of Clean Architecture to create a system that is scalable, maintainable, and testable.

## Core Principles

*   **Separation of Concerns**: Each layer has a distinct responsibility. Handlers manage HTTP requests, Services contain business logic, and Repositories handle data access.
*   **Dependency Rule**: Dependencies flow inwards. The business logic (Services) does not depend on the presentation layer (Handlers) or the database implementation (Repositories).
*   **Decoupling**: Components are decoupled through interfaces and asynchronous communication (e.g., NATS for notifications), making the system more resilient and flexible.

## Architecture Diagram

The following diagram illustrates the overall architecture and the flow of key user interactions. It now includes the media processing flow.

```mermaid
graph TD
    subgraph Client
        UserApp[User Application]
    end

    subgraph API Gateway [Gin Router]
        Router
        subgraph Middleware
            Auth[Auth Middleware]
            RateLimit[Rate Limiter]
        end
    end

    subgraph Handlers
        UserHandler
        ContentHandler
        ReactionHandler
        FeedHandler
    end

    subgraph Services
        UserService
        ContentService
        ReactionService
        FeedService
        NotificationPublisher
        WalletService
        EmailService
    end
    
    subgraph Background Workers
        NATS_Worker[NATS Worker]
        Cron_Worker[Cron Service]
    end

    subgraph Service Logic
        NotificationService[Notification Service (DB Writer)]
    end
    
    subgraph Media Processing
        MediaProcessor[Media Processor (FFmpeg)]
    end

    subgraph Repositories
        UserRepo[UserRepository]
        ContentRepo[ContentRepository]
        FollowRepo[FollowRepository]
        ReactionRepo[ReactionRepository]
        NotificationRepo[NotificationRepository]
        StoryRepo[StoryRepository]
    end

    subgraph External Systems & Datastores
        MongoDB[(MongoDB)]
        MinIO[(MinIO Storage)]
        Redis[(Redis Cache)]
        NATS[("NATS Message Queue")]
        ResendAPI[Resend API]
    end

    %% --- Flows ---

    %% Create Post Flow with Media Processing
    UserApp -- 1. POST /posts --> Router
    Router -- 2. --> Auth
    Auth -- 3. --> ContentHandler
    ContentHandler -- 4. --> ContentService
    ContentService -- 5a. Generate Thumbnail --> MediaProcessor
    ContentService -- 5b. Upload Media --> MinIO
    ContentService -- 5c. Create Post Doc --> ContentRepo
    ContentRepo -- 6. --> MongoDB
    
    %% User Registration Flow
    UserApp -- "POST /register" --> Router
    Router --> UserHandler --> UserService
    UserService --> UserRepo & WalletService & EmailService
    UserRepo --> MongoDB
    WalletService --> UserRepo
    EmailService --> ResendAPI

    %% Like Post Flow (Sync + Async)
    UserApp -- "POST /posts/:id/like" --> Router
    Router --> Auth --> ReactionHandler --> ReactionService
    ReactionService -- Sync --> ReactionRepo --> MongoDB
    ReactionService -- Async --> NotificationPublisher --> NATS

    %% Notification Worker Flow
    NATS --> NATS_Worker --> NotificationService --> NotificationRepo --> MongoDB

    %% Feed Request Flow
    UserApp -- "GET /feeds/friends" --> Router
    Router --> Auth --> FeedHandler --> FeedService
    FeedService --> FollowRepo & ContentRepo
    FollowRepo --> MongoDB
    ContentRepo --> MongoDB

    %% Story Cleanup Cron Job
    Cron_Worker --> StoryRepo --> MongoDB
    Cron_Worker --> MinIO
    
    %% User Profile Flow with Cache
    UserApp -- "GET /users/:username" --> Router
    Router --> Auth --> UserHandler --> UserService
    UserService -- Check --> Redis
    UserService -- Cache Miss --> UserRepo --> MongoDB
    UserService -- Set --> Redis
```

### Key Components

*   **API Gateway (Gin)**: The entry point for all HTTP requests. It uses middleware for authentication and rate limiting before routing requests to the appropriate handlers.
*   **Handlers**: Responsible for parsing HTTP requests, validating input, and calling the appropriate service methods. They are the bridge between the web framework and the core business logic.
*   **Services**: Contain the core business logic of the application. They orchestrate data from repositories and other services to perform complex operations.
*   **Repositories**: An abstraction layer for data persistence. They provide a clean API for services to interact with the database without knowing the underlying implementation details (MongoDB).
*   **Media Processor**: A component responsible for handling media-related tasks. Currently, it uses FFmpeg to generate video thumbnails upon upload.
*   **Background Workers**:
    *   **NATS Worker**: Subscribes to the NATS message queue to process asynchronous tasks like creating notifications. This improves API response times by offloading work from the main request-response cycle.
    *   **Cron Service**: Runs scheduled tasks, such as cleaning up expired story media files from MinIO storage.
*   **External Systems**:
    *   **MongoDB**: The primary NoSQL database for storing all application data.
    *   **MinIO**: S3-compatible object storage for user-generated media (posts, stories, thumbnails).
    *   **Redis**: In-memory cache used to store frequently accessed data (like user profiles) to reduce database load and improve read performance.
    *   **NATS**: A lightweight, high-performance message queue used for asynchronous communication between services.
    *   **Resend**: An external API service for sending transactional emails (e.g., welcome emails, password resets).