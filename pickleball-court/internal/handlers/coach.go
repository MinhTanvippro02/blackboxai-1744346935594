package handlers

import (
	"database/sql"
	"net/http"
	"pickleball-court/internal/middleware"
	"pickleball-court/internal/models"
	"github.com/gin-gonic/gin"
	"time"
)

// CoachDashboardHandler handles the coach dashboard page
func CoachDashboardHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify coach role
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleCoach {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		// Get statistics
		var stats struct {
			ActiveSessions  int
			TotalStudents   int
			HoursTaught     float64
		}

		// Get active sessions count
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM training_sessions 
			WHERE coach_id = ? AND end_time > CURRENT_TIMESTAMP
		`, user.ID).Scan(&stats.ActiveSessions)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get total unique students
		err = db.QueryRow(`
			SELECT COUNT(DISTINCT user_id) 
			FROM training_session_participants tsp
			JOIN training_sessions ts ON ts.id = tsp.session_id
			WHERE ts.coach_id = ?
		`, user.ID).Scan(&stats.TotalStudents)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Calculate total hours taught
		err = db.QueryRow(`
			SELECT COALESCE(SUM(
				ROUND(CAST(
					(JULIANDAY(end_time) - JULIANDAY(start_time)) * 24 
				AS REAL), 2)
			), 0)
			FROM training_sessions
			WHERE coach_id = ? AND end_time <= CURRENT_TIMESTAMP
		`, user.ID).Scan(&stats.HoursTaught)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load statistics"})
			return
		}

		// Get all available courts
		courts, err := models.GetAvailableCourts(db)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load courts"})
			return
		}

		// Get coach's training sessions
		sessions, err := models.GetTrainingSessionsByCoach(db, user.ID)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to load sessions"})
			return
		}

		c.HTML(http.StatusOK, "coach_dashboard.html", gin.H{
			"title": "Coach Dashboard",
			"user":  user,
			"stats": stats,
			"courts": courts,
			"sessions": sessions,
		})
	}
}

// CreateTrainingSessionHandler handles creation of new training sessions
func CreateTrainingSessionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleCoach {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var session models.TrainingSession
		if err := c.ShouldBindJSON(&session); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session.CoachID = user.ID

		// Validate time slot availability
		available, err := models.IsCourtAvailable(db, session.CourtID, session.StartTime, session.EndTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check court availability"})
			return
		}
		if !available {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Court is not available for the selected time slot"})
			return
		}

		err = models.CreateTrainingSession(db, &session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create training session"})
			return
		}

		c.JSON(http.StatusOK, session)
	}
}

// UpdateTrainingSessionHandler handles updates to training sessions
func UpdateTrainingSessionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleCoach {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var session models.TrainingSession
		if err := c.ShouldBindJSON(&session); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Verify the session belongs to this coach
		existingSession, err := models.GetTrainingSessionByID(db, session.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		if existingSession.CoachID != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Check if the time slot is still available
		if session.StartTime != existingSession.StartTime || session.EndTime != existingSession.EndTime {
			available, err := models.IsCourtAvailable(db, session.CourtID, session.StartTime, session.EndTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check court availability"})
				return
			}
			if !available {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Court is not available for the selected time slot"})
				return
			}
		}

		err = models.UpdateTrainingSession(db, &session)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update training session"})
			return
		}

		c.JSON(http.StatusOK, session)
	}
}

// DeleteTrainingSessionHandler handles deletion of training sessions
func DeleteTrainingSessionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleCoach {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sessionID := c.Param("id")

		// Verify the session belongs to this coach
		session, err := models.GetTrainingSessionByID(db, sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}
		if session.CoachID != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Don't allow deletion of sessions that have already started
		if session.StartTime.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete sessions that have already started"})
			return
		}

		err = models.DeleteTrainingSession(db, sessionID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete training session"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Training session deleted successfully"})
	}
}

// GetTrainingSessionHandler handles retrieving a single training session
func GetTrainingSessionHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := middleware.GetCurrentUser(c)
		if user == nil || user.Role != models.RoleCoach {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		sessionID := c.Param("id")
		session, err := models.GetTrainingSessionByID(db, sessionID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Verify the session belongs to this coach
		if session.CoachID != user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.JSON(http.StatusOK, session)
	}
}
