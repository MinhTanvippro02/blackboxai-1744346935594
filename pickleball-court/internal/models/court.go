package models

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type Court struct {
	ID          int64
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
}

const (
	CourtStatusAvailable = "available"
	CourtStatusBooked    = "booked"
	CourtStatusMaintenance = "maintenance"
)

// CreateCourt creates a new court in the database
func CreateCourt(db *sql.DB, court *Court) error {
	query := `
		INSERT INTO courts (name, description, status, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query, court.Name, court.Description, court.Status)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	court.ID = id
	return nil
}

// GetCourtByID retrieves a court by its ID
func GetCourtByID(db *sql.DB, id int64) (*Court, error) {
	court := &Court{}
	query := `SELECT id, name, description, status, created_at FROM courts WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&court.ID, &court.Name, &court.Description, &court.Status, &court.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("court not found")
		}
		return nil, err
	}
	return court, nil
}

// GetAllCourts retrieves all courts from the database
func GetAllCourts(db *sql.DB) ([]*Court, error) {
	query := `SELECT id, name, description, status, created_at FROM courts`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courts []*Court
	for rows.Next() {
		court := &Court{}
		err := rows.Scan(&court.ID, &court.Name, &court.Description, &court.Status, &court.CreatedAt)
		if err != nil {
			return nil, err
		}
		courts = append(courts, court)
	}
	return courts, nil
}

// GetAvailableCourts retrieves all available courts
func GetAvailableCourts(db *sql.DB) ([]*Court, error) {
	query := `SELECT id, name, description, status, created_at FROM courts WHERE status = ?`
	rows, err := db.Query(query, CourtStatusAvailable)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courts []*Court
	for rows.Next() {
		court := &Court{}
		err := rows.Scan(&court.ID, &court.Name, &court.Description, &court.Status, &court.CreatedAt)
		if err != nil {
			return nil, err
		}
		courts = append(courts, court)
	}
	return courts, nil
}

// UpdateCourt updates court information
func UpdateCourt(db *sql.DB, court *Court) error {
	query := `
		UPDATE courts 
		SET name = ?, description = ?, status = ?
		WHERE id = ?
	`
	_, err := db.Exec(query, court.Name, court.Description, court.Status, court.ID)
	return err
}

// UpdateCourtStatus updates only the court's status
func UpdateCourtStatus(db *sql.DB, courtID int64, status string) error {
	query := `UPDATE courts SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, courtID)
	return err
}

// DeleteCourt deletes a court from the database
func DeleteCourt(db *sql.DB, id interface{}) error {
	var courtID int64
	switch v := id.(type) {
	case int64:
		courtID = v
	case string:
		var err error
		courtID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid ID type")
	}

	// First check if there are any active bookings for this court
	query := `
		SELECT COUNT(*) FROM bookings 
		WHERE court_id = ? AND end_time > CURRENT_TIMESTAMP
	`
	var count int
	err := db.QueryRow(query, courtID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete court with active bookings")
	}

	// If no active bookings, proceed with deletion
	query = `DELETE FROM courts WHERE id = ?`
	_, err = db.Exec(query, courtID)
	return err
}

// IsCourtAvailable checks if a court is available for booking in a given time slot
func IsCourtAvailable(db *sql.DB, courtID int64, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) FROM bookings 
		WHERE court_id = ? 
		AND status != 'cancelled'
		AND (
			(start_time <= ? AND end_time > ?) OR
			(start_time < ? AND end_time >= ?) OR
			(start_time >= ? AND end_time <= ?)
		)
	`
	var count int
	err := db.QueryRow(
		query,
		courtID,
		startTime, startTime,
		endTime, endTime,
		startTime, endTime,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
