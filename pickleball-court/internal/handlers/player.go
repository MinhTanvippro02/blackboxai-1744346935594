package handlers

import (
	"database/sql"
	"net/http"
	"pickleball-court/internal/middleware"
	"pickleball-court/internal/models"
	"github.com/gin-gonic/gin"
	"time"
)

// PlayerDashboardHandler handles the player dashboard page
func PlayerDashboardHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify player role
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RolePlayer {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Get statistics
		var stats struct {
			UpcomingBookings  int
			HoursPlayed      float64
			TrainingSessions int
		}

		// Get upcoming bookings count
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM bookings 
			WHERE user_id = ? AND end_time > CURRENT_TIMESTAMP AND status != 'cancelled'
		`, user.ID).Scan(&stats.UpcomingBookings)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Calculate total hours played
		err = db.QueryRow(`
			SELECT COALESCE(SUM(
				ROUND(CAST(
					(JULIANDAY(end_time) - JULIANDAY(start_time)) * 24 
				AS REAL), 2)
			), 0)
			FROM bookings
			WHERE user_id = ? AND end_time <= CURRENT_TIMESTAMP AND status = 'confirmed'
		`, user.ID).Scan(&stats.HoursPlayed)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get training sessions count
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM training_session_participants tsp
			JOIN training_sessions ts ON ts.id = tsp.session_id
			WHERE tsp.user_id = ?
		`, user.ID).Scan(&stats.TrainingSessions)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get available courts with time slots
		courts, err := models.GetAvailableCourts(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load courts"})
			return
		}

		// Get user's bookings
		bookings, err := models.GetUserBookings(db, user.ID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load bookings"})
			return
		}

		// Get available training sessions
		trainingSessions, err := models.GetAvailableTrainingSessions(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load training sessions"})
			return
		}

		c.HTML(http.StatusOK, "player_dashboard.html", gin.H{
			"title": "Player Dashboard",
			"user":  user,
			"stats": stats,
			"courts": courts,
			"bookings": bookings,
			"trainingSessions": trainingSessions,
			"today": time.Now().Format("2006-01-02"),
		})
	}
}

// CreateBookingHandler handles court booking creation
func CreateBookingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RolePlayer {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var booking models.Booking
		if err := c.ShouldBindJSON(&booking); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		booking.UserID = user.ID
		booking.Status = models.BookingStatusPending
		booking.BookingType = models.BookingTypeRegular

		// Set end time to 1 hour after start time
		booking.EndTime = booking.StartTime.Add(time.Hour)

		// Validate time slot availability
		available, err := models.IsCourtAvailable(db, booking.CourtID, booking.StartTime, booking.EndTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check court availability"})
			return
		}
		if !available {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Court is not available for the selected time slot"})
			return
		}

		err = models.CreateBooking(db, &booking)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
			return
		}

		c.JSON(http.StatusOK, booking)
	}
}

// CancelBookingHandler handles booking cancellation
func CancelBookingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RolePlayer {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		bookingID := c.Param("id")

		// Verify the booking belongs to this user
		booking, err := models.GetBookingByID(db, bookingID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
			return
		}
		if booking.UserID != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		err = models.CancelBooking(db, bookingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel booking"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled successfully"})
	}
}

// EnrollTrainingHandler handles enrollment in training sessions
func EnrollTrainingHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RolePlayer {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sessionID := c.Param("id")

		// Check if already enrolled
		enrolled, err := models.IsUserEnrolled(db, user.ID, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check enrollment status"})
			return
		}
		if enrolled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Already enrolled in this session"})
			return
		}

		err = models.EnrollInTrainingSession(db, user.ID, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enroll in training session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully enrolled in training session"})
	}
}

// CancelTrainingEnrollmentHandler handles cancellation of training session enrollment
func CancelTrainingEnrollmentHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RolePlayer {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sessionID := c.Param("id")

		// Verify enrollment exists
		enrolled, err := models.IsUserEnrolled(db, user.ID, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check enrollment status"})
			return
		}
		if !enrolled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enrolled in this session"})
			return
		}

		err = models.CancelTrainingEnrollment(db, user.ID, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel enrollment"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Successfully cancelled enrollment"})
	}
}
