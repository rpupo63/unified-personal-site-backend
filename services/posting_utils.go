package services

import "strings"

// FormatHashtag formats a tag value as a valid hashtag for social media platforms
// (Twitter, LinkedIn, etc.). It removes spaces and special characters, keeping only
// letters, numbers, and underscores. Hashtags cannot start with a number.
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

	formatted := result.String()
	// Ensure it doesn't start with a number (social media platform requirement)
	if len(formatted) > 0 && formatted[0] >= '0' && formatted[0] <= '9' {
		return ""
	}

	return formatted
}
