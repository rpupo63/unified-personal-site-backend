package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rpupo63/unified-personal-site-backend/config"
	"github.com/rpupo63/unified-personal-site-backend/models"
	"github.com/rs/zerolog/log"
)

// SubstackPostResponse represents the response from Substack API
type SubstackPostResponse struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	DraftId     int64  `json:"draft_id"`
	Publication string `json:"publication_id"`
}

// SubstackErrorResponse handles error messages from the private API
type SubstackErrorResponse struct {
	Errors []string `json:"errors"`
	Msg    string   `json:"msg"` // Sometimes they use 'msg' instead of 'errors'
}

// PostToSubstack posts a blog post to Substack
// Requires environment variables in .env:
//   - SUBSTACK_COOKIE: The 'connect.sid' cookie value from your browser session
//   - SUBSTACK_DOMAIN: Your subdomain (e.g., "betopupo" for betopupo.substack.com)
//
// Optional environment variables:
//   - BASE_URL: Optional unified base URL for constructing blog post links (defaults to empty if not set)
func PostToSubstack(blogPost models.BlogPost, tags []models.BlogTag, mainImageURL string) error {
	// 1. Load Configuration (Copying logic from your LinkedIn function)
	possiblePaths := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("backend", ".env"),
	}

	var envLoaded bool
	for _, envPath := range possiblePaths {
		if err := godotenv.Load(envPath); err == nil {
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Warn().Msg("Failed to load .env file, using existing environment variables")
	}

	cfg := config.New()

	// 2. Get Substack Specific Credentials
	cookie := config.GetString(cfg, "SUBSTACK_COOKIE", "")
	if cookie == "" {
		return fmt.Errorf("SUBSTACK_COOKIE environment variable is required (connect.sid)")
	}

	subdomain := config.GetString(cfg, "SUBSTACK_DOMAIN", "")
	if subdomain == "" {
		return fmt.Errorf("SUBSTACK_DOMAIN environment variable is required")
	}

	baseURL := GetBaseURL(cfg, "")

	// 3. Construct the HTML Body
	// Substack expects HTML. We combine the Image, Content, Tags, and Link.
	htmlBody := buildSubstackHtml(blogPost, tags, mainImageURL, baseURL)

	// 4. Build Payload
	payload := map[string]interface{}{
		"title":    blogPost.Title,
		"body":     htmlBody,
		"audience": "public", // or "only_paid", "everyone"
		"type":     "newsletter",
		"draft":    false, // Set to true if you only want to save a draft
	}

	// Add tags to payload if available (Substack API may support this)
	// Note: This is an unofficial API, so tags field may or may not be supported
	if len(tags) > 0 {
		var tagList []string
		for _, tag := range tags {
			// Substack tags are typically lowercase and can contain spaces/hyphens
			// Remove "#" if present, keep the tag value as-is (Substack handles formatting)
			tagValue := strings.TrimSpace(tag.Value)
			tagValue = strings.TrimPrefix(tagValue, "#")
			tagValue = strings.TrimSpace(tagValue)
			if tagValue != "" {
				tagList = append(tagList, tagValue)
			}
		}
		if len(tagList) > 0 {
			payload["tags"] = tagList
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Substack payload: %w", err)
	}

	// 5. Create Request
	// Note: This endpoint is reverse-engineered and unofficial
	url := fmt.Sprintf("https://%s.substack.com/api/v1/posts", subdomain)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Substack request: %w", err)
	}

	// 6. Set Headers (Crucial for bypassing bot detection)
	// The Cookie is the authentication mechanism here.
	req.Header.Set("Cookie", fmt.Sprintf("connect.sid=%s", cookie))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 7. Send Request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Substack: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Substack response: %w", err)
	}

	// 8. Handle Response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("substack API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var postResp SubstackPostResponse
	if err := json.Unmarshal(bodyBytes, &postResp); err != nil {
		// If unmarshal fails but status was 200, we assume success but warn
		log.Warn().Err(err).Msg("Post likely succeeded, but failed to parse response JSON")
	} else {
		log.Info().Int64("substackId", postResp.ID).Msg("Successfully posted to Substack")
	}

	return nil
}

// buildSubstackHtml converts the raw text content into simple HTML
// and embeds the main image at the top if provided. Also includes tags as hashtags.
func buildSubstackHtml(blogPost models.BlogPost, tags []models.BlogTag, imageURL, baseURL string) string {
	var sb strings.Builder

	// 1. Embed Main Image
	if imageURL != "" {
		// Substack figures out layout based on simple HTML tags
		sb.WriteString(fmt.Sprintf(`<figure><img src="%s"><figcaption></figcaption></figure><br>`, imageURL))
	}

	// 2. Add Content
	// Assuming blogPost.Content is plain text. We wrap paragraphs.
	// If blogPost.Content is already HTML, you can append it directly.
	content := blogPost.Content
	if !strings.Contains(content, "<p>") {
		// rudimentary plain-text to HTML conversion
		paragraphs := strings.Split(content, "\n\n")
		for _, p := range paragraphs {
			if strings.TrimSpace(p) != "" {
				sb.WriteString(fmt.Sprintf("<p>%s</p>", p))
			}
		}
	} else {
		sb.WriteString(content)
	}

	// 3. Add Tags as Hashtags (if available)
	// Include tags in the HTML body so they're visible even if the API doesn't support the tags field
	if len(tags) > 0 {
		var hashtags []string
		for _, tag := range tags {
			// Format tag as hashtag for display in content
			hashtag := FormatHashtag(tag.Value)
			if hashtag != "" {
				hashtags = append(hashtags, "#"+hashtag)
			}
		}
		if len(hashtags) > 0 {
			sb.WriteString(fmt.Sprintf(`<p>%s</p>`, strings.Join(hashtags, " ")))
		}
	}

	// 4. Add Footer / Original Link
	var postURL string
	if blogPost.URL != nil && *blogPost.URL != "" {
		postURL = *blogPost.URL
	} else if baseURL != "" {
		postURL = BuildBlogPostURL(baseURL, blogPost.ID.String())
	}
	if postURL != "" {
		sb.WriteString(fmt.Sprintf(`<p><i>Originally published at <a href="%s">%s</a></i></p>`, postURL, postURL))
	}

	return sb.String()
}
