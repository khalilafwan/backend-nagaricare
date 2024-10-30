package main

import (
	"backend-nagaricare/database"
	routes "backend-nagaricare/routers"
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Initialize Fiber app
	app := fiber.New()

	// Connect to the Database
	database.ConnectDB()

	// Setup Routes
	routes.SetupRoutes(app)

	// Start the server
	log.Fatal(app.Listen(":3000"))
}
