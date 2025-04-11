package models

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type Booking struct {
	ID         int64
	CourtID    int64
	UserID     int64
	StartTime  time.Time
	EndTime    time.Time
	Status     string
	BookingType string
	CreatedAt  time.Time
	
	// Additional fields for joins
	CourtName  string
	UserName   string
}

type TrainingSession struct {
	ID              int64
	CoachID         int64
	CourtID         int64
	Title           string
	Description     string
	StartTime       time.Time
	EndTime         time.Time
	MaxParticipants int
	CreatedAt       time.Time

	// Additional fields for joins
	CoachName       string
	CourtName       string
}

const (
	BookingStatusPending   = "pending"
	BookingStatusConfirmed = "confirmed"
	BookingStatusCancelled = "cancelled"
	
	BookingTypeRegular  = "regular"
	BookingTypeTraining = "training"
)

// CreateBooking creates a new booking in the database
func CreateBooking(db *sql.DB, booking *Booking) error {
	// Check if the court is available
	available, err := IsCourtAvailable(db, booking.CourtID, booking.StartTime, booking.EndTime)
	if err != nil {
		return err
	}
	if !available {
		return errors.New("court is not available for the selected time slot")
	}

	query := `
		INSERT INTO bookings (court_id, user_id, start_time, end_time, status, booking_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query, 
		booking.CourtID, 
		booking.UserID, 
		booking.StartTime, 
		booking.EndTime, 
		booking.Status,
		booking.BookingType,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	booking.ID = id
	return nil
}

// GetBookingByID retrieves a booking by its ID with joined court and user information
func GetBookingByID(db *sql.DB, id interface{}) (*Booking, error) {
	var bookingID int64
	switch v := id.(type) {
	case int64:
		bookingID = v
	case string:
		var err error
		bookingID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid ID type")
	}

	booking := &Booking{}
	query := `
		SELECT 
			b.id, b.court_id, b.user_id, b.start_time, b.end_time, 
			b.status, b.booking_type, b.created_at,
			c.name as court_name, u.username as user_name
		FROM bookings b
		JOIN courts c ON b.court_id = c.id
		JOIN users u ON b.user_id = u.id
		WHERE b.id = ?
	`
	err := db.QueryRow(query, bookingID).Scan(
		&booking.ID, &booking.CourtID, &booking.UserID, 
		&booking.StartTime, &booking.EndTime, &booking.Status, 
		&booking.BookingType, &booking.CreatedAt,
		&booking.CourtName, &booking.UserName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("booking not found")
		}
		return nil, err
	}
	return booking, nil
}

// GetRecentBookings retrieves the most recent bookings
func GetRecentBookings(db *sql.DB, limit int) ([]*Booking, error) {
	query := `
		SELECT 
			b.id, b.court_id, b.user_id, b.start_time, b.end_time, 
			b.status, b.booking_type, b.created_at,
			c.name as court_name, u.username as user_name
		FROM bookings b
		JOIN courts c ON b.court_id = c.id
		JOIN users u ON b.user_id = u.id
		ORDER BY b.created_at DESC
		LIMIT ?
	`
	return executeBookingQuery(db, query, limit)
}

// GetAllBookings retrieves all bookings
func GetAllBookings(db *sql.DB) ([]*Booking, error) {
	query := `
		SELECT 
			b.id, b.court_id, b.user_id, b.start_time, b.end_time, 
			b.status, b.booking_type, b.created_at,
			c.name as court_name, u.username as user_name
		FROM bookings b
		JOIN courts c ON b.court_id = c.id
		JOIN users u ON b.user_id = u.id
		ORDER BY b.start_time DESC
	`
	return executeBookingQuery(db, query)
}

// GetUserBookings retrieves all bookings for a specific user
func GetUserBookings(db *sql.DB, userID int64) ([]*Booking, error) {
	query := `
		SELECT 
			b.id, b.court_id, b.user_id, b.start_time, b.end_time, 
			b.status, b.booking_type, b.created_at,
			c.name as court_name, u.username as user_name
		FROM bookings b
		JOIN courts c ON b.court_id = c.id
		JOIN users u ON b.user_id = u.id
		WHERE b.user_id = ?
		ORDER BY b.start_time DESC
	`
	return executeBookingQuery(db, query, userID)
}

// GetCourtBookings retrieves all bookings for a specific court
func GetCourtBookings(db *sql.DB, courtID int64) ([]*Booking, error) {
	query := `
		SELECT 
			b.id, b.court_id, b.user_id, b.start_time, b.end_time, 
			b.status, b.booking_type, b.created_at,
			c.name as court_name, u.username as user_name
		FROM bookings b
		JOIN courts c ON b.court_id = c.id
		JOIN users u ON b.user_id = u.id
		WHERE b.court_id = ?
		ORDER BY b.start_time DESC
	`
	return executeBookingQuery(db, query, courtID)
}

// Helper function to execute booking queries
func executeBookingQuery(db *sql.DB, query string, args ...interface{}) ([]*Booking, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.CourtID, &booking.UserID, 
			&booking.StartTime, &booking.EndTime, &booking.Status, 
			&booking.BookingType, &booking.CreatedAt,
			&booking.CourtName, &booking.UserName,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

// UpdateBookingStatus updates the status of a booking
func UpdateBookingStatus(db *sql.DB, id interface{}, status string) error {
	var bookingID int64
	switch v := id.(type) {
	case int64:
		bookingID = v
	case string:
		var err error
		bookingID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid ID type")
	}

	query := `UPDATE bookings SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, bookingID)
	return err
}

