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

	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Info: No .env file found (using system environment variables): %v\n", err)
	}

	// 2. Get Connection String
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Fallback for individual env vars if DATABASE_URL is missing
		host := os.Getenv("SUPABASE_DB_HOST")
		user := os.Getenv("SUPABASE_DB_USER")
		pass := os.Getenv("SUPABASE_DB_PASSWORD")
		name := os.Getenv("SUPABASE_DB_NAME")
		port := os.Getenv("SUPABASE_DB_PORT") // Should be 5432

		if host != "" && user != "" {
			if port == "" {
				port = "5432"
			}
			dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require", host, user, pass, name, port)
		} else {
			log.Fatal("Error: DATABASE_URL not set. Please set it in your .env file.")
		}
	}

	// 3. Connect to Supabase (Session Mode / Port 5432)
	// We use the standard logger configuration
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info, // Info helps debug connection issues
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	fmt.Println("Connecting to Supabase (Session Mode)...")

	// Standard GORM connection.
	// We do NOT disable PrepareStmt because we are using Port 5432 (Session Mode).
	// This allows Go to cache statements for better performance.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// 4. Enable Required Extensions
	// We use a separate session context for this to prevent timeouts
	setupExtensions(db)

	// 5. Configure Connection Pool
	// This optimization is important for long-running Go services
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error getting generic database object: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping to verify
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	fmt.Println("🚀 Connected to Supabase successfully")

	// 6. Wrap for Application
	currentDB := database.New(db)

	// 7. Handle CLI Flags
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

	// 8. Start Server
	startServer(currentDB)
}

func setupExtensions(db *gorm.DB) {
	// Use a higher timeout for extension creation
	ctx := db.Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Warn),
	})

	if err := ctx.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Fatalf("Error enabling uuid-ossp extension: %v", err)
	}
	if err := ctx.Exec("CREATE EXTENSION IF NOT EXISTS \"vector\"").Error; err != nil {
		log.Fatalf("Error enabling vector extension: %v", err)
	}
}

func startServer(db database.Database) {
	errChannel := make(chan error)
	defer close(errChannel)

	server, err := api.NewServer(db)
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	go server.Start(errChannel)
	go listenToInterrupt(errChannel)

	fatalErr := <-errChannel
	fmt.Printf("Closing server: %v\n", fatalErr)

	server.ShutdownGracefully(30 * time.Second)
}

func listenToInterrupt(errChannel chan<- error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	errChannel <- fmt.Errorf("%s", <-c)
}
