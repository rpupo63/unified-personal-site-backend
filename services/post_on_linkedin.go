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

// LinkedInPostResponse represents the response from LinkedIn API
type LinkedInPostResponse struct {
	ID string `json:"id"`
}

// LinkedInErrorResponse represents an error response from LinkedIn API
type LinkedInErrorResponse struct {
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode,omitempty"`
}

// PostToLinkedIn posts a blog post to LinkedIn using the LinkedIn UGC Posts API
// It formats the blog post content with title, summary/content, and tags as hashtags
// Loads configuration from .env file in the backend root directory
// Requires environment variables in .env:
//   - LINKEDIN_ACCESS_TOKEN: OAuth 2.0 access token with w_member_social permission
//   - LINKEDIN_PERSON_URN: Your LinkedIn person URN (e.g., "urn:li:person:YOUR_ID")
//   - LINKEDIN_BASE_URL: Optional base URL for constructing blog post links (defaults to empty if not set)
func PostToLinkedIn(blogPost models.BlogPost, tags []models.BlogTag) error {
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
	accessToken := config.GetString(cfg, "LINKEDIN_ACCESS_TOKEN", "")
	if accessToken == "" {
		return fmt.Errorf("LINKEDIN_ACCESS_TOKEN environment variable is required in .env file")
	}

	personURN := config.GetString(cfg, "LINKEDIN_PERSON_URN", "")
	if personURN == "" {
		return fmt.Errorf("LINKEDIN_PERSON_URN environment variable is required in .env file")
	}

	baseURL := config.GetString(cfg, "LINKEDIN_BASE_URL", "")

	// Construct the post text
	postText := buildLinkedInPostText(blogPost, tags, baseURL)

	// Build the LinkedIn API payload
	payload := buildLinkedInPayload(personURN, postText, blogPost.URL)

	// Marshal payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal LinkedIn payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.linkedin.com/v2/ugcPosts", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create LinkedIn API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to LinkedIn API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read LinkedIn API response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		var errorResp LinkedInErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			return fmt.Errorf("LinkedIn API error (status %d): %s", resp.StatusCode, errorResp.Message)
		}
		return fmt.Errorf("LinkedIn API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse successful response
	var postResponse LinkedInPostResponse
	if err := json.Unmarshal(bodyBytes, &postResponse); err != nil {
		log.Warn().Err(err).Msg("Failed to parse LinkedIn post response, but post was created")
	} else {
		log.Info().Str("postId", postResponse.ID).Msg("Successfully posted to LinkedIn")
	}

	return nil
}

// buildLinkedInPostText constructs the text content for the LinkedIn post
func buildLinkedInPostText(blogPost models.BlogPost, tags []models.BlogTag, baseURL string) string {
	var parts []string

	// Add title
	if blogPost.Title != "" {
		parts = append(parts, blogPost.Title)
	}

	// Add summary or truncated content
	if blogPost.Summary != nil && *blogPost.Summary != "" {
		parts = append(parts, *blogPost.Summary)
	} else if blogPost.Content != "" {
		// Truncate content to reasonable length for LinkedIn (max ~3000 chars total)
		content := blogPost.Content
		maxContentLength := 2000 // Leave room for title, tags, and URL
		if len(content) > maxContentLength {
			// Try to truncate at a sentence boundary
			truncated := content[:maxContentLength]
			lastPeriod := strings.LastIndex(truncated, ".")
			if lastPeriod > maxContentLength/2 {
				content = truncated[:lastPeriod+1] + "..."
			} else {
				content = truncated + "..."
			}
		}
		parts = append(parts, content)
	}

	// Add URL if available
	if blogPost.URL != nil && *blogPost.URL != "" {
		parts = append(parts, fmt.Sprintf("Read more: %s", *blogPost.URL))
	} else if baseURL != "" {
		// Construct URL from base URL and post ID
		url := fmt.Sprintf("%s/blog/%s", strings.TrimSuffix(baseURL, "/"), blogPost.ID.String())
		parts = append(parts, fmt.Sprintf("Read more: %s", url))
	}

	// Add hashtags from tags
	if len(tags) > 0 {
		var hashtags []string
		for _, tag := range tags {
			// Format tag as hashtag (remove spaces, special chars)
			hashtag := FormatHashtag(tag.Value)
			if hashtag != "" {
				hashtags = append(hashtags, "#"+hashtag)
			}
		}
		if len(hashtags) > 0 {
			parts = append(parts, strings.Join(hashtags, " "))
		}
	}

	return strings.Join(parts, "\n\n")
}

// buildLinkedInPayload constructs the LinkedIn API payload
func buildLinkedInPayload(personURN, postText string, postURL *string) map[string]interface{} {
	payload := map[string]interface{}{
		"author":         personURN,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]string{
					"text": postText,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]string{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	// If URL is provided, add it as a link
	if postURL != nil && *postURL != "" {
		shareContent := payload["specificContent"].(map[string]interface{})["com.linkedin.ugc.ShareContent"].(map[string]interface{})
		shareContent["shareMediaCategory"] = "ARTICLE"
		shareContent["media"] = []map[string]string{
			{
				"status":      "READY",
				"originalUrl": *postURL,
			},
		}
	}

	return payload
}
