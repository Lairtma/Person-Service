package main

import (
	"fmt"
	"log"
	"os"

	"person-service/docs"
	"person-service/handlers"
	"person-service/migrations"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	// Создаем базу данных, если она не существует
	if err := migrations.EnsureDatabase(os.Getenv("DB_NAME")); err != nil {
		log.Fatal("Failed to ensure database:", err)
	}

	// Создаем подключение к базе данных
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Инициализируем GORM
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Выполняем миграции
	if err := migrations.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
}

// @title Person Service API
// @version 1.0
// @description A service that enriches person data with age, gender, and nationality information
// @host localhost:8080
// @BasePath /
func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	initDB()

	// Initialize router
	r := gin.Default()

	// Swagger documentation
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Routes
	r.POST("/people", func(c *gin.Context) {
		handlers.CreatePerson(c, db)
	})
	r.GET("/people", func(c *gin.Context) {
		handlers.GetPeople(c, db)
	})
	r.PUT("/people/:id", func(c *gin.Context) {
		handlers.UpdatePerson(c, db)
	})
	r.DELETE("/people/:id", func(c *gin.Context) {
		handlers.DeletePerson(c, db)
	})

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
