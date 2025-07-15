package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"
	"vybes/internal/config"
	"vybes/internal/domain"
	httphandler "vybes/internal/handler/http"
	"vybes/internal/repository"
	"vybes/internal/service"
	"vybes/pkg/cache"
	"vybes/pkg/storage"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// main initializes and starts the Vybes API server.
// It sets up all necessary components including database connections,
// services, handlers, and background workers.
func main() {
	// Initialize structured logging with console output for development
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load application configuration from environment variables
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Establish connection to MongoDB database
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to establish MongoDB connection")
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.DBName)
	log.Info().Str("database", db.Name()).Msg("Successfully connected to MongoDB")

	// Create database indexes for optimal query performance
	repository.SetupIndexes(context.Background(), db)

	// Initialize cloud storage client for file uploads
	storageClient, err := storage.NewClient(context.Background(), cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage client")
	}

	// Initialize Redis cache client for session and data caching
	cacheClient, err := cache.NewClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize cache client")
	}

	// Initialize NATS message broker for real-time notifications
	notificationPublisher, err := service.NewNATSNotificationPublisher(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize NATS notification publisher")
	}

	// Initialize all data access layer repositories
	userRepository := repository.NewMongoUserRepository(db)
	followRepository := repository.NewMongoFollowRepository(db)
	counterRepository := repository.NewMongoCounterRepository(db)
	storyRepository := repository.NewMongoStoryRepository(db)
	contentRepository := repository.NewMongoContentRepository(db)
	reactionRepository := repository.NewMongoReactionRepository(db)
	bookmarkRepository := repository.NewMongoBookmarkRepository(db)
	notificationRepository := repository.NewMongoNotificationRepository(db)
	sessionRepository := repository.NewSessionRepository(db)

	// Initialize all business logic services with their dependencies
	emailService := service.NewResendEmailService(cfg)
	walletService := service.NewWalletService(cfg)
	notificationService := service.NewNotificationService(notificationRepository)
	sessionService := service.NewSessionService(sessionRepository)
	// Pass pointers to the session repository and service
// Cast the pointers to interfaces to satisfy the function signature
// Cast the pointers to interfaces to satisfy the function signature
	userService := service.NewUserService(userRepository, followRepository, counterRepository, sessionRepository, walletService, emailService, sessionService, cacheClient, cfg.JWTSecret, cfg.WalletEncryptionKey)
	followService := service.NewFollowService(followRepository, userRepository, notificationPublisher)
	suggestionService := service.NewSuggestionService(userRepository, followRepository)
	storyService := service.NewStoryService(storyRepository, followRepository, storageClient, cfg)
	contentService := service.NewContentService(contentRepository, userRepository, followRepository, storageClient, notificationPublisher, cfg)
	reactionService := service.NewReactionService(reactionRepository, contentRepository, userRepository, notificationPublisher)
	feedService := service.NewFeedService(contentRepository, followRepository)
	bookmarkService := service.NewBookmarkService(bookmarkRepository, contentRepository)
	searchService := service.NewSearchService(userRepository)
	cronService := service.NewCronService(cfg, storyRepository, storageClient)

	// Start background NATS worker for processing notification events
	go startNATSWorker(cfg, notificationService)

	// Start background cron jobs for scheduled tasks (e.g., story cleanup)
	go cronService.Start()

	// Initialize all HTTP handlers with their corresponding services
	userHandler := httphandler.NewUserHandler(userService)
	followHandler := httphandler.NewFollowHandler(followService)
	suggestionHandler := httphandler.NewSuggestionHandler(suggestionService)
	storyHandler := httphandler.NewStoryHandler(storyService)
	contentHandler := httphandler.NewContentHandler(contentService)
	reactionHandler := httphandler.NewReactionHandler(reactionService)
	feedHandler := httphandler.NewFeedHandler(feedService)
	bookmarkHandler := httphandler.NewBookmarkHandler(bookmarkService)
	searchHandler := httphandler.NewSearchHandler(searchService)
	notificationHandler := httphandler.NewNotificationHandler(notificationService)
	sessionHandler := httphandler.NewSessionHandler(sessionService)

	// Configure HTTP router with all endpoints and middleware
	router := httphandler.SetupRouter(userHandler, followHandler, suggestionHandler, storyHandler, contentHandler, reactionHandler, feedHandler, bookmarkHandler, searchHandler, notificationHandler, sessionHandler, sessionService, cfg)

	// Configure HTTP server with appropriate timeouts and settings
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,  // Timeout for reading request body
		WriteTimeout: 10 * time.Second, // Timeout for writing response
		IdleTimeout:  15 * time.Second, // Timeout for idle connections
	}

	// Start the HTTP server and begin accepting requests
	log.Info().Str("port", cfg.Port).Msg("Starting Vybes API server")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// startNATSWorker initializes and starts a NATS worker for processing
// notification events asynchronously. It subscribes to the notification
// subject and processes incoming messages.
//
// Parameters:
//   - cfg: Application configuration containing NATS connection details
//   - notificationService: Service for creating notifications
func startNATSWorker(cfg *config.Config, notificationService service.NotificationService) {
	// Connect to NATS message broker
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS for worker")
	}

	// Subscribe to notification events and process them
	_, err = nc.Subscribe(service.NotificationSubject, func(m *nats.Msg) {
		var event domain.Notification
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal notification event from NATS")
			return
		}

		// Log the received notification event for debugging
		log.Info().
			Str("recipient_id", event.UserID.Hex()).
			Str("actor_id", event.ActorID.Hex()).
			Str("type", string(event.Type)).
			Msg("Received notification event from NATS")

		// Create notification in database using the event data
		// The original CreateNotification service method expected metadata as the last argument.
		// Now we pass the specific PostID from the event for better context.
		if err := notificationService.CreateNotification(context.Background(), event.UserID, event.ActorID, event.Type, event.PostID); err != nil {
			log.Error().Err(err).Msg("Failed to create notification from NATS event")
		}
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to NATS subject")
	}

	log.Info().Str("subject", service.NotificationSubject).Msg("NATS worker subscribed and listening")
	// Keep the worker running indefinitely to process events
	select {}
}
