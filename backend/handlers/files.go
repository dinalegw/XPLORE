package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file provided"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot open file"})
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cannot read file"})
	}

	fileName := fmt.Sprintf("%s-%s", uuid.New().String(), file.Filename)
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceKey := os.Getenv("SUPABASE_SERVICE_KEY")
	uploadURL := fmt.Sprintf("%s/storage/v1/object/chat-files/%s", supabaseURL, fileName)

	req, _ := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	req.Header.Set("Authorization", "Bearer "+serviceKey)
	req.Header.Set("Content-Type", file.Header.Get("Content-Type"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode >= 400 {
		return c.Status(500).JSON(fiber.Map{"error": "Upload failed"})
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/chat-files/%s", supabaseURL, fileName)
	return c.JSON(fiber.Map{"url": publicURL})
}