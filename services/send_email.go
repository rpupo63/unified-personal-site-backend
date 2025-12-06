package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/ProNexus-Startup/ProNexus/backend/config"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// ResendEmailRequest represents the request payload for Resend API
type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// ResendEmailResponse represents the response from Resend API
type ResendEmailResponse struct {
	ID string `json:"id"`
}

// ResendErrorResponse represents an error response from Resend API
type ResendErrorResponse struct {
	Message string `json:"message"`
}

// SendEmail sends an email using the Resend API
// Parameters:
//   - subject: The email subject line
//   - body: The email body (HTML or plain text)
//   - recipients: A list of recipient email addresses
//
// Requires environment variables in .env:
//   - RESEND_API_KEY: Your Resend API key
//   - RESEND_FROM_EMAIL: The sender email address (e.g., "Your Name <[email protected]>")
//
// Optional environment variables:
//   - RESEND_FROM_EMAIL: If not provided, defaults to a generic sender format
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

	// Build the Resend API payload
	payload := ResendEmailRequest{
		From:    fromEmail,
		To:      recipients,
		Subject: subject,
		Html:    body, // Assuming body is HTML by default
	}

	// Marshal payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Resend API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Resend API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Resend API response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResp ResendErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, errorResp.Message)
		}
		return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse successful response
	var emailResponse ResendEmailResponse
	if err := json.Unmarshal(bodyBytes, &emailResponse); err != nil {
		log.Warn().Err(err).Msg("Failed to parse Resend email response, but email was sent")
	} else {
		log.Info().Str("emailId", emailResponse.ID).Msg("Successfully sent email via Resend")
	}

	return nil
}
