package middleware

import (
	"database/sql"
	"net/http"
	"pickleball-court/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
)

const (
	UserKey = "user"
)

// AuthRequired ensures the user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(UserKey)
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RoleRequired ensures the user has the required role
func RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(UserKey)
		if userID == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Get user from context (set by AuthRequired middleware)
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		// Check if user's role is in allowed roles
		currentUser := user.(*models.User)
		allowed := false
		for _, role := range allowedRoles {
			if currentUser.Role == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoadUser middleware loads the user from the session and adds it to the context
func LoadUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get(UserKey)
		if userID == nil {
			c.Next()
			return
		}

		user, err := models.GetUserByID(db, userID.(int64))
		if err != nil {
			session.Clear()
			session.Save()
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}
	return user.(*models.User).Role == models.RoleAdmin
}

// IsCoach checks if the current user is a coach
func IsCoach(c *gin.Context) bool {
	user, exists := c.Get("user")
	if !exists {
		return false
	}
	return user.(*models.User).Role == models.RoleCoach
}

// GetCurrentUser returns the current logged-in user
func GetCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}

// SetUserSession sets the user session after successful login
func SetUserSession(c *gin.Context, userID int64) error {
	session := sessions.Default(c)
	session.Set(UserKey, userID)
	return session.Save()
}

// ClearUserSession clears the user session on logout
func ClearUserSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}
