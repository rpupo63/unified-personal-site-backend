package services

import (
	"fmt"
	"strings"

	"github.com/rpupo63/unified-personal-site-backend/models"
	"github.com/rs/zerolog/log"
)

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// PostEverywhere posts a blog post to selected social media platforms
// It calls PostToSubstack, PostToMedium, PostToTwitter, and PostToLinkedIn
// based on the platforms specified in the platformsToPost parameter.
//
// Parameters:
//   - blogPost: The blog post to share
//   - tags: List of tags associated with the blog post
//   - mainImageURL: Optional URL of the main image for the post (required for Substack)
//   - platformsToPost: Slice of platform names to post to. Valid values: "substack", "medium", "twitter", "linkedin"
//     If empty or nil, no platforms will be posted to. Platform names are case-insensitive.
//
// Returns:
//   - error: Combined error message if any platform failed, nil if all succeeded
//     Individual platform errors are logged but the function continues to attempt
//     posting to all selected platforms even if some fail.
func PostEverywhere(blogPost models.BlogPost, tags []models.BlogTag, mainImageURL string, platformsToPost []string) error {
	var errors []string
	var successes []string

	// Post to Substack (requires mainImageURL)
	if contains(platformsToPost, "substack") {
		if mainImageURL != "" {
			log.Info().Msg("Posting to Substack...")
			if err := PostToSubstack(blogPost, tags, mainImageURL); err != nil {
				log.Error().Err(err).Msg("Failed to post to Substack")
				errors = append(errors, fmt.Sprintf("Substack: %v", err))
			} else {
				successes = append(successes, "Substack")
			}
		} else {
			log.Warn().Msg("Skipping Substack: mainImageURL is required but not provided")
			errors = append(errors, "Substack: mainImageURL is required but not provided")
		}
	}

	// Post to Medium
	if contains(platformsToPost, "medium") {
		log.Info().Msg("Posting to Medium...")
		if err := PostToMedium(blogPost, tags); err != nil {
			log.Error().Err(err).Msg("Failed to post to Medium")
			errors = append(errors, fmt.Sprintf("Medium: %v", err))
		} else {
			successes = append(successes, "Medium")
		}
	}

	// Post to Twitter
	if contains(platformsToPost, "twitter") {
		log.Info().Msg("Posting to Twitter...")
		if err := PostToTwitter(blogPost, tags); err != nil {
			log.Error().Err(err).Msg("Failed to post to Twitter")
			errors = append(errors, fmt.Sprintf("Twitter: %v", err))
		} else {
			successes = append(successes, "Twitter")
		}
	}

	// Post to LinkedIn
	if contains(platformsToPost, "linkedin") {
		log.Info().Msg("Posting to LinkedIn...")
		if err := PostToLinkedIn(blogPost, tags); err != nil {
			log.Error().Err(err).Msg("Failed to post to LinkedIn")
			errors = append(errors, fmt.Sprintf("LinkedIn: %v", err))
		} else {
			successes = append(successes, "LinkedIn")
		}
	}

	// Log summary
	if len(successes) > 0 {
		log.Info().Strs("platforms", successes).Msg("Successfully posted to platforms")
	}

	// Return combined error if any failures occurred
	if len(errors) > 0 {
		errorMsg := fmt.Sprintf("some platforms failed: %s", strings.Join(errors, "; "))
		log.Error().Msg(errorMsg)
		return fmt.Errorf("some platforms failed: %s", strings.Join(errors, "; "))
	}

	if len(platformsToPost) > 0 {
		log.Info().Msg("Successfully posted to all selected platforms")
	} else {
		log.Info().Msg("No platforms selected for posting")
	}
	return nil
}
