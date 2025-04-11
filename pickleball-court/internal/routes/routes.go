package routes

import (
	"database/sql"
	"net/http"
	"pickleball-court/internal/handlers"
	"pickleball-court/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, db *sql.DB) {
	// Middleware
	router.Use(middleware.LoadUser(db))

	// Static files
	router.Static("/static", "./static")

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
		// Common user routes
		authorized.GET("/profile", handlers.ProfileHandler(db))
		authorized.POST("/profile/update", handlers.UpdateProfileHandler(db))
		authorized.POST("/profile/password", handlers.UpdatePasswordHandler(db))

		// Court viewing routes
		authorized.GET("/courts", handlers.ListCourtsHandler(db))
		authorized.GET("/courts/:id", handlers.GetCourtHandler(db))

		// Booking routes
		authorized.GET("/bookings", handlers.ListBookingsHandler(db))
		authorized.GET("/bookings/:id", handlers.GetBookingHandler(db))
		authorized.POST("/bookings", handlers.CreateBookingHandler(db))
		authorized.POST("/bookings/:id/cancel", handlers.CancelBookingHandler(db))

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
			admin.POST("/courts", handlers.CreateCourtHandler(db))
			admin.PUT("/courts/:id", handlers.UpdateCourtHandler(db))
			admin.DELETE("/courts/:id", handlers.DeleteCourtHandler(db))
			
			// Booking management
			admin.GET("/bookings/all", handlers.ListAllBookingsHandler(db))
			admin.PUT("/bookings/:id", handlers.UpdateBookingHandler(db))
		}

		// Coach routes
		coach := authorized.Group("/coach")
		coach.Use(middleware.RoleRequired("coach"))
		{
			coach.GET("/dashboard", handlers.CoachDashboardHandler(db))
			
			// Training session management
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
			
			// Training session enrollment
			player.GET("/training", handlers.ListAvailableTrainingHandler(db))
			player.POST("/training/:id/enroll", handlers.EnrollTrainingHandler(db))
			player.POST("/training/:id/cancel", handlers.CancelTrainingEnrollmentHandler(db))
		}
	}

	// Error handlers
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"title": "Page Not Found",
			"error": "The page you're looking for doesn't exist.",
		})
	})
}
