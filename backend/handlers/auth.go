package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type supabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(fiber.Map{"error": "Missing token"})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")

	req, _ := http.NewRequest("GET", supabaseURL+"/auth/v1/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	req.Header.Set("apikey", serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Supabase auth request failed: %v", err)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}
	if resp.StatusCode != 200 {
		log.Printf("Supabase returned status: %d for URL: %s", resp.StatusCode, supabaseURL)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token"})
	}
	defer resp.Body.Close()

	var user supabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil || user.ID == "" {
		log.Printf("Failed to decode user: %v", err)
		return c.Status(401).JSON(fiber.Map{"error": "Invalid user"})
	}

	c.Locals("userID", user.ID)
	return c.Next()
}