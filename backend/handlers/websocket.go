package handlers

import (
	"chat-app/db"
	"chat-app/models"
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Client struct {
	UserID string
	RoomID string
	Conn   *websocket.Conn
}

var (
	clients   = make(map[*Client]bool)
	clientsMu sync.Mutex
)

func broadcast(roomID string, event models.WSEvent) {
	data, _ := json.Marshal(event)
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if client.RoomID == roomID {
			client.Conn.WriteMessage(1, data)
		}
	}
}

func broadcastAll(event models.WSEvent) {
	data, _ := json.Marshal(event)
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		client.Conn.WriteMessage(1, data)
	}
}

func WSHandler(c *websocket.Conn) {
	userID := c.Query("user_id")
	roomID := c.Query("room_id")

	client := &Client{UserID: userID, RoomID: roomID, Conn: c}

	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	db.DB.Exec(`UPDATE profiles SET is_online = true WHERE id = $1`, userID)
	broadcastAll(models.WSEvent{Type: "presence.online", Payload: fiber.Map{"user_id": userID}})

	defer func() {
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
		db.DB.Exec(`UPDATE profiles SET is_online = false, last_seen = NOW() WHERE id = $1`, userID)
		broadcastAll(models.WSEvent{Type: "presence.offline", Payload: fiber.Map{"user_id": userID}})
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var event models.WSEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		switch event.Type {
		case "message.send":
			payload := event.Payload.(map[string]interface{})
			rID := payload["room_id"].(string)
			content, _ := payload["content"].(string)
			fileURL, _ := payload["file_url"].(string)

			var msgID string
			db.DB.QueryRow(`
				INSERT INTO messages (room_id, sender_id, content, file_url)
				VALUES ($1, $2, $3, NULLIF($4, '')) RETURNING id
			`, rID, userID, content, fileURL).Scan(&msgID)

			broadcast(rID, models.WSEvent{
				Type: "message.new",
				Payload: fiber.Map{
					"id": msgID, "room_id": rID,
					"sender_id": userID, "content": content,
					"file_url": fileURL,
				},
			})

		case "typing.start", "typing.stop":
			payload := event.Payload.(map[string]interface{})
			rID := payload["room_id"].(string)
			broadcast(rID, models.WSEvent{Type: event.Type, Payload: payload})

		case "receipt.read":
			payload := event.Payload.(map[string]interface{})
			msgID := payload["message_id"].(string)
			rID := payload["room_id"].(string)
			db.DB.Exec(`
				INSERT INTO read_receipts (message_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING
			`, msgID, userID)
			broadcast(rID, models.WSEvent{Type: "receipt.read", Payload: payload})
		}

		log.Printf("Event: %s from %s", event.Type, userID)
	}
}