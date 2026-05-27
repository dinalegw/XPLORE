package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ValidateToken validates the Supabase JWT token and returns the user ID
func ValidateToken(tokenStr string) (string, error) {
	if tokenStr == "" || !strings.HasPrefix(tokenStr, "Bearer ") {
		return "", fmt.Errorf("missing or malformed token")
	}

	// Extract the actual token (remove "Bearer " prefix)
	token := strings.TrimPrefix(tokenStr, "Bearer ")
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	req, _ := http.NewRequest("GET", supabaseURL+"/auth/v1/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", serviceKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("supabase auth request failed: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("supabase returned status: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var user supabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil || user.ID == "" {
		return "", fmt.Errorf("failed to decode user: %w", err)
	}

	return user.ID, nil
}

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := ValidateToken(authHeader)
	if err != nil {
		log.Printf("Auth middleware error: %v", err)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or missing token"})
	}

	c.Locals("userID", userID)
	return c.Next()
}