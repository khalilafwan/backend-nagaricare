package controllers

import (
	"backend-nagaricare/database"
	"backend-nagaricare/entity"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetAllPosts retrieves all posts from the database
func GetAllPosts(c *fiber.Ctx) error {
	// Query the database for all posts
	rows, err := database.DB.Query("SELECT id_posts, title, content, id_user, created_at FROM posts")
	if err != nil {
		log.Println("Error querying posts from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error querying posts"})
	}
	defer rows.Close()

	var posts []entity.Post
	for rows.Next() {
		var post entity.Post
		var createdAtStr string // Temporarily hold created_at as string

		// Scan the data into the Post struct fields, created_at goes into createdAtStr
		if err := rows.Scan(&post.ID_Posts, &post.Title, &post.Content, &post.ID_user, &createdAtStr); err != nil {
			log.Println("Error scanning post:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error scanning post"})
		}

		// Convert the string to time.Time
		post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			log.Println("Error parsing created_at:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error parsing created_at"})
		}

		posts = append(posts, post)
	}

	// Return the list of posts as JSON
	return c.JSON(posts)
}

// GetPostByID retrieves a specific post by its ID
func GetPostByID(c *fiber.Ctx) error {
	id := c.Params("id_post") // Post ID from URL parameters

	var post entity.Post
	var createdAtStr string // Hold created_at as a string

	// Query the database to get the post by its ID
	err := database.DB.QueryRow("SELECT id_posts, title, content, id_user, created_at FROM posts WHERE id_posts = ?", id).Scan(
		&post.ID_Posts, &post.Title, &post.Content, &post.ID_user, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
		}
		log.Println("Error querying post by ID:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Convert created_at to time.Time
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		log.Println("Error parsing created_at:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid created_at format"})
	}
	post.CreatedAt = createdAt

	// Return the post as JSON
	return c.JSON(post)
}

// GetPostByUserID retrieves posts for a specific user based on their ID_user
func GetPostByUserID(c *fiber.Ctx) error {
	ID_user := c.Params("id_user") // Retrieve user ID from URL parameters

	var posts []entity.Post
	rows, err := database.DB.Query("SELECT id_posts, title, content, id_user, created_at FROM posts WHERE id_user = ?", ID_user)
	if err != nil {
		log.Println("Error querying posts by user ID:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}
	defer rows.Close()

	for rows.Next() {
		var post entity.Post
		var createdAtStr string
		if err := rows.Scan(&post.ID_Posts, &post.Title, &post.Content, &post.ID_user, &createdAtStr); err != nil {
			log.Println("Error scanning post:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
		}

		// Parse the created_at string into a time.Time object
		post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			log.Println("Error parsing created_at:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Invalid created_at format"})
		}

		// Append the post to the posts slice
		posts = append(posts, post)
	}

	// Check if no posts were found
	if len(posts) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No posts found for this user"})
	}

	// Return the posts as JSON
	return c.JSON(posts)
}

// CreatePost inserts a new post into the database
func CreatePost(c *fiber.Ctx) error {
	var req entity.Post
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Insert new post into the database
	_, err := database.DB.Exec("INSERT INTO posts (title, content, created_at, id_user) VALUES (?, ?, NOW(), ?)", req.Title, req.Content, req.ID_user)
	if err != nil {
		log.Println("Error inserting post into database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create post"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Post created successfully"})
}

// UpdatePost updates an existing post
func UpdatePost(c *fiber.Ctx) error {
	id := c.Params("id_post")
	var req entity.Post
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if post exists
	var existingPost string
	err := database.DB.QueryRow("SELECT id_posts FROM posts WHERE id_posts = ?", id).Scan(&existingPost)
	if err == sql.ErrNoRows {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	} else if err != nil {
		log.Println("Error querying post from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Update post
	_, err = database.DB.Exec("UPDATE posts SET title = ?, content = ? WHERE id_posts = ?", req.Title, req.Content, id)
	if err != nil {
		log.Println("Error updating post in database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update post"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Post updated successfully"})
}

// DeletePost deletes a post from the database
func DeletePost(c *fiber.Ctx) error {
	id := c.Params("id_post")

	// Check if post exists
	var existingPost string
	err := database.DB.QueryRow("SELECT id_posts FROM posts WHERE id_posts = ?", id).Scan(&existingPost)
	if err == sql.ErrNoRows {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Post not found"})
	} else if err != nil {
		log.Println("Error querying post from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Delete post
	_, err = database.DB.Exec("DELETE FROM posts WHERE id_posts = ?", id)
	if err != nil {
		log.Println("Error deleting post from database:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete post"})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Post deleted successfully"})
}
