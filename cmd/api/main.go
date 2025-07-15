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

func main() {
	// Setup structured logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.DBName)
	log.Info().Str("database", db.Name()).Msg("Successfully connected to MongoDB")

	// Setup database indexes
	repository.SetupIndexes(context.Background(), db)

	// Initialize storage client
	storageClient, err := storage.NewClient(context.Background(), cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage client")
	}

	// Initialize cache client
	cacheClient, err := cache.NewClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize cache client")
	}

	// Initialize NATS Notification Publisher
	notificationPublisher, err := service.NewNATSNotificationPublisher(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize NATS notification publisher")
	}

	// Initialize repositories
	userRepository := repository.NewMongoUserRepository(db)
	followRepository := repository.NewMongoFollowRepository(db)
	counterRepository := repository.NewMongoCounterRepository(db)
	storyRepository := repository.NewMongoStoryRepository(db)
	contentRepository := repository.NewMongoContentRepository(db)
	reactionRepository := repository.NewMongoReactionRepository(db)
	bookmarkRepository := repository.NewMongoBookmarkRepository(db)
	notificationRepository := repository.NewMongoNotificationRepository(db)

	// Initialize services
	emailService := service.NewResendEmailService(cfg)
	walletService := service.NewWalletService(cfg)
	notificationService := service.NewNotificationService(notificationRepository)
	userService := service.NewUserService(userRepository, followRepository, counterRepository, walletService, emailService, cacheClient, cfg.JWTSecret, cfg.WalletEncryptionKey)
	followService := service.NewFollowService(followRepository, userRepository, notificationPublisher)
	suggestionService := service.NewSuggestionService(userRepository, followRepository)
	storyService := service.NewStoryService(storyRepository, followRepository, storageClient, cfg)
	contentService := service.NewContentService(contentRepository, userRepository, storageClient, notificationPublisher, cfg)
	reactionService := service.NewReactionService(reactionRepository, contentRepository, userRepository, notificationPublisher)
	feedService := service.NewFeedService(contentRepository, followRepository)
	bookmarkService := service.NewBookmarkService(bookmarkRepository, contentRepository)
	searchService := service.NewSearchService(userRepository)
	cronService := service.NewCronService(cfg, storyRepository, storageClient)

	// Start NATS worker in the background
	go startNATSWorker(cfg, notificationService)

	// Start cron jobs in the background
	go cronService.Start()

	// Initialize handlers
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

	// Setup router
	router := httphandler.SetupRouter(userHandler, followHandler, suggestionHandler, storyHandler, contentHandler, reactionHandler, feedHandler, bookmarkHandler, searchHandler, notificationHandler, cfg)

	// Setup robust HTTP server
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// Start server
	log.Info().Str("port", cfg.Port).Msg("Server starting")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

func startNATSWorker(cfg *config.Config, notificationService service.NotificationService) {
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS for worker")
	}

	_, err = nc.Subscribe(service.NotificationSubject, func(m *nats.Msg) {
		var event domain.Notification
		if err := json.Unmarshal(m.Data, &event); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal notification event from NATS")
			return
		}

		log.Info().
			Str("recipient_id", event.UserID.Hex()).
			Str("actor_id", event.ActorID.Hex()).
			Str("type", string(event.Type)).
			Msg("Received notification event from NATS")

		// The original CreateNotification service method expected metadata as the last argument.
		// Now we pass the specific PostID from the event.
		if err := notificationService.CreateNotification(context.Background(), event.UserID, event.ActorID, event.Type, event.PostID); err != nil {
			log.Error().Err(err).Msg("Failed to create notification from NATS event")
		}
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to subscribe to NATS subject")
	}

	log.Info().Str("subject", service.NotificationSubject).Msg("NATS worker subscribed and listening")
	// Keep the worker running
	select {}
}