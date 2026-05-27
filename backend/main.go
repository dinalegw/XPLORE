package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"

	"chat-app/db"
	"chat-app/handlers"
)

func main() {
	godotenv.Load()

	db.Connect()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// WebSocket upgrade middleware
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket route
	app.Get("/ws", fiberws.New(handlers.WSHandler))

	// Protected API routes
	api := app.Group("/api", handlers.AuthMiddleware)
	api.Get("/rooms", handlers.GetRooms)
	api.Post("/rooms", handlers.CreateRoom)
	api.Get("/rooms/browse", handlers.BrowseRooms)
	api.Get("/rooms/requests", handlers.GetJoinRequests)
	api.Post("/rooms/:id/request-join", handlers.RequestJoin)
	api.Post("/rooms/requests/:id/respond", handlers.RespondJoinRequest)
	api.Post("/rooms/:id/join", handlers.JoinRoom)
	api.Get("/rooms/:id/messages", handlers.GetMessages)
	api.Post("/upload", handlers.UploadFile)
	api.Get("/users/search", handlers.SearchUsers)
	api.Post("/users/friend-request", handlers.SendFriendRequest)
	api.Get("/users/friend-requests", handlers.GetFriendRequests)
	api.Post("/users/friend-requests/:id/respond", handlers.RespondFriendRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
