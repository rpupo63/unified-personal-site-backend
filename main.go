package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	api "github.com/rpupo63/unified-personal-site-backend/api"
	"github.com/rpupo63/unified-personal-site-backend/database"
	_ "github.com/rpupo63/unified-personal-site-backend/docs" // Swagger docs
	"github.com/rpupo63/unified-personal-site-backend/models"
)

// @title           Personal Site API
// @version         1.0
// @description     API for managing personal site projects and blog posts
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @schemes   http https

func main() {
	fmt.Println("Initializing app...")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// -------------------------------------------------------------------------
	// DATABASE CONNECTION LOGIC
	// -------------------------------------------------------------------------
	var db *gorm.DB
	var err error
	var currentDB database.Database

	// Priority 1: Full Connection String (Recommended for Pooler)
	connStr := getEnv("DATABASE_URL", "")
	if connStr == "" {
		connStr = getEnv("SUPABASE_DB_URL", "")
	}

	// Priority 2: Individual Components
	if connStr == "" {
		host := getEnv("SUPABASE_DB_HOST", "")
		// Note: Supabase Transaction Pooler uses port 6543
		port := getEnv("SUPABASE_DB_PORT", "6543")
		user := getEnv("SUPABASE_DB_USER", "")
		password := getEnv("SUPABASE_DB_PASSWORD", "")
		dbname := getEnv("SUPABASE_DB_NAME", "postgres")

		if host == "" || user == "" || password == "" {
			fmt.Printf("Error: Missing required database configuration.\n")
			fmt.Printf("Set DATABASE_URL or (SUPABASE_DB_HOST, SUPABASE_DB_USER, SUPABASE_DB_PASSWORD)\n")
			os.Exit(1)
		}

		connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
			host, user, password, dbname, port)
	} else {
		// Validate provided string
		normalized, err := normalizeConnectionString(connStr)
		if err != nil {
			fmt.Printf("Error: Invalid connection string: %v\n", err)
			os.Exit(1)
		}
		connStr = normalized
	}

	fmt.Println("Connecting to Supabase (Transaction Pooler)...")

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// -------------------------------------------------------------------------
	// CRITICAL SUPABASE POOLER CONFIGURATION
	// -------------------------------------------------------------------------
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN: connStr,
		// PreferSimpleProtocol is CRITICAL for the Transaction Pooler (port 6543).
		// It disables the extended query protocol which creates prepared statements.
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		// PrepareStmt must be FALSE. The pooler does not support prepared statements
		// in Transaction mode.
		PrepareStmt: false,
		Logger:      newLogger,
	})

	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "network is unreachable") {
			fmt.Printf("Network Error: Unreachable. (Check IPv6 settings or Host address)\n")
		} else {
			fmt.Printf("Error connecting to database: %v\n", err)
		}
		os.Exit(1)
	}

	// Enable required PostgreSQL extensions
	// Note: 'vector' extension is required for Embeddings/AI features
	// Use a separate session with a higher slow threshold for extension creation
	// to avoid warnings during startup (extension creation can be slow on first run)
	extensionLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             1 * time.Second, // Higher threshold for extension creation
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	extensionDB := db.Session(&gorm.Session{
		Logger: extensionLogger,
	})

	if err := extensionDB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		fmt.Printf("Error enabling uuid-ossp extension: %v\n", err)
		os.Exit(1)
	}
	if err := extensionDB.Exec("CREATE EXTENSION IF NOT EXISTS \"vector\"").Error; err != nil {
		fmt.Printf("Error enabling vector extension: %v\n", err)
		os.Exit(1)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("Error getting generic database object: %v\n", err)
		os.Exit(1)
	}

	// Set connection pool settings to prevent opening too many connections
	// in the container, though the Supabase Pooler handles the hard limit.
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		fmt.Printf("Error pinging database: %v\n", err)
		os.Exit(1)
	}

	currentDB = database.New(db)

	// Handle Generation flags
	if strings.ToLower(os.Getenv("GENERATE_MODELS")) == "true" {
		fmt.Println("Generating models...")
		models.GenerateModels(db)
		return
	}

	if os.Getenv("GENERATE_COLUMN_REPORT") == "true" {
		fmt.Println("Generating column report...")
		models.GenerateColumnMismatchReportStandalone(db)
		return
	}

	// Initialize Server
	errChannel := make(chan error)
	defer close(errChannel)

	server, err := api.NewServer(currentDB)
	if err != nil {
		fmt.Printf("Error initializing server: %v\n", err)
		os.Exit(1)
	}

	go server.Start(errChannel)
	go listenToInterrupt(errChannel)

	fatalErr := <-errChannel
	fmt.Printf("Closing server: %v\n", fatalErr)

	server.ShutdownGracefully(30 * time.Second)
}

// listenToInterrupt waits for SIGINT or SIGTERM
func listenToInterrupt(errChannel chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	errChannel <- fmt.Errorf("%s", <-c)
}

// getEnv returns the value or fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// normalizeConnectionString validates and standardizes the connection string
func normalizeConnectionString(connStr string) (string, error) {
	connStr = strings.TrimSpace(connStr)
	if connStr == "" {
		return "", fmt.Errorf("connection string is empty")
	}

	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		return normalizeURLConnectionString(connStr)
	}
	return normalizeKeyValueConnectionString(connStr)
}

func normalizeURLConnectionString(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return "", fmt.Errorf("missing host")
	}

	port := parsedURL.Port()
	if port == "" {
		port = "5432"
	} // Default fallback, though Pooler is 6543

	user := parsedURL.User.Username()
	if user == "" {
		return "", fmt.Errorf("missing user")
	}

	password, hasPassword := parsedURL.User.Password()
	if !hasPassword {
		return "", fmt.Errorf("missing password")
	}

	dbname := strings.TrimPrefix(parsedURL.Path, "/")
	if dbname == "" {
		return "", fmt.Errorf("missing dbname")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		host, port, user, password, dbname)

	sslmode := parsedURL.Query().Get("sslmode")
	if sslmode == "" {
		sslmode = "require"
	}
	connStr += fmt.Sprintf(" sslmode=%s", sslmode)

	return connStr, nil
}

func normalizeKeyValueConnectionString(connStr string) (string, error) {
	if !strings.Contains(connStr, "host=") {
		return "", fmt.Errorf("missing 'host='")
	}
	if !strings.Contains(connStr, "user=") {
		return "", fmt.Errorf("missing 'user='")
	}
	if !strings.Contains(connStr, "dbname=") && !strings.Contains(connStr, "database=") {
		return "", fmt.Errorf("missing 'dbname='")
	}
	if !strings.Contains(connStr, "sslmode=") {
		connStr += " sslmode=require"
	}
	return connStr, nil
}
