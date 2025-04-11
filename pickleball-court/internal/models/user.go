package models

import (
	"database/sql"
	"errors"
	"strconv"
	"time"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64
	Username  string
	Password  string
	Email     string
	Role      string
	CreatedAt time.Time
}

const (
	RoleAdmin  = "admin"
	RoleCoach  = "coach"
	RolePlayer = "player"
)

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, user *User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (username, password, email, role, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`

	result, err := db.Exec(query, user.Username, string(hashedPassword), user.Email, user.Role)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(db *sql.DB, id int64) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, created_at FROM users WHERE id = ?`
	err := db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username
func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, email, role, created_at FROM users WHERE username = ?`
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// AuthenticateUser verifies user credentials and returns the user if valid
func AuthenticateUser(db *sql.DB, username, password string) (*User, error) {
	user, err := GetUserByUsername(db, username)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// UpdateUser updates user information
func UpdateUser(db *sql.DB, user *User) error {
	query := `
		UPDATE users 
		SET username = ?, email = ?, role = ?
		WHERE id = ?
	`
	_, err := db.Exec(query, user.Username, user.Email, user.Role, user.ID)
	return err
}

// UpdatePassword updates a user's password
func UpdatePassword(db *sql.DB, userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err = db.Exec(query, string(hashedPassword), userID)
	return err
}

// DeleteUser deletes a user from the database
func DeleteUser(db *sql.DB, id interface{}) error {
	var userID int64
	switch v := id.(type) {
	case int64:
		userID = v
	case string:
		var err error
		userID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid ID type")
	}

	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Exec(query, userID)
	return err
}

// GetAllUsers retrieves all users from the database
func GetAllUsers(db *sql.DB) ([]*User, error) {
	query := `SELECT id, username, password, email, role, created_at FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetUsersByRole retrieves all users with a specific role
func GetUsersByRole(db *sql.DB, role string) ([]*User, error) {
	query := `SELECT id, username, password, email, role, created_at FROM users WHERE role = ?`
	rows, err := db.Query(query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
