package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"github.com/rpupo63/unified-personal-site-backend/config"
	"github.com/rpupo63/unified-personal-site-backend/models"
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
//   - TWITTER_API_KEY: OAuth 1.0a API Key (Consumer Key)
//   - TWITTER_API_KEY_SECRET: OAuth 1.0a API Key Secret (Consumer Secret)
//   - TWITTER_ACCESS_TOKEN: OAuth 1.0a Access Token
//   - TWITTER_ACCESS_TOKEN_SECRET: OAuth 1.0a Access Token Secret
//   - BASE_URL: Optional unified base URL for constructing blog post links (defaults to empty if not set)
//   - TWITTER_BASE_URL: Optional platform-specific base URL (fallback for backward compatibility)
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
		log.Debug().Msg("No .env file found, using system environment variables (e.g., from Coolify)")
	}

	// Get config from environment variables
	cfg := config.New()

	// Get required OAuth 1.0a configuration
	apiKey := config.GetString(cfg, "TWITTER_API_KEY", "")
	apiKeySecret := config.GetString(cfg, "TWITTER_API_KEY_SECRET", "")
	accessToken := config.GetString(cfg, "TWITTER_ACCESS_TOKEN", "")
	accessTokenSecret := config.GetString(cfg, "TWITTER_ACCESS_TOKEN_SECRET", "")

	if apiKey == "" {
		return fmt.Errorf("TWITTER_API_KEY environment variable is required")
	}
	if apiKeySecret == "" {
		return fmt.Errorf("TWITTER_API_KEY_SECRET environment variable is required")
	}
	if accessToken == "" {
		return fmt.Errorf("TWITTER_ACCESS_TOKEN environment variable is required")
	}
	if accessTokenSecret == "" {
		return fmt.Errorf("TWITTER_ACCESS_TOKEN_SECRET environment variable is required")
	}

	baseURL := GetBaseURL(cfg, "twitter")

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

	// Set Content-Type header
	req.Header.Set("Content-Type", "application/json")

	// Configure OAuth 1.0a
	oauthConfig := oauth1.NewConfig(apiKey, apiKeySecret)
	oauthToken := oauth1.NewToken(accessToken, accessTokenSecret)

	// Sign the request with OAuth 1.0a
	httpClient := oauthConfig.Client(context.Background(), oauthToken)

	// Send request
	resp, err := httpClient.Do(req)
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

// calculateTwitterLength calculates the effective length of text for Twitter's 280 char limit
// URLs count as 23 characters regardless of their actual length
func calculateTwitterLength(text string) int {
	baseLength := len(text)

	// Find URLs in text and adjust length
	// URLs are counted as 23 chars by Twitter, not their actual length
	urlPatterns := []string{"http://", "https://"}
	for _, pattern := range urlPatterns {
		idx := strings.Index(text, pattern)
		for idx != -1 {
			// Find end of URL (space, newline, or end of string)
			urlEnd := len(text)
			for i := idx; i < len(text); i++ {
				if text[i] == ' ' || text[i] == '\n' || text[i] == '\t' {
					urlEnd = i
					break
				}
			}
			actualURLLength := urlEnd - idx
			if actualURLLength > 23 {
				// Adjust: subtract actual length, add 23
				baseLength = baseLength - actualURLLength + 23
			}
			// Look for next URL after this one
			idx = strings.Index(text[urlEnd:], pattern)
			if idx != -1 {
				idx += urlEnd
			}
		}
	}
	return baseLength
}

// buildTwitterPostText constructs the text content for the Twitter post
// Twitter has a 280 character limit, so we need to be more concise
// URLs count as 23 characters regardless of their actual length
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
		// Account for URL being 23 chars, not full length
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

	// Add URL if available (URLs count as 23 chars in Twitter, not their actual length)
	if blogPost.URL != nil && *blogPost.URL != "" {
		parts = append(parts, *blogPost.URL)
	} else if baseURL != "" {
		// Construct URL from base URL and post ID
		url := BuildBlogPostURL(baseURL, blogPost.ID.String())
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

	// Join parts
	postText := strings.Join(parts, "\n\n")

	// Calculate effective length (accounting for URL being 23 chars)
	effectiveLength := calculateTwitterLength(postText)

	// If still too long, truncate more aggressively
	if effectiveLength > 280 {
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
			url = "\n\n" + BuildBlogPostURL(baseURL, blogPost.ID.String())
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
		// URL counts as 23 chars, not its actual length
		urlLength := 23 // Twitter counts URLs as 23 characters
		if url == "" {
			urlLength = 0
		}
		availableSpace := 280 - len(title) - urlLength - len(hashtags)
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

		// Final safety check - calculate effective length (URL = 23 chars)
		effectiveLength := calculateTwitterLength(postText)
		if effectiveLength > 280 {
			// Need to truncate more - reduce content
			excess := effectiveLength - 280
			contentLen := len(content)
			if contentLen > excess+3 {
				content = content[:contentLen-excess-3] + "..."
				postText = title + content + url + hashtags
			} else {
				// Last resort: hard truncate
				maxActualLength := 280 - urlLength + len(url) - 3
				if maxActualLength > 0 && len(postText) > maxActualLength {
					postText = postText[:maxActualLength] + "..."
				}
			}
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
