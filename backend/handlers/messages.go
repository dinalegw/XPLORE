package handlers

import (
	"chat-app/db"
	"chat-app/models"

	"github.com/gofiber/fiber/v2"
)

func GetMessages(c *fiber.Ctx) error {
	roomID := c.Params("id")

	rows, err := db.DB.Query(`
		SELECT m.id, m.room_id, m.sender_id, m.content, COALESCE(m.file_url, ''), m.created_at,
		       p.username, p.avatar_url
		FROM messages m
		JOIN profiles p ON p.id = m.sender_id
		WHERE m.room_id = $1
		ORDER BY m.created_at ASC
		LIMIT 100
	`, roomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var sender models.User
		rows.Scan(&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.FileURL,
			&msg.CreatedAt, &sender.Username, &sender.AvatarURL)
		msg.Sender = &sender
		messages = append(messages, msg)
	}
	return c.JSON(messages)
}