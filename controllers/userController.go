package controllers

import (
	"backend-nagaricare/database"
	"backend-nagaricare/entity"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CreateUser handles user sign-up and inserts user details into the database
func CreateUser(c *fiber.Ctx) error {
	// Parse the request body into the User struct
	var req entity.User
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if the email is already registered
	var existingUser string
	err := database.DB.QueryRow("SELECT email FROM users WHERE id_user = ?", req.ID_user).Scan(&existingUser)
	if err == nil {
		// Email already exists
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email already registered"})
	} else if err != sql.ErrNoRows {
		// Database error
		log.Println("Error querying user from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Insert the new user into the database
	_, err = database.DB.Exec("INSERT INTO users (id_user, email, name, phone, profile_picture) VALUES (?, ?, ?, ?, ?)", req.ID_user, req.Email, req.Name, req.Phone, req.Picture)
	if err != nil {
		log.Println("Error inserting user into database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User created successfully"})
}

// GetUsers retrieves all users from the database
func GetUsers(c *fiber.Ctx) error {
	// Query the database for all users
	rows, err := database.DB.Query("SELECT id_user, email, name, phone, profile_picture FROM users")
	if err != nil {
		log.Println("Error querying users from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error querying users"})
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		// Scan each row into the User struct
		if err := rows.Scan(&user.ID_user, &user.Email, &user.Name, &user.Phone, &user.Picture); err != nil {
			log.Println("Error scanning user:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error scanning user"})
		}

		// // No need to reassign; just check if the phone is valid
		// if !user.Phone.Valid {
		// 	user.Phone.String = "" // If phone is NULL, set it to an empty string
		// }
		// if !user.Picture.Valid {
		// 	user.Picture.String = "" // If picture is NULL, set it to an empty string
		// }

		users = append(users, user)
	}

	// Return the list of users as JSON
	return c.JSON(users)
}

// GetUserDetails retrieves the details of the currently logged-in user
func GetUserDetails(c *fiber.Ctx) error {
	// Get the email from the query parameter
	ID_user := c.Params("id_user")
	if ID_user == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No user id provided"})
	}

	// Query the database for the user details
	var user entity.User
	err := database.DB.QueryRow("SELECT id_user, email, name, phone, profile_picture FROM users WHERE id_user = ?", ID_user).Scan(&user.ID_user, &user.Email, &user.Name, &user.Phone, &user.Picture)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		log.Println("Error querying user details from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Prepare response struct to handle NULL phone values
	response := struct {
		ID_user int     `json:"id_user"`
		Email   string  `json:"email"`
		Name    string  `json:"name"`
		Phone   *string `json:"phone"`
		Picture *string `json:"profile_picture"`
	}{
		ID_user: user.ID_user,
		Email:   user.Email,
		Name:    user.Name,
		Phone:   user.Phone,
		Picture: user.Picture,
	}

	// Return the user details as JSON
	return c.JSON(response)
}

// UserSignIn handles user sign-in via Google and inserts user details if not already present
func SignInGoogle(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	// Parse the request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if the user already exists in the database
	var existingUser string
	err := database.DB.QueryRow("SELECT email FROM users WHERE email = ?", req.Email).Scan(&existingUser)

	if err == sql.ErrNoRows {
		// If user doesn't exist, insert new user
		_, err := database.DB.Exec("INSERT INTO users (email, name) VALUES (?, ?)", req.Email, req.Name)
		if err != nil {
			log.Println("Error inserting new user into database:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create user"})
		}
	} else if err != nil {
		log.Println("Error querying user from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User signed in successfully"})
}

// SaveUserPhoto saves the user's photo and updates the profile_picture field in the database.
func SaveUserPhoto(c *fiber.Ctx) error {
	// Parse user ID from URL parameter
	ID_user := c.Params("id_user")
	if ID_user == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Check if the user exists and retrieve the current profile picture path
	db := database.DB
	var currentProfilePicturePath sql.NullString // Allows for null values
	err := db.QueryRow(`SELECT profile_picture FROM users WHERE id_user = ?`, ID_user).Scan(&currentProfilePicturePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database query failed",
		})
	}

	// Check if no file is uploaded to set profile_picture to NULL
	file, err := c.FormFile("profile_picture")
	if file == nil {
		// Set profile_picture to NULL in the database
		query := `UPDATE users SET profile_picture = NULL WHERE id_user = ?`
		_, err = db.Exec(query, ID_user)
		if err != nil {
			log.Printf("Failed to execute query: %s, error: %s", query, err.Error())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Database update failed",
			})
		}

		// Return success response indicating profile picture was set to NULL
		return c.JSON(fiber.Map{
			"message":         "Profile picture removed successfully",
			"profile_picture": nil,
		})
	} else if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File upload failed",
		})
	}

	// Generate a unique file name and save path
	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), fileExt)
	saveDir := "./userProfile/"
	filePath := filepath.Join(saveDir, fileName)

	// Create the directory if it doesn't exist
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		os.MkdirAll(saveDir, os.ModePerm)
	}

	// Save the new profile picture
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Delete the previous profile picture if it exists and is not the default image
	if currentProfilePicturePath.Valid && currentProfilePicturePath.String != "/default/profile_picture.png" {
		oldFilePath := filepath.Join(".", currentProfilePicturePath.String)
		if _, err := os.Stat(oldFilePath); err == nil {
			if err := os.Remove(oldFilePath); err != nil {
				log.Printf("Failed to delete old profile picture: %s", err)
			}
		}
	}

	// Define relative path to save in database
	relativePath := fmt.Sprintf("/userProfile/%s", fileName)

	// Update the profile picture path in the database
	query := `UPDATE users SET profile_picture = ? WHERE id_user = ?`
	_, err = db.Exec(query, relativePath, ID_user)
	if err != nil {
		log.Printf("Failed to execute query: %s, error: %s", query, err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database update failed",
		})
	}

	// Return success response with new profile picture path
	return c.JSON(fiber.Map{
		"message":         "Profile picture updated successfully",
		"profile_picture": relativePath,
	})
}

// GetUserPhoto retrieves the user's profile picture based on their ID
func GetUserPhoto(c *fiber.Ctx) error {
	// Parse the user ID from the URL parameter
	ID_user := c.Params("id_user")
	if ID_user == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Check if the user exists and retrieve the profile picture path
	db := database.DB
	var profilePicturePath string
	err := db.QueryRow(`SELECT profile_picture FROM users WHERE id_user = ?`, ID_user).Scan(&profilePicturePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database query failed",
		})
	}

	// If the profile picture path is empty, return null in the JSON response
	if profilePicturePath == "" {
		return c.JSON(fiber.Map{
			"profile_picture": nil,
		})
	}

	// Build the absolute path for the profile picture
	absolutePath := filepath.Join(".", profilePicturePath)

	// Open and read the profile picture file
	file, err := os.Open(absolutePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open profile picture",
		})
	}
	defer file.Close()

	// Read file content as bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read profile picture file",
		})
	}

	// Set the appropriate content type
	fileExt := strings.ToLower(filepath.Ext(absolutePath))
	var contentType string
	switch fileExt {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	default:
		contentType = "application/octet-stream" // Fallback for unknown types
	}

	// Set the response content type and send the file content
	c.Set("Content-Type", contentType)
	return c.Send(fileBytes)
}

// UpdateUser updates an existing user data
func UpdateUser(c *fiber.Ctx) error {
	ID_user := c.Params("id_user")
	var req entity.User
	log.Printf("Request data: %+v\n", req)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if the user exists
	var existingUser string
	err := database.DB.QueryRow("SELECT id_user FROM users WHERE id_user = ?", ID_user).Scan(&existingUser)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	} else if err != nil {
		log.Println("Error querying user from database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	// Set default values if any fields are nil
	if req.Phone == nil {
		req.Phone = new(string)
		*req.Phone = "" // Set a default empty string if needed
	}
	// if req.Picture == nil {
	// 	req.Picture = new(string)
	// 	*req.Picture = "" // Set a default empty string if needed
	// }

	// Update user
	_, err = database.DB.Exec(
		"UPDATE users SET email = ?, name = ?, phone = ? WHERE id_user = ?",
		req.Email, req.Name, req.Phone, ID_user,
	)
	if err != nil {
		log.Println("Error updating user in database:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update user"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "User updated successfully"})
}
