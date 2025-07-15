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

	// R2 Configuration
	R2Endpoint        string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2PostsBucket     string
	R2StoriesBucket   string

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
		R2Endpoint:          os.Getenv("R2_ENDPOINT"),
		R2AccessKeyID:       os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:   os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2PostsBucket:       os.Getenv("R2_POSTS_BUCKET"),
		R2StoriesBucket:     os.Getenv("R2_STORIES_BUCKET"),
		RedisAddr:           os.Getenv("REDIS_ADDR"),
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		RedisDB:             redisDB,
		NatsURL:             os.Getenv("NATS_URL"),
	}, nil
}