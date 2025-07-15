package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	Port                string
	MongoURI            string
	DBName              string
	JWTSecret           string
	ResendAPIKey        string
	SenderEmail         string
	WalletEncryptionKey string
	EthRPCURL           string

	// MinIO Configuration
	MinioEndpoint      string
	MinioAccessKey     string
	MinioSecretKey     string
	MinioUseSSL        bool
	MinioPostsBucket   string
	MinioStoriesBucket string

	// Redis Configuration
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// NATS Configuration
	NatsURL string
}

// LoadConfig loads configuration from environment variables or a .env file
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // Default port
	}

	useSSL, err := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	if err != nil {
		useSSL = false // Default to false
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redisDB = 0 // Default DB
	}

	return &Config{
		Port:                port,
		MongoURI:            os.Getenv("MONGO_URI"),
		DBName:              os.Getenv("DB_NAME"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		ResendAPIKey:        os.Getenv("RESEND_API_KEY"),
		SenderEmail:         os.Getenv("SENDER_EMAIL"),
		WalletEncryptionKey: os.Getenv("WALLET_ENCRYPTION_KEY"),
		EthRPCURL:           os.Getenv("ETH_RPC_URL"),
		MinioEndpoint:       os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:      os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:      os.Getenv("MINIO_SECRET_KEY"),
		MinioUseSSL:         useSSL,
		MinioPostsBucket:    os.Getenv("MINIO_POSTS_BUCKET"),
		MinioStoriesBucket:  os.Getenv("MINIO_STORIES_BUCKET"),
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		RedisDB:             redisDB,
		NatsURL:             os.Getenv("NATS_URL"),
	}, nil
}