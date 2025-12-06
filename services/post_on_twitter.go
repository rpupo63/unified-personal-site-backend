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

// TwitterPostResponse represents the response from Twitter API v2
type TwitterPostResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

// TwitterErrorResponse represents an error response from Twitter API
type TwitterErrorResponse struct {
	Errors []struct {
		Message string `json:"message"`
		Code    int    `json:"code,omitempty"`
	} `json:"errors"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
	Type   string `json:"type,omitempty"`
}

// PostToTwitter posts a blog post to Twitter using the Twitter API v2
// It formats the blog post content with title, summary/content, and tags as hashtags
// Loads configuration from .env file in the backend root directory
// Requires environment variables in .env:
//   - TWITTER_BEARER_TOKEN: OAuth 2.0 Bearer token with tweet.write permission
//   - TWITTER_BASE_URL: Optional base URL for constructing blog post links (defaults to empty if not set)
func PostToTwitter(blogPost models.BlogPost, tags []models.BlogTag) error {
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
	bearerToken := config.GetString(cfg, "TWITTER_BEARER_TOKEN", "")
	if bearerToken == "" {
		return fmt.Errorf("TWITTER_BEARER_TOKEN environment variable is required in .env file")
	}

	baseURL := config.GetString(cfg, "TWITTER_BASE_URL", "")

	// Construct the post text
	postText := buildTwitterPostText(blogPost, tags, baseURL)

	// Build the Twitter API payload
	payload := buildTwitterPayload(postText)

	// Marshal payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Twitter payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.twitter.com/2/tweets", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Twitter API request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Twitter API: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Twitter API response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		var errorResp TwitterErrorResponse
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			if len(errorResp.Errors) > 0 {
				return fmt.Errorf("twitter API error (status %d): %s", resp.StatusCode, errorResp.Errors[0].Message)
			}
			if errorResp.Detail != "" {
				return fmt.Errorf("twitter API error (status %d): %s", resp.StatusCode, errorResp.Detail)
			}
		}
		return fmt.Errorf("twitter API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse successful response
	var postResponse TwitterPostResponse
	if err := json.Unmarshal(bodyBytes, &postResponse); err != nil {
		log.Warn().Err(err).Msg("Failed to parse Twitter post response, but post was created")
	} else {
		log.Info().Str("tweetId", postResponse.Data.ID).Msg("Successfully posted to Twitter")
	}

	return nil
}

// buildTwitterPostText constructs the text content for the Twitter post
// Twitter has a 280 character limit, so we need to be more concise
func buildTwitterPostText(blogPost models.BlogPost, tags []models.BlogTag, baseURL string) string {
	var parts []string

	// Add title
	if blogPost.Title != "" {
		parts = append(parts, blogPost.Title)
	}

	// Add summary or truncated content
	// Twitter has 280 char limit, so we need to be more aggressive with truncation
	if blogPost.Summary != nil && *blogPost.Summary != "" {
		summary := *blogPost.Summary
		// Truncate summary if too long (leave room for URL and hashtags)
		maxSummaryLength := 150
		if len(summary) > maxSummaryLength {
			truncated := summary[:maxSummaryLength]
			lastPeriod := strings.LastIndex(truncated, ".")
			if lastPeriod > maxSummaryLength/2 {
				summary = truncated[:lastPeriod+1] + "..."
			} else {
				summary = truncated + "..."
			}
		}
		parts = append(parts, summary)
	} else if blogPost.Content != "" {
		// Truncate content to fit Twitter's 280 char limit
		content := blogPost.Content
		maxContentLength := 150 // Leave room for title, tags, and URL
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

	// Add URL if available (URLs count as 23 chars in Twitter)
	if blogPost.URL != nil && *blogPost.URL != "" {
		parts = append(parts, *blogPost.URL)
	} else if baseURL != "" {
		// Construct URL from base URL and post ID
		url := fmt.Sprintf("%s/blog/%s", strings.TrimSuffix(baseURL, "/"), blogPost.ID.String())
		parts = append(parts, url)
	}

	// Add hashtags from tags (limit to 3-4 hashtags to save space)
	if len(tags) > 0 {
		var hashtags []string
		maxHashtags := 4
		for i, tag := range tags {
			if i >= maxHashtags {
				break
			}
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

	// Join parts and ensure total length is within Twitter's 280 character limit
	postText := strings.Join(parts, "\n\n")

	// If still too long, truncate more aggressively
	if len(postText) > 280 {
		// Try to keep title and URL, truncate content more
		title := ""
		url := ""
		hashtags := ""

		if blogPost.Title != "" {
			title = blogPost.Title + "\n\n"
		}

		// Extract URL
		if blogPost.URL != nil && *blogPost.URL != "" {
			url = "\n\n" + *blogPost.URL
		} else if baseURL != "" {
			url = "\n\n" + fmt.Sprintf("%s/blog/%s", strings.TrimSuffix(baseURL, "/"), blogPost.ID.String())
		}

		// Extract hashtags
		if len(tags) > 0 {
			var hashtagList []string
			maxHashtags := 3
			for i, tag := range tags {
				if i >= maxHashtags {
					break
				}
				hashtag := FormatHashtag(tag.Value)
				if hashtag != "" {
					hashtagList = append(hashtagList, "#"+hashtag)
				}
			}
			if len(hashtagList) > 0 {
				hashtags = "\n\n" + strings.Join(hashtagList, " ")
			}
		}

		// Calculate available space for content
		availableSpace := 280 - len(title) - len(url) - len(hashtags)
		if availableSpace < 50 {
			availableSpace = 50 // Minimum space for content
		}

		// Get content
		content := ""
		if blogPost.Summary != nil && *blogPost.Summary != "" {
			content = *blogPost.Summary
		} else if blogPost.Content != "" {
			content = blogPost.Content
		}

		// Truncate content to fit
		if len(content) > availableSpace {
			truncated := content[:availableSpace-3]
			lastPeriod := strings.LastIndex(truncated, ".")
			if lastPeriod > availableSpace/2 {
				content = truncated[:lastPeriod+1] + "..."
			} else {
				content = truncated + "..."
			}
		}

		postText = title + content + url + hashtags

		// Final safety check - if still too long, hard truncate
		if len(postText) > 280 {
			postText = postText[:277] + "..."
		}
	}

	return postText
}

// buildTwitterPayload constructs the Twitter API v2 payload
// Twitter API v2 only requires the text field for a simple tweet
func buildTwitterPayload(postText string) map[string]interface{} {
	payload := map[string]interface{}{
		"text": postText,
	}

	return payload
}
