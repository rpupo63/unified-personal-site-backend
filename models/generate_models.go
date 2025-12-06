package models

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

/*
Column Mismatch Report Usage:

This file contains functionality to generate a report of database columns that aren't
accounted for as variables in the corresponding Go model structs.

To generate the report:

1. Set the environment variable: GENERATE_COLUMN_REPORT=true
2. Run the application: go run main.go

The report will show:
- Each table name
- List of columns that exist in the database but not in the Go model
- Summary of total mismatched columns across all tables

Example output:
=== COLUMN MISMATCH REPORT ===
--- Table: experts ---
Found 2 columns not accounted for in model:
  - created_at
  - updated_at

--- Table: organizations ---
All columns are accounted for in the model.

=== SUMMARY ===
Total mismatched columns across all tables: 2
*/

func GenerateModels(db *gorm.DB) {
	// First, ensure the database is ready
	if err := db.Exec("SELECT 1").Error; err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	// Set up verbose logging for migration
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             0,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)
	db = db.Session(&gorm.Session{
		Logger: newLogger,
		// Skip data validation during migration
		SkipDefaultTransaction: true,
		PrepareStmt:            false,
	})

	g := gen.NewGenerator(gen.Config{
		OutPath:           "./generated",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})

	// Use GORM's database connection
	g.UseDB(db)

	// Specify models for which to generate code
	g.ApplyBasic(
		BlogPost{},
		BlogTag{},
		Project{},
		ProjectTag{},
	)

	fmt.Println("Starting database migration...")

	// Create a new session for migration with specific settings
	migrateDB := db.Session(&gorm.Session{
		SkipDefaultTransaction: true,
		PrepareStmt:            false,
		Logger:                 newLogger,
	})

	// Migrate all models
	fmt.Println("Migrating models...")
	if err := migrateDB.AutoMigrate(
		&BlogPost{},
		&BlogTag{},
		&Project{},
		&ProjectTag{},
	); err != nil {
		fmt.Printf("Error during models migration: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Database migration completed successfully!")

	// Generate column mismatch report
	GenerateColumnMismatchReport(db)

	// Execute the code generation
	g.Execute()
	fmt.Println("Model generation complete!")
}

// GenerateColumnMismatchReport generates a report of database columns that aren't accounted for in Go models
func GenerateColumnMismatchReport(db *gorm.DB) {
	fmt.Println("=== COLUMN MISMATCH REPORT ===")
	fmt.Println("Generating report of database columns not accounted for in Go models...")

	// Define model mappings (table name -> struct type)
	modelMappings := map[string]interface{}{
		"blog_posts":   BlogPost{},
		"blog_tags":    BlogTag{},
		"projects":     Project{},
		"project_tags": ProjectTag{},
	}

	totalMismatches := 0

	for tableName, modelStruct := range modelMappings {
		fmt.Printf("\n--- Table: %s ---\n", tableName)

		// Get database columns
		dbColumns, err := getTableColumns(db, tableName)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				fmt.Printf("Table does not exist yet (will be created during migration)\n")
			} else {
				fmt.Printf("Error getting columns for table %s: %v\n", tableName, err)
			}
			continue
		}

		// Get model fields
		modelFields := getModelFields(modelStruct)

		// Find mismatches
		mismatches := findColumnMismatches(dbColumns, modelFields)

		if len(mismatches) > 0 {
			fmt.Printf("Found %d columns not accounted for in model:\n", len(mismatches))
			for _, col := range mismatches {
				fmt.Printf("  - %s\n", col)
			}
			totalMismatches += len(mismatches)
		} else {
			fmt.Println("All columns are accounted for in the model.")
		}
	}

	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Total mismatched columns across all tables: %d\n", totalMismatches)
}

// getTableColumns retrieves column names from a database table
func getTableColumns(db *gorm.DB, tableName string) ([]string, error) {
	var columns []string
	query := `
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_name = ? 
		AND table_schema = CURRENT_SCHEMA()
		ORDER BY ordinal_position
	`

	err := db.Raw(query, tableName).Scan(&columns).Error
	if err != nil {
		return nil, fmt.Errorf("error querying columns for table %s: %w", tableName, err)
	}

	// Check if table exists
	if len(columns) == 0 {
		// Verify if table exists
		var tableExists bool
		tableQuery := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = CURRENT_SCHEMA() 
				AND table_name = ?
			)
		`
		if err := db.Raw(tableQuery, tableName).Scan(&tableExists).Error; err != nil {
			return nil, fmt.Errorf("error checking if table %s exists: %w", tableName, err)
		}

		if !tableExists {
			return nil, fmt.Errorf("table %s does not exist", tableName)
		}
	}

	return columns, nil
}

// getModelFields extracts field names from a Go struct using reflection
func getModelFields(model interface{}) []string {
	var fields []string
	t := reflect.TypeOf(model)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip embedded structs (foreign key relationships)
		if field.Anonymous {
			continue
		}

		// Get the GORM column name from the tag
		gormTag := field.Tag.Get("gorm")
		if gormTag != "" {
			// Parse the gorm tag to find the column name
			columnName := extractColumnNameFromGormTag(gormTag)
			if columnName != "" {
				fields = append(fields, columnName)
			}
		}
	}

	return fields
}

// extractColumnNameFromGormTag extracts the column name from a GORM tag
func extractColumnNameFromGormTag(gormTag string) string {
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// findColumnMismatches finds columns that exist in the database but not in the model
func findColumnMismatches(dbColumns, modelFields []string) []string {
	modelFieldSet := make(map[string]bool)
	for _, field := range modelFields {
		modelFieldSet[field] = true
	}

	var mismatches []string
	for _, col := range dbColumns {
		if !modelFieldSet[col] {
			mismatches = append(mismatches, col)
		}
	}

	return mismatches
}

// GenerateColumnMismatchReportStandalone generates a report without running migrations
func GenerateColumnMismatchReportStandalone(db *gorm.DB) {
	// First, ensure the database is ready
	if err := db.Exec("SELECT 1").Error; err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	GenerateColumnMismatchReport(db)
}
