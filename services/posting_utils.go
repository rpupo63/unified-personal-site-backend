package services

import (
	"fmt"
	"strings"

	"github.com/rpupo63/unified-personal-site-backend/config"
)

// FormatHashtag formats a tag value as a valid hashtag for social media platforms
// (Twitter, LinkedIn, etc.). It removes spaces and special characters, keeping only
// letters, numbers, and underscores. Hashtags cannot start with a number.
// Returns lowercase for consistency (best practice, though not required).
func FormatHashtag(tag string) string {
	// Remove leading/trailing whitespace
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return ""
	}

	// Replace spaces and special characters with nothing or underscores
	// Hashtags can contain letters, numbers, and underscores
	var result strings.Builder
	for _, r := range tag {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			// Convert spaces and hyphens to nothing (or keep underscores)
			if r == '_' {
				result.WriteRune(r)
			}
		}
	}

	formatted := strings.ToLower(result.String()) // Convert to lowercase for consistency
	// Ensure it doesn't start with a number (social media platform requirement)
	if len(formatted) > 0 && formatted[0] >= '0' && formatted[0] <= '9' {
		return ""
	}

	return formatted
}

// GetBaseURL retrieves the base URL from configuration
// It first checks for a unified BASE_URL environment variable,
// then falls back to platform-specific variables for backward compatibility:
//   - TWITTER_BASE_URL for Twitter
//   - LINKEDIN_BASE_URL for LinkedIn
//
// Parameters:
//   - cfg: Configuration map from config.New()
//   - platform: Optional platform name ("twitter", "linkedin") for fallback
//
// Returns:
//   - The base URL string, or empty string if not found
func GetBaseURL(cfg map[string]string, platform string) string {
	// First, try unified BASE_URL
	baseURL := config.GetString(cfg, "BASE_URL", "")
	if baseURL != "" {
		return baseURL
	}

	// Fall back to platform-specific variables for backward compatibility
	switch strings.ToLower(platform) {
	case "twitter":
		return config.GetString(cfg, "TWITTER_BASE_URL", "")
	case "linkedin":
		return config.GetString(cfg, "LINKEDIN_BASE_URL", "")
	default:
		// If no platform specified, try common fallbacks
		if baseURL := config.GetString(cfg, "TWITTER_BASE_URL", ""); baseURL != "" {
			return baseURL
		}
		return config.GetString(cfg, "LINKEDIN_BASE_URL", "")
	}
}

// BuildBlogPostURL constructs a blog post URL from base URL and post ID
// Parameters:
//   - baseURL: The base URL (e.g., "https://example.com")
//   - postID: The blog post ID (UUID string)
//
// Returns:
//   - The full blog post URL (e.g., "https://example.com/blog/{postID}")
func BuildBlogPostURL(baseURL, postID string) string {
	if baseURL == "" || postID == "" {
		return ""
	}
	return fmt.Sprintf("%s/blog/%s", strings.TrimSuffix(baseURL, "/"), postID)
}