// CancelBooking cancels a booking
func CancelBooking(db *sql.DB, id interface{}) error {
	var bookingID int64
	switch v := id.(type) {
	case int64:
		bookingID = v
	case string:
		var err error
		bookingID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid ID type")
	}

	return UpdateBookingStatus(db, bookingID, BookingStatusCancelled)
}

// CreateTrainingSession creates a new training session
func CreateTrainingSession(db *sql.DB, session *TrainingSession) error {
	// Create a booking for the training session
	booking := &Booking{
		CourtID:    session.CourtID,
		UserID:     session.CoachID,
		StartTime:  session.StartTime,
		EndTime:    session.EndTime,
		Status:     BookingStatusConfirmed,
		BookingType: BookingTypeTraining,
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Create the booking first
	err = CreateBooking(db, booking)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Then create the training session
	query := `
		INSERT INTO training_sessions (
			coach_id, court_id, title, description, 
			start_time, end_time, max_participants, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := tx.Exec(query,
		session.CoachID, session.CourtID, session.Title,
		session.Description, session.StartTime, session.EndTime,
		session.MaxParticipants,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	session.ID = id
	return tx.Commit()
}

// GetTrainingSessionByID retrieves a training session by its ID
func GetTrainingSessionByID(db *sql.DB, id interface{}) (*TrainingSession, error) {
	var sessionID int64
	switch v := id.(type) {
	case int64:
		sessionID = v
	case string:
		var err error
		sessionID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid ID type")
	}

	session := &TrainingSession{}
	query := `
		SELECT 
			t.id, t.coach_id, t.court_id, t.title, t.description,
			t.start_time, t.end_time, t.max_participants, t.created_at,
			u.username as coach_name, c.name as court_name
		FROM training_sessions t
		JOIN users u ON t.coach_id = u.id
		JOIN courts c ON t.court_id = c.id
		WHERE t.id = ?
	`
	err := db.QueryRow(query, sessionID).Scan(
		&session.ID, &session.CoachID, &session.CourtID,
		&session.Title, &session.Description, &session.StartTime,
		&session.EndTime, &session.MaxParticipants, &session.CreatedAt,
		&session.CoachName, &session.CourtName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("training session not found")
		}
		return nil, err
	}
	return session, nil
}

// UpdateTrainingSession updates an existing training session
func UpdateTrainingSession(db *sql.DB, session *TrainingSession) error {
	query := `
		UPDATE training_sessions 
		SET title = ?, description = ?, court_id = ?,
			start_time = ?, end_time = ?, max_participants = ?
		WHERE id = ? AND coach_id = ?
	`
	result, err := db.Exec(query,
		session.Title, session.Description, session.CourtID,
		session.StartTime, session.EndTime, session.MaxParticipants,
		session.ID, session.CoachID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("training session not found or not authorized")
	}
	return nil
}

// DeleteTrainingSession deletes a training session
func DeleteTrainingSession(db *sql.DB, id interface{}) error {
	var sessionID int64
	switch v := id.(type) {
	case int64:
		sessionID = v
	case string:
		var err error
		sessionID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid ID type")
	}

	query := `DELETE FROM training_sessions WHERE id = ?`
	result, err := db.Exec(query, sessionID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("training session not found")
	}
	return nil
}

// GetAvailableTrainingSessions retrieves all available training sessions
func GetAvailableTrainingSessions(db *sql.DB) ([]*TrainingSession, error) {
	query := `
		SELECT 
			t.id, t.coach_id, t.court_id, t.title, t.description,
			t.start_time, t.end_time, t.max_participants, t.created_at,
			u.username as coach_name, c.name as court_name,
			(SELECT COUNT(*) FROM training_session_participants WHERE session_id = t.id) as current_participants
		FROM training_sessions t
		JOIN users u ON t.coach_id = u.id
		JOIN courts c ON t.court_id = c.id
		WHERE t.end_time > CURRENT_TIMESTAMP
		ORDER BY t.start_time ASC
	`
	return executeTrainingSessionQuery(db, query)
}

// IsUserEnrolled checks if a user is enrolled in a training session
func IsUserEnrolled(db *sql.DB, userID int64, sessionID interface{}) (bool, error) {
	var sID int64
	switch v := sessionID.(type) {
	case int64:
		sID = v
	case string:
		var err error
		sID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return false, err
		}
	default:
		return false, errors.New("invalid session ID type")
	}

	var count int
	query := `SELECT COUNT(*) FROM training_session_participants WHERE user_id = ? AND session_id = ?`
	err := db.QueryRow(query, userID, sID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// EnrollInTrainingSession enrolls a user in a training session
func EnrollInTrainingSession(db *sql.DB, userID int64, sessionID interface{}) error {
	var sID int64
	switch v := sessionID.(type) {
	case int64:
		sID = v
	case string:
		var err error
		sID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid session ID type")
	}

	// Check if session exists and has space
	session, err := GetTrainingSessionByID(db, sID)
	if err != nil {
		return err
	}

	// Get current participant count
	var count int
	query := `SELECT COUNT(*) FROM training_session_participants WHERE session_id = ?`
	err = db.QueryRow(query, sID).Scan(&count)
	if err != nil {
		return err
	}

	if count >= session.MaxParticipants {
		return errors.New("session is full")
	}

	// Check if user is already enrolled
	enrolled, err := IsUserEnrolled(db, userID, sID)
	if err != nil {
		return err
	}
	if enrolled {
		return errors.New("already enrolled in this session")
	}

	// Enroll user
	query = `INSERT INTO training_session_participants (user_id, session_id) VALUES (?, ?)`
	_, err = db.Exec(query, userID, sID)
	return err
}

// CancelTrainingEnrollment cancels a user's enrollment in a training session
func CancelTrainingEnrollment(db *sql.DB, userID int64, sessionID interface{}) error {
	var sID int64
	switch v := sessionID.(type) {
	case int64:
		sID = v
	case string:
		var err error
		sID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid session ID type")
	}

	query := `DELETE FROM training_session_participants WHERE user_id = ? AND session_id = ?`
	result, err := db.Exec(query, userID, sID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("enrollment not found")
	}
	return nil
}

// GetTrainingSessionsByCoach retrieves all training sessions for a specific coach
func GetTrainingSessionsByCoach(db *sql.DB, coachID int64) ([]*TrainingSession, error) {
	query := `
		SELECT 
			t.id, t.coach_id, t.court_id, t.title, t.description,
			t.start_time, t.end_time, t.max_participants, t.created_at,
			u.username as coach_name, c.name as court_name
		FROM training_sessions t
		JOIN users u ON t.coach_id = u.id
		JOIN courts c ON t.court_id = c.id
		WHERE t.coach_id = ?
		ORDER BY t.start_time DESC
	`
	return executeTrainingSessionQuery(db, query, coachID)
}

// Helper function to execute training session queries
func executeTrainingSessionQuery(db *sql.DB, query string, args ...interface{}) ([]*TrainingSession, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*TrainingSession
	for rows.Next() {
		session := &TrainingSession{}
		err := rows.Scan(
			&session.ID, &session.CoachID, &session.CourtID,
			&session.Title, &session.Description, &session.StartTime,
			&session.EndTime, &session.MaxParticipants, &session.CreatedAt,
			&session.CoachName, &session.CourtName,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}
