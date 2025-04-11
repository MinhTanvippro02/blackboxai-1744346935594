package main

import (
	"log"
	"os"
	"time"

	"pickleball-court/internal/handlers"
	"pickleball-court/internal/middleware"
	"pickleball-court/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	// Create the router
	router := gin.Default()

	// Initialize session middleware
	store := cookie.NewStore([]byte(getEnv("SESSION_SECRET", "your-secret-key")))
	store.Options(sessions.Options{
		MaxAge:   int(24 * time.Hour.Seconds()), // 24 hours
		Path:     "/",
		HttpOnly: true,
		Secure:   getEnv("ENV", "development") == "production",
	})
	router.Use(sessions.Sessions("pickleball_session", store))

	// Initialize database
	db, err := models.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create necessary directories if they don't exist
	dirs := []string{"static", "static/css", "static/js", "static/images", "templates"}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatal("Failed to create directory:", err)
		}
	}

	// Set up template rendering
	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Set up middleware
	router.Use(middleware.LoadUser(db))

	// Public routes
	router.GET("/", handlers.HomeHandler(db))
	router.GET("/login", handlers.ShowLoginHandler())
	router.POST("/login", handlers.LoginHandler(db))
	router.GET("/register", handlers.ShowRegisterHandler())
	router.POST("/register", handlers.RegisterHandler(db))
	router.GET("/logout", handlers.LogoutHandler())

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(middleware.AuthRequired())
	{
		// Profile routes
		authorized.GET("/profile", handlers.ProfileHandler(db))
		authorized.POST("/profile/update", handlers.UpdateProfileHandler(db))
		authorized.POST("/profile/password", handlers.UpdatePasswordHandler(db))

		// Admin routes
		admin := authorized.Group("/admin")
		admin.Use(middleware.RoleRequired("admin"))
		{
			admin.GET("/dashboard", handlers.AdminDashboardHandler(db))
			
			// User management
			admin.GET("/users", handlers.ListUsersHandler(db))
			admin.POST("/users", handlers.CreateUserHandler(db))
			admin.PUT("/users/:id", handlers.UpdateUserHandler(db))
			admin.DELETE("/users/:id", handlers.DeleteUserHandler(db))
			
			// Court management
			admin.GET("/courts", handlers.ListCourtsHandler(db))
			admin.POST("/courts", handlers.CreateCourtHandler(db))
			admin.PUT("/courts/:id", handlers.UpdateCourtHandler(db))
			admin.DELETE("/courts/:id", handlers.DeleteCourtHandler(db))
			
			// Booking management
			admin.GET("/bookings", handlers.ListAllBookingsHandler(db))
			admin.PUT("/bookings/:id", handlers.UpdateBookingHandler(db))
		}

		// Coach routes
		coach := authorized.Group("/coach")
		coach.Use(middleware.RoleRequired("coach"))
		{
			coach.GET("/dashboard", handlers.CoachDashboardHandler(db))
			coach.GET("/sessions", handlers.ListTrainingSessionsHandler(db))
			coach.POST("/sessions", handlers.CreateTrainingSessionHandler(db))
			coach.PUT("/sessions/:id", handlers.UpdateTrainingSessionHandler(db))
			coach.DELETE("/sessions/:id", handlers.DeleteTrainingSessionHandler(db))
		}

		// Player routes
		player := authorized.Group("/player")
		player.Use(middleware.RoleRequired("player"))
		{
			player.GET("/dashboard", handlers.PlayerDashboardHandler(db))
			player.GET("/courts/availability", handlers.GetCourtAvailabilityHandler(db))
			player.POST("/bookings", handlers.CreateBookingHandler(db))
			player.POST("/bookings/:id/cancel", handlers.CancelBookingHandler(db))
			player.POST("/training/:id/enroll", handlers.EnrollTrainingHandler(db))
			player.POST("/training/:id/cancel", handlers.CancelTrainingEnrollmentHandler(db))
		}
	}

	// Error handlers
	router.NoRoute(func(c *gin.Context) {
		c.HTML(404, "error.html", gin.H{
			"title": "Page Not Found",
			"code":  404,
		})
	})

	// Start the server
	port := getEnv("PORT", "8000")
	log.Printf("Server starting on port %s...\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
