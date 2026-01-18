package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (only in development)
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env file not found (this is normal in production)")
	} else {
		log.Println("✓ .env file loaded")
	}

	// Set Gin mode based on environment
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db := initDB()
	defer db.Close()

	// Set DB for middleware
	SetDB(db)

	// Create Gin router
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Initialize handlers
	handlers := NewHandlers(db)

	// Public routes
	r.GET("/", handlers.HomeHandler)
	r.POST("/signup", handlers.SignupHandler)

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Admin routes with basic auth
	admin := r.Group("/admin")
	admin.Use(BasicAuth())
	{
		admin.GET("", handlers.AdminHandler)
		admin.POST("create-workshop", handlers.CreateWorkshopHandler)
		admin.POST("change-password", handlers.ChangePasswordHandler)
		admin.GET("export-csv", handlers.ExportCSVHandler)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s\n", port)
	log.Fatal(r.Run(":" + port))
}
