package main

import (
	"log"
	"pickleball-court/internal/models"
	"pickleball-court/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

func main() {
	// Initialize database
	db, err := models.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	// Initialize router
	router := gin.Default()

	// Setup session middleware
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("pickleball_session", store))

	// Setup templates
	router.LoadHTMLGlob("templates/*")

	// Setup static file serving
	router.Static("/static", "./static")

	// Initialize routes
	routes.SetupRoutes(router, db)

	// Start server
	log.Println("Server starting on :8000")
	if err := router.Run(":8000"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
