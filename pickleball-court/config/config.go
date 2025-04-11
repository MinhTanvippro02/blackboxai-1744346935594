package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration settings
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Session  SessionConfig
	Booking  BookingConfig
	Email    EmailConfig
}

// ServerConfig holds server-related settings
type ServerConfig struct {
	Port         string
	Environment  string
	AllowOrigins []string
	TimeZone     *time.Location
}

// DatabaseConfig holds database-related settings
type DatabaseConfig struct {
	Path string
}

// SessionConfig holds session-related settings
type SessionConfig struct {
	Secret   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

// BookingConfig holds booking-related settings
type BookingConfig struct {
	MaxDaysAhead     int
	MinHoursAdvance  int
	MaxHoursPerWeek  int
	OpeningHour      int
	ClosingHour      int
	SlotDuration     time.Duration
	CancellationTime time.Duration
}

// EmailConfig holds email-related settings
type EmailConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

var (
	config *Config
)

// Load initializes the configuration from environment variables
func Load() *Config {
	timezone, err := time.LoadLocation(getEnv("TZ", "UTC"))
	if err != nil {
		timezone = time.UTC
	}

	config = &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8000"),
			Environment:  getEnv("ENV", "development"),
			AllowOrigins: []string{"http://localhost:8000"},
			TimeZone:     timezone,
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./pickleball.db"),
		},
		Session: SessionConfig{
			Secret:   getEnv("SESSION_SECRET", "your-secret-key"),
			MaxAge:   getEnvAsInt("SESSION_MAX_AGE", 86400), // 24 hours
			Secure:   getEnv("ENV", "development") == "production",
			HttpOnly: true,
		},
		Booking: BookingConfig{
			MaxDaysAhead:     getEnvAsInt("BOOKING_MAX_DAYS_AHEAD", 14),
			MinHoursAdvance:  getEnvAsInt("BOOKING_MIN_HOURS_ADVANCE", 1),
			MaxHoursPerWeek:  getEnvAsInt("BOOKING_MAX_HOURS_PER_WEEK", 10),
			OpeningHour:      getEnvAsInt("OPENING_HOUR", 6),  // 6 AM
			ClosingHour:      getEnvAsInt("CLOSING_HOUR", 22), // 10 PM
			SlotDuration:     time.Hour,                        // 1 hour slots
			CancellationTime: time.Hour * 24,                   // 24 hours notice required
		},
		Email: EmailConfig{
			Enabled:  getEnvAsBool("EMAIL_ENABLED", false),
			Host:     getEnv("EMAIL_HOST", "smtp.gmail.com"),
			Port:     getEnvAsInt("EMAIL_PORT", 587),
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
			From:     getEnv("EMAIL_FROM", "noreply@picklecourt.com"),
		},
	}

	return config
}

// Get returns the current configuration
func Get() *Config {
	if config == nil {
		return Load()
	}
	return config
}

// Helper functions to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// IsEmailEnabled returns true if email notifications are enabled
func (c *Config) IsEmailEnabled() bool {
	return c.Email.Enabled
}

// GetBookingTimeConstraints returns the valid booking hours
func (c *Config) GetBookingTimeConstraints() (openingHour, closingHour int) {
	return c.Booking.OpeningHour, c.Booking.ClosingHour
}

// GetMaxBookingDaysAhead returns the maximum number of days ahead that bookings can be made
func (c *Config) GetMaxBookingDaysAhead() int {
	return c.Booking.MaxDaysAhead
}

// GetMinBookingAdvanceTime returns the minimum hours in advance required for bookings
func (c *Config) GetMinBookingAdvanceTime() int {
	return c.Booking.MinHoursAdvance
}

// GetMaxBookingHoursPerWeek returns the maximum hours a user can book per week
func (c *Config) GetMaxBookingHoursPerWeek() int {
	return c.Booking.MaxHoursPerWeek
}

// GetCancellationNoticeRequired returns the required notice period for cancellations
func (c *Config) GetCancellationNoticeRequired() time.Duration {
	return c.Booking.CancellationTime
}

// GetTimeZone returns the application's timezone
func (c *Config) GetTimeZone() *time.Location {
	return c.Server.TimeZone
}
