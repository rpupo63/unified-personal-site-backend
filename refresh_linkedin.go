//go:build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🔄 LinkedIn Token Refresher")
	fmt.Println("---------------------------")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
		fmt.Println("Continuing with environment variables...")
	}

	// Get configuration from environment variables
	clientID := getEnv("LINKEDIN_CLIENT_ID", "")
	clientSecret := getEnv("LINKEDIN_CLIENT_SECRET", "")
	redirectURI := getEnv("LINKEDIN_REDIRECT_URI", "https://www.linkedin.com/developers/tools/oauth/redirect")
	envFilePath := getEnv("ENV_FILE_PATH", ".env")

	if clientID == "" {
		fmt.Println("❌ LINKEDIN_CLIENT_ID environment variable is required")
		fmt.Println("Please set it in your .env file or environment")
		os.Exit(1)
	}

	if clientSecret == "" {
		fmt.Println("❌ LINKEDIN_CLIENT_SECRET environment variable is required")
		fmt.Println("Please set it in your .env file or environment")
		os.Exit(1)
	}

	// 1. Generate and Print the Authorization URL
	authURL := fmt.Sprintf(
		"https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=%s&redirect_uri=%s&scope=openid%%20profile%%20w_member_social%%20email",
		clientID, url.QueryEscape(redirectURI),
	)

	fmt.Println("\n1. Click this link to authorize the app:")
	fmt.Printf("\n%s\n\n", authURL)

	// 2. Prompt for the Code
	fmt.Print("2. Paste the 'code' from the browser URL here: ")
	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	if code == "" {
		fmt.Println("❌ No code provided. Exiting.")
		os.Exit(1)
	}

	// 3. Exchange Code for Access Token
	fmt.Println("\n⏳ Exchanging code for access token...")
	accessToken, err := getAccessToken(code, clientID, clientSecret, redirectURI)
	if err != nil {
		fmt.Printf("❌ Error fetching token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Token received successfully!")

	// 4. Update the .env file
	fmt.Println("📝 Updating .env file...")
	if err := updateEnvFile(accessToken, envFilePath); err != nil {
		fmt.Printf("❌ Error updating .env: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🎉 Success! Your .env file has been updated with the new token.")
}

func getAccessToken(authCode, clientID, clientSecret, redirectURI string) (string, error) {
	apiURL := "https://www.linkedin.com/oauth/v2/accessToken"

	// Create URL-encoded data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", authCode)

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API responded with %s: %s", resp.Status, string(body))
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("response did not contain access_token")
	}

	return token, nil
}

func updateEnvFile(newToken, envFilePath string) error {
	// Read the file
	content, err := os.ReadFile(envFilePath)
	if err != nil {
		return fmt.Errorf("could not read %s: %v", envFilePath, err)
	}

	fileContent := string(content)
	lines := strings.Split(fileContent, "\n")
	found := false
	newLines := []string{}

	// Regex to find the variable (handles spaces around =)
	regex := regexp.MustCompile(`^LINKEDIN_ACCESS_TOKEN\s*=.*`)

	for _, line := range lines {
		if regex.MatchString(line) {
			// Replace the line with the new token
			newLines = append(newLines, fmt.Sprintf("LINKEDIN_ACCESS_TOKEN=%s", newToken))
			found = true
		} else {
			// Keep existing line
			newLines = append(newLines, line)
		}
	}

	// If the variable wasn't found, append it to the end
	if !found {
		newLines = append(newLines, fmt.Sprintf("LINKEDIN_ACCESS_TOKEN=%s", newToken))
	}

	// Write back to file
	output := strings.Join(newLines, "\n")
	if err := os.WriteFile(envFilePath, []byte(output), 0644); err != nil {
		return err
	}

	return nil
}

// getEnv returns the value of the environment variable key or a fallback value.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
