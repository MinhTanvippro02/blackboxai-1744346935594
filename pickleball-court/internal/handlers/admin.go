package handlers

import (
	"database/sql"
	"net/http"
	"pickleball-court/internal/middleware"
	"pickleball-court/internal/models"
	"time"
	"github.com/gin-gonic/gin"
)

// AdminDashboardHandler handles the admin dashboard page
func AdminDashboardHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify admin role
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleAdmin {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Get statistics
		var stats struct {
			Courts          int
			Users          int
			Bookings       int
			TrainingSessions int
		}

		// Get court count
		err := db.QueryRow("SELECT COUNT(*) FROM courts").Scan(&stats.Courts)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get user count
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.Users)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get active bookings count
		err = db.QueryRow("SELECT COUNT(*) FROM bookings WHERE status != 'cancelled'").Scan(&stats.Bookings)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get training sessions count
		err = db.QueryRow("SELECT COUNT(*) FROM training_sessions").Scan(&stats.TrainingSessions)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get all courts
		courts, err := models.GetAllCourts(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load courts"})
			return
		}

		// Get all users
		users, err := models.GetAllUsers(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load users"})
			return
		}

		// Get recent bookings
		bookings, err := models.GetRecentBookings(db, 10) // Get last 10 bookings
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load bookings"})
			return
		}

		c.HTML(http.StatusOK, "admin_dashboard.html", gin.H{
			"title": "Admin Dashboard",
			"user":  user,
			"stats": stats,
			"courts": courts,
			"users": users,
			"bookings": bookings,
		})
	}
}

// CreateCourtHandler handles court creation
func CreateCourtHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var court models.Court
		if err := c.ShouldBindJSON(&court); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := models.CreateCourt(db, &court)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create court"})
			return
		}

		c.JSON(http.StatusOK, court)
	}
}

// UpdateCourtHandler handles court updates
func UpdateCourtHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var court models.Court
		if err := c.ShouldBindJSON(&court); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := models.UpdateCourt(db, &court)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update court"})
			return
		}

		c.JSON(http.StatusOK, court)
	}
}

// DeleteCourtHandler handles court deletion
func DeleteCourtHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		courtID := c.Param("id")
		err := models.DeleteCourt(db, courtID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete court"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Court deleted successfully"})
	}
}

// CreateUserHandler handles user creation by admin
func CreateUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := models.CreateUser(db, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// UpdateUserHandler handles user updates by admin
func UpdateUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := models.UpdateUser(db, &user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// ListUsersHandler handles listing all users
func ListUsersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify admin role
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		users, err := models.GetAllUsers(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load users"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

// ListCourtsHandler handles listing all courts
func ListCourtsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify admin role
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleAdmin {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		courts, err := models.GetAllCourts(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load courts"})
			return
		}

		c.JSON(http.StatusOK, courts)
	}
}

// ListTrainingSessionsHandler handles listing all training sessions
func ListTrainingSessionsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify admin or coach role
		user := middleware.GetCurrentUser(c)
		if user == nil || (user.Role != models.RoleAdmin && user.Role != models.RoleCoach) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var sessions []*models.TrainingSession
		var err error

		if user.Role == models.RoleCoach {
			sessions, err = models.GetTrainingSessionsByCoach(db, user.ID)
		} else {
			// For admin, get all sessions
			sessions, err = models.GetAvailableTrainingSessions(db)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load training sessions"})
			return
		}

		c.JSON(http.StatusOK, sessions)
	}
}

// GetCourtAvailabilityHandler handles retrieving court availability
func GetCourtAvailabilityHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Query("date")
		if date == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required"})
			return
		}

		// Parse the date
		startDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		// Get all courts
		courts, err := models.GetAllCourts(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load courts"})
			return
		}

		// For each court, get its availability for the day
		type TimeSlot struct {
			StartTime     time.Time `json:"start_time"`
			Available     bool      `json:"available"`
			FormattedTime string    `json:"formatted_time"`
		}

		type CourtAvailability struct {
			ID          int64      `json:"id"`
			Name        string     `json:"name"`
			Description string     `json:"description"`
			TimeSlots   []TimeSlot `json:"time_slots"`
		}

		var availability []CourtAvailability

		// Operating hours (6 AM to 10 PM)
		openingHour := 6
		closingHour := 22

		for _, court := range courts {
			courtAvail := CourtAvailability{
				ID:          court.ID,
				Name:        court.Name,
				Description: court.Description,
			}

			// Generate time slots for the day
			for hour := openingHour; hour < closingHour; hour++ {
				slotStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), hour, 0, 0, 0, time.Local)
				slotEnd := slotStart.Add(time.Hour)

				// Skip if the slot is in the past
				if slotStart.Before(time.Now()) {
					continue
				}

				// Check availability
				available, err := models.IsCourtAvailable(db, court.ID, slotStart, slotEnd)
				if err != nil {
					continue
				}

				timeSlot := TimeSlot{
					StartTime:     slotStart,
					Available:     available,
					FormattedTime: slotStart.Format("15:04"),
				}

				courtAvail.TimeSlots = append(courtAvail.TimeSlots, timeSlot)
			}

			availability = append(availability, courtAvail)
		}

		c.JSON(http.StatusOK, availability)
	}
}

// DeleteUserHandler handles user deletion by admin
func DeleteUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		err := models.DeleteUser(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

// UpdateBookingHandler handles booking updates by admin
func UpdateBookingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookingID := c.Param("id")
		var status struct {
			Status string `json:"status"`
		}
		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := models.UpdateBookingStatus(db, bookingID, status.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Booking updated successfully"})
	}
}

// ListAllBookingsHandler handles listing all bookings for admin
func ListAllBookingsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		bookings, err := models.GetAllBookings(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load bookings"})
			return
		}

		c.JSON(http.StatusOK, bookings)
	}
}
