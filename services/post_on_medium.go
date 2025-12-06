package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ProNexus-Startup/ProNexus/backend/config"
	"github.com/ProNexus-Startup/ProNexus/backend/models"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// MediumUserResponse represents the response from Medium API /me endpoint
type MediumUserResponse struct {
	Data struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
	} `json:"data"`
}

// MediumPostResponse represents the response from Medium API when creating a post
type MediumPostResponse struct {
	Data struct {
		ID            string   `json:"id"`
		Title         string   `json:"title"`
		AuthorID      string   `json:"authorId"`
		Tags          []string `json:"tags"`
		URL           string   `json:"url"`
		CanonicalURL  string   `json:"canonicalUrl"`
		PublishStatus string   `json:"publishStatus"`
		PublishedAt   int64    `json:"publishedAt"`
	} `json:"data"`
}

// MediumErrorResponse represents an error response from Medium API
type MediumErrorResponse struct {
	Errors []struct {
		Message string `json:"message"`
		Code    int    `json:"code,omitempty"`
	} `json:"errors"`
}

// PostToMedium posts a blog post to Medium using the Medium API
// It formats the blog post content with title, content, and tags
// Loads configuration from .env file in the backend root directory
// Requires environment variables in .env:
//   - MEDIUM_INTEGRATION_TOKEN: Integration token from Medium settings
//   - MEDIUM_PUBLISH_STATUS: Optional publish status (public, draft, unlisted) - defaults to "public"
//   - MEDIUM_CONTENT_FORMAT: Optional content format (html, markdown) - defaults to "html"
func PostToMedium(blogPost models.BlogPost, tags []models.BlogTag) error {
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
	integrationToken := config.GetString(cfg, "MEDIUM_INTEGRATION_TOKEN", "")
	if integrationToken == "" {
		return fmt.Errorf("MEDIUM_INTEGRATION_TOKEN environment variable is required in .env file")
	}

	publishStatus := config.GetString(cfg, "MEDIUM_PUBLISH_STATUS", "public")
	if publishStatus != "public" && publishStatus != "draft" && publishStatus != "unlisted" {
		publishStatus = "public" // Default to public if invalid
		log.Warn().Str("status", publishStatus).Msg("Invalid MEDIUM_PUBLISH_STATUS, defaulting to 'public'")
	}

	contentFormat := config.GetString(cfg, "MEDIUM_CONTENT_FORMAT", "html")
	if contentFormat != "html" && contentFormat != "markdown" {
		contentFormat = "html" // Default to html if invalid
		log.Warn().Str("format", contentFormat).Msg("Invalid MEDIUM_CONTENT_FORMAT, defaulting to 'html'")
	}

	// First, get the user ID by calling /me endpoint
	userID, err := getMediumUserID(integrationToken)
	if err != nil {
		return fmt.Errorf("failed to get Medium user ID: %w", err)
	}

	// Build the Medium API payload
	payload := buildMediumPayload(blogPost, tags, contentFormat, publishStatus)

	// Marshal payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Medium payload: %w", err)
	}

	// Create HTTP request to create post
	url := fmt.Sprintf("https://api.medium.com/v1/users/%s/posts", userID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Medium API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+integrationToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Medium API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Medium API response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResp MediumErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			if len(errorResp.Errors) > 0 {
				return fmt.Errorf("Medium API error (status %d): %s", resp.StatusCode, errorResp.Errors[0].Message)
			}
		}
		return fmt.Errorf("Medium API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse successful response
	var postResponse MediumPostResponse
	if err := json.Unmarshal(bodyBytes, &postResponse); err != nil {
		log.Warn().Err(err).Msg("Failed to parse Medium post response, but post was created")
	} else {
		log.Info().
			Str("postId", postResponse.Data.ID).
			Str("url", postResponse.Data.URL).
			Str("status", postResponse.Data.PublishStatus).
			Msg("Successfully posted to Medium")
	}

	return nil
}

// getMediumUserID retrieves the user ID from Medium API /me endpoint
func getMediumUserID(integrationToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.medium.com/v1/me", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Medium API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+integrationToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Medium API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Medium API response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var errorResp MediumErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			if len(errorResp.Errors) > 0 {
				return "", fmt.Errorf("Medium API error (status %d): %s", resp.StatusCode, errorResp.Errors[0].Message)
			}
		}
		return "", fmt.Errorf("Medium API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse successful response
	var userResponse MediumUserResponse
	if err := json.Unmarshal(bodyBytes, &userResponse); err != nil {
		return "", fmt.Errorf("failed to parse Medium user response: %w", err)
	}

	if userResponse.Data.ID == "" {
		return "", fmt.Errorf("Medium API returned empty user ID")
	}

	return userResponse.Data.ID, nil
}

// buildMediumPayload constructs the Medium API payload
func buildMediumPayload(blogPost models.BlogPost, tags []models.BlogTag, contentFormat, publishStatus string) map[string]interface{} {
	payload := map[string]interface{}{
		"title":         blogPost.Title,
		"contentFormat": contentFormat,
		"content":       blogPost.Content,
		"publishStatus": publishStatus,
	}

	// Add tags if available
	if len(tags) > 0 {
		var tagList []string
		for _, tag := range tags {
			// Medium tags should be lowercase and can contain spaces/hyphens
			// They don't need hashtag formatting
			tagValue := strings.TrimSpace(tag.Value)
			if tagValue != "" {
				// Medium allows tags with spaces, but we'll keep them simple
				tagList = append(tagList, tagValue)
			}
		}
		if len(tagList) > 0 {
			// Medium allows up to 5 tags
			if len(tagList) > 5 {
				tagList = tagList[:5]
			}
			payload["tags"] = tagList
		}
	}

	// Add canonical URL if available
	if blogPost.URL != nil && *blogPost.URL != "" {
		payload["canonicalUrl"] = *blogPost.URL
	}

	return payload
}
