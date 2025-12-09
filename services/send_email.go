package services

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
	"github.com/rpupo63/unified-personal-site-backend/config"
	"github.com/rs/zerolog/log"
)

// SendEmail sends an email using the Resend API
// Parameters:
//   - subject: The email subject line
//   - body: The email body (HTML or plain text)
//   - recipients: A list of recipient email addresses
//
// Requires environment variables in .env:
//   - RESEND_API_KEY: Your Resend API key
//   - RESEND_FROM_EMAIL: The sender email address (e.g., "Your Name <[email protected]>")
func SendEmail(subject, body string, recipients []string) error {
	if len(recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	// Load .env file from backend root directory
	// Try multiple possible paths to find the .env file
	possiblePaths := []string{
		".env",                           // Current directory (if running from backend/)
		filepath.Join("..", ".env"),      // Parent directory
		filepath.Join("backend", ".env"), // backend/.env from project root
	}

	var envLoaded bool
	for _, envPath := range possiblePaths {
		if err := godotenv.Load(envPath); err == nil {
			envLoaded = true
			log.Debug().Str("path", envPath).Msg("Loaded .env file")
			break
		}
	}

	if !envLoaded {
		log.Warn().Msg("Failed to load .env file from any expected location, using existing environment variables")
	}

	// Get config from environment variables
	cfg := config.New()

	// Get required configuration
	apiKey := config.GetString(cfg, "RESEND_API_KEY", "")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY environment variable is required in .env file")
	}

	fromEmail := config.GetString(cfg, "RESEND_FROM_EMAIL", "")
	if fromEmail == "" {
		return fmt.Errorf("RESEND_FROM_EMAIL environment variable is required in .env file")
	}

	// Create Resend client
	client := resend.NewClient(apiKey)

	// Build the email request
	params := &resend.SendEmailRequest{
		From:    fromEmail,
		To:      recipients,
		Subject: subject,
		Html:    body,
	}

	// Send email
	sent, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	log.Info().Str("emailId", sent.Id).Msg("Successfully sent email via Resend")

	return nil
}
