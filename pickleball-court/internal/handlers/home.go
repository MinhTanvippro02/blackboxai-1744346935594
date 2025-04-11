package handlers

import (
	"database/sql"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

// HomeHandler handles the home page
func HomeHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, _ := c.Get("user")
		
		// Get some basic stats for the home page
		var courtCount, userCount, bookingCount int
		db.QueryRow("SELECT COUNT(*) FROM courts").Scan(&courtCount)
		db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		db.QueryRow("SELECT COUNT(*) FROM bookings WHERE status != 'cancelled'").Scan(&bookingCount)

		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Welcome to PickleCourt",
			"user": user,
			"currentYear": time.Now().Year(),
			"stats": gin.H{
				"courts": courtCount,
				"users": userCount,
				"bookings": bookingCount,
			},
		})
	}
}
