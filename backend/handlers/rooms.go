package handlers

import (
	"chat-app/db"
	"chat-app/models"

	"github.com/gofiber/fiber/v2"
)

func GetRooms(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	rows, err := db.DB.Query(`
		SELECT r.id, r.name, r.is_private, r.created_at
		FROM rooms r
		JOIN room_members rm ON rm.room_id = r.id
		WHERE rm.user_id = $1
	`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		rows.Scan(&room.ID, &room.Name, &room.IsPrivate, &room.CreatedAt)
		rooms = append(rooms, room)
	}
	return c.JSON(rooms)
}

func BrowseRooms(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	rows, err := db.DB.Query(`
		SELECT r.id, r.name, r.is_private, r.created_at,
		       COALESCE(rm.user_id::text, '') as is_member,
		       COALESCE(jr.status, '') as request_status
		FROM rooms r
		LEFT JOIN room_members rm ON rm.room_id = r.id AND rm.user_id = $1
		LEFT JOIN join_requests jr ON jr.room_id = r.id AND jr.user_id = $1
		WHERE r.is_private = false
		ORDER BY r.created_at DESC
	`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type BrowseRoom struct {
		models.Room
		IsMember      bool   `json:"is_member"`
		RequestStatus string `json:"request_status"`
	}

	var rooms []BrowseRoom
	for rows.Next() {
		var room BrowseRoom
		var isMemberStr string
		rows.Scan(&room.ID, &room.Name, &room.IsPrivate, &room.CreatedAt, &isMemberStr, &room.RequestStatus)
		room.IsMember = isMemberStr != ""
		rooms = append(rooms, room)
	}
	return c.JSON(rooms)
}

func CreateRoom(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	body := struct {
		Name      string   `json:"name"`
		IsPrivate bool     `json:"is_private"`
		MemberIDs []string `json:"member_ids"`
	}{}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	var roomID string
	err := db.DB.QueryRow(`
		INSERT INTO rooms (name, is_private, created_by) VALUES ($1, $2, $3) RETURNING id
	`, body.Name, body.IsPrivate, userID).Scan(&roomID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	db.DB.Exec(`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`, roomID, userID)

	for _, memberID := range body.MemberIDs {
		db.DB.Exec(`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`, roomID, memberID)
	}

	return c.JSON(fiber.Map{"id": roomID})
}

func JoinRoom(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	roomID := c.Params("id")

	_, err := db.DB.Exec(`
		INSERT INTO room_members (room_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING
	`, roomID, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func RequestJoin(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	roomID := c.Params("id")

	// Check if already a member
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM room_members WHERE room_id = $1 AND user_id = $2`, roomID, userID).Scan(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Already a member"})
	}

	_, err := db.DB.Exec(`
		INSERT INTO join_requests (room_id, user_id, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (room_id, user_id) DO UPDATE SET status = 'pending'
	`, roomID, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true})
}

func GetJoinRequests(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	rows, err := db.DB.Query(`
		SELECT jr.id, jr.room_id, jr.user_id, jr.status, jr.created_at,
		       p.username, p.avatar_url, r.name as room_name
		FROM join_requests jr
		JOIN profiles p ON p.id = jr.user_id
		JOIN rooms r ON r.id = jr.room_id
		WHERE r.created_by = $1 AND jr.status = 'pending'
		ORDER BY jr.created_at DESC
	`, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type JoinRequest struct {
		ID        string `json:"id"`
		RoomID    string `json:"room_id"`
		UserID    string `json:"user_id"`
		Status    string `json:"status"`
		Username  string `json:"username"`
		AvatarURL string `json:"avatar_url"`
		RoomName  string `json:"room_name"`
	}

	var requests []JoinRequest
	for rows.Next() {
		var req JoinRequest
		var createdAt interface{}
		rows.Scan(&req.ID, &req.RoomID, &req.UserID, &req.Status, &createdAt,
			&req.Username, &req.AvatarURL, &req.RoomName)
		requests = append(requests, req)
	}
	return c.JSON(requests)
}

func RespondJoinRequest(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	requestID := c.Params("id")

	body := struct {
		Approve bool `json:"approve"`
	}{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	// Verify the requester owns the room
	var roomID, requesterID string
	err := db.DB.QueryRow(`
		SELECT jr.room_id, jr.user_id FROM join_requests jr
		JOIN rooms r ON r.id = jr.room_id
		WHERE jr.id = $1 AND r.created_by = $2
	`, requestID, userID).Scan(&roomID, &requesterID)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Not authorized"})
	}

	status := "rejected"
	if body.Approve {
		status = "approved"
		db.DB.Exec(`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, roomID, requesterID)
	}

	db.DB.Exec(`UPDATE join_requests SET status = $1 WHERE id = $2`, status, requestID)

	return c.JSON(fiber.Map{"success": true, "status": status})
}

func SearchUsers(c *fiber.Ctx) error {
	myID := c.Locals("userID").(string)
	query := c.Query("q")
	if query == "" {
		return c.JSON([]fiber.Map{})
	}

	rows, err := db.DB.Query(`
		SELECT p.id, p.username, p.avatar_url, p.is_online,
		       COALESCE(fr.status, '') as friend_status
		FROM profiles p
		LEFT JOIN friend_requests fr ON (
			(fr.sender_id = $1 AND fr.receiver_id = p.id) OR
			(fr.receiver_id = $1 AND fr.sender_id = p.id)
		)
		WHERE p.username ILIKE $2 AND p.id != $1
		LIMIT 10
	`, myID, "%"+query+"%")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type UserResult struct {
		ID           string `json:"id"`
		Username     string `json:"username"`
		AvatarURL    string `json:"avatar_url"`
		IsOnline     bool   `json:"is_online"`
		FriendStatus string `json:"friend_status"`
	}

	var users []UserResult
	for rows.Next() {
		var u UserResult
		rows.Scan(&u.ID, &u.Username, &u.AvatarURL, &u.IsOnline, &u.FriendStatus)
		users = append(users, u)
	}
	return c.JSON(users)
}

func SendFriendRequest(c *fiber.Ctx) error {
	senderID := c.Locals("userID").(string)
	body := struct {
		ReceiverID string `json:"receiver_id"`
	}{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	_, err := db.DB.Exec(`
		INSERT INTO friend_requests (sender_id, receiver_id, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (sender_id, receiver_id) DO UPDATE SET status = 'pending'
	`, senderID, body.ReceiverID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func GetFriendRequests(c *fiber.Ctx) error {
	myID := c.Locals("userID").(string)

	rows, err := db.DB.Query(`
		SELECT fr.id, fr.sender_id, fr.status, p.username, p.avatar_url
		FROM friend_requests fr
		JOIN profiles p ON p.id = fr.sender_id
		WHERE fr.receiver_id = $1 AND fr.status = 'pending'
		ORDER BY fr.created_at DESC
	`, myID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	type FriendRequest struct {
		ID       string `json:"id"`
		SenderID string `json:"sender_id"`
		Username string `json:"username"`
		AvatarURL string `json:"avatar_url"`
		Status   string `json:"status"`
	}

	var requests []FriendRequest
	for rows.Next() {
		var r FriendRequest
		rows.Scan(&r.ID, &r.SenderID, &r.Status, &r.Username, &r.AvatarURL)
		requests = append(requests, r)
	}
	return c.JSON(requests)
}

func RespondFriendRequest(c *fiber.Ctx) error {
	myID := c.Locals("userID").(string)
	requestID := c.Params("id")

	body := struct {
		Approve bool `json:"approve"`
	}{}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	var senderID string
	err := db.DB.QueryRow(`
		SELECT sender_id FROM friend_requests
		WHERE id = $1 AND receiver_id = $2
	`, requestID, myID).Scan(&senderID)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{"error": "Not authorized"})
	}

	if body.Approve {
		// Create a private room between the two users
		var roomID string
		db.DB.QueryRow(`
			INSERT INTO rooms (name, is_private, created_by)
			VALUES ('', true, $1) RETURNING id
		`, myID).Scan(&roomID)

		db.DB.Exec(`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`, roomID, myID)
		db.DB.Exec(`INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`, roomID, senderID)
		db.DB.Exec(`UPDATE friend_requests SET status = 'accepted' WHERE id = $1`, requestID)

		return c.JSON(fiber.Map{"success": true, "room_id": roomID})
	}

	db.DB.Exec(`UPDATE friend_requests SET status = 'rejected' WHERE id = $1`, requestID)
	return c.JSON(fiber.Map{"success": true})
}