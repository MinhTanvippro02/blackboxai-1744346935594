package handlers

import (
	"database/sql"
	"net/http"
	"pickleball-court/internal/middleware"
	"pickleball-court/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
)

// ShowLoginHandler displays the login page
func ShowLoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
		})
	}
}

// LoginHandler processes the login form
func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		user, err := models.AuthenticateUser(db, username, password)
		if err != nil {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{
				"title": "Login",
				"error": "Invalid username or password",
			})
			return
		}

		// Set user session
		session := sessions.Default(c)
		session.Set(middleware.UserKey, user.ID)
		session.Save()

		// Redirect based on user role
		switch user.Role {
		case models.RoleAdmin:
			c.Redirect(http.StatusFound, "/admin/dashboard")
		case models.RoleCoach:
			c.Redirect(http.StatusFound, "/coach/dashboard")
		default:
			c.Redirect(http.StatusFound, "/player/dashboard")
		}
	}
}

// ShowRegisterHandler displays the registration page
func ShowRegisterHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"title": "Register",
		})
	}
}

// RegisterHandler processes the registration form
func RegisterHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		email := c.PostForm("email")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")
		role := c.PostForm("role")

		// Validate passwords match
		if password != confirmPassword {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{
				"title": "Register",
				"error": "Passwords do not match",
			})
			return
		}

		// Validate role
		if role != models.RolePlayer && role != models.RoleCoach {
			role = models.RolePlayer // Default to player if invalid role
		}

		// Create new user
		user := &models.User{
			Username: username,
			Password: password,
			Email:    email,
			Role:     role,
		}

		err := models.CreateUser(db, user)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "register.html", gin.H{
				"title": "Register",
				"error": "Failed to create account. Username or email may already be in use.",
			})
			return
		}

		// Set user session
		session := sessions.Default(c)
		session.Set(middleware.UserKey, user.ID)
		session.Save()

		// Redirect based on user role
		switch user.Role {
		case models.RoleCoach:
			c.Redirect(http.StatusFound, "/coach/dashboard")
		default:
			c.Redirect(http.StatusFound, "/player/dashboard")
		}
	}
}

// LogoutHandler handles user logout
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.Redirect(http.StatusFound, "/")
	}
}


// ProfileHandler displays the user profile page
func ProfileHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Get user's booking history
		bookings, err := models.GetUserBookings(db, user.ID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error": "Failed to load booking history",
			})
			return
		}

		c.HTML(http.StatusOK, "profile.html", gin.H{
			"title": "My Profile",
			"user": user,
			"bookings": bookings,
		})
	}
}

// UpdateProfileHandler handles profile updates
func UpdateProfileHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Update user fields
		user.Username = c.PostForm("username")
		user.Email = c.PostForm("email")

		err := models.UpdateUser(db, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
	}
}

// UpdatePasswordHandler handles password updates
func UpdatePasswordHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		currentPassword := c.PostForm("current_password")
		newPassword := c.PostForm("new_password")
		confirmPassword := c.PostForm("confirm_password")

		// Verify current password
		_, err := models.AuthenticateUser(db, user.Username, currentPassword)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
			return
		}

		// Validate new password
		if newPassword != confirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New passwords do not match"})
			return
		}

		// Update password
		err = models.UpdatePassword(db, user.ID, newPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
	}
}
