package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	api "github.com/ProNexus-Startup/ProNexus/backend/api"
	"github.com/ProNexus-Startup/ProNexus/backend/database"
	"github.com/ProNexus-Startup/ProNexus/backend/models"
)

func main() {
	fmt.Println("Initializing app...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	dbType := os.Getenv("DB_TYPE")
	var db *gorm.DB
	var err error
	var currentDB database.Database

	// Build connection string based on DB_TYPE
	var connStr string
	fmt.Printf("DB_TYPE: %s\n", dbType)
	switch dbType {
	case "supa":
		connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
			getEnv("SUPABASE_DB_HOST", ""),
			getEnv("SUPABASE_DB_USER", ""),
			getEnv("SUPABASE_DB_PASSWORD", ""),
			getEnv("SUPABASE_DB_NAME", ""),
			getEnv("SUPABASE_DB_PORT", "5432"),
		)
		fmt.Println("Connecting to Supabase database...")
	default:
		fmt.Println("Unsupported DB_TYPE. Exiting...")
		os.Exit(1)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             10 * time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  connStr,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false,
		Logger:      newLogger,
	})
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	// Enable required PostgreSQL extensions
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		fmt.Printf("Error enabling uuid-ossp extension: %v\n", err)
		os.Exit(1)
	}
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"vector\"").Error; err != nil {
		fmt.Printf("Error enabling vector extension: %v\n", err)
		os.Exit(1)
	}

	// Test database connection
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		fmt.Printf("Error testing database connection: %v\n", err)
		os.Exit(1)
	}

	currentDB = database.New(db)

	// If generating models, run generation and exit
	if strings.ToLower(os.Getenv("GENERATE_MODELS")) == "true" {
		fmt.Println("Generating models and query helpers...")
		models.GenerateModels(db)
		return
	}

	// If generating column mismatch report, run report and exit
	if os.Getenv("GENERATE_COLUMN_REPORT") == "true" {
		fmt.Println("Generating column mismatch report...")
		models.GenerateColumnMismatchReportStandalone(db)
		return
	}

	errChannel := make(chan error)
	defer close(errChannel)

	server, err := api.NewServer(currentDB)
	if err != nil {
		fmt.Printf("Error initializing server: %v\n", err)
		os.Exit(1)
	}

	go server.Start(errChannel)

	// Listen for interrupt signals to gracefully shutdown the server
	go listenToInterrupt(errChannel)

	fatalErr := <-errChannel
	fmt.Printf("Closing server: %v\n", fatalErr)

	server.ShutdownGracefully(30 * time.Second)
}

// listenToInterrupt waits for SIGINT or SIGTERM and then sends an error to the error channel.
func listenToInterrupt(errChannel chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	errChannel <- fmt.Errorf("%s", <-c)
}

// getEnv returns the value of the environment variable key or a fallback value.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
