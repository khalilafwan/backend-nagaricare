package routes

import (
	"backend-nagaricare/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Forum routes
	forum := app.Group("/posts") // Create a group for forum posts

	forum.Post("/", controllers.CreatePost)                  // Create a new post
	forum.Get("/", controllers.GetAllPosts)                  // Get all posts
	forum.Get("/:id_post", controllers.GetPostByID)          // Get a specific post by ID
	forum.Get("/user/:id_user", controllers.GetPostByUserID) // Get all posts by a specific user (email)
	forum.Put("/:id_post", controllers.UpdatePost)           // Update a specific post by ID
	forum.Delete("/:id_post", controllers.DeletePost)        // Delete a post by ID

	// User routes
	user := app.Group("/users") // Create a group for user-related routes

	user.Post("/", controllers.CreateUser)                                // Create a new user
	user.Post("/signin", controllers.SignInGoogle)                        // Google Sign-In
	user.Get("/", controllers.GetUsers)                                   // Get all users
	user.Get("/:id_user", controllers.GetUserDetails)                     // Get user details by email
	user.Put("/uploadprofilepicture/:id_user", controllers.SaveUserPhoto) // Save user profile picture
}
