package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "golang-test/docs" // Импортируем сгенерированную документацию

	"github.com/gin-gonic/gin"

	"golang-test/internal/db"
	"golang-test/internal/handlers"
	"golang-test/internal/middleware"
	"golang-test/internal/repository"

	"github.com/pressly/goose/v3"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Golang Test API
// @version 1.0
// @description API для управления объявлениями и пользователями
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey APIKey
// @in header
// @name Key
func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		if err := os.Mkdir("migrations", 0755); err != nil {
			return err
		}
		slog.Info("created migrations directory")
	}
	return goose.Up(db, "migrations")
}

func createUploadsDir() error {
	dirs := []string{"uploads", "uploads/images"}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0755); err != nil {
				return err
			}
			slog.Info("created directory", "dir", dir)
		}
	}
	return nil
}

// @Summary Проверка работоспособности сервера
// @Description Health check endpoint
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// @Summary Инициализация категорий
// @Description Добавляет тестовые категории в базу данных
// @Tags categories
// @Accept json
// @Produce json
// @Security APIKey
// @Success 200 {object} map[string]string
// @Failure 500 {object} ErrorResponse
// @Router /init-categories [get]
func initCategories(c *gin.Context, database *sql.DB) {
	ctx := c.Request.Context()
	_, err := database.ExecContext(ctx, `
		INSERT INTO categories (name, extra_property) 
		VALUES 
			('Electronics', 'Color: Black'),
			('Clothing', 'Size: M')
		ON CONFLICT (name) DO NOTHING
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "categories initialized"})
}

func main() {
	// Создаем директории для загрузки файлов
	if err := createUploadsDir(); err != nil {
		slog.Error("failed to create uploads directory", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := "host=localhost user=postgres password=123 dbname=golang_db port=5432 sslmode=disable"

	database, err := db.Connect(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := runMigrations(database); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("database migrations applied successfully")

	// Инициализируем репозитории
	adRepo := repository.NewAdRepository(database)
	userRepo := repository.NewUserRepository(database)

	// Инициализируем обработчики
	adHandler := handlers.NewAdHandler(adRepo)
	userHandler := handlers.NewUserHandler(userRepo)

	r := gin.Default()

	// Добавляем Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Добавляем middleware авторизации ко всем маршрутам
	r.Use(middleware.AuthMiddleware())

	// Маршруты для объявлений
	adRoutes := r.Group("/ads")
	{
		adRoutes.GET("/:id", adHandler.GetAdByID)
		adRoutes.GET("", adHandler.GetAllAds)
		adRoutes.POST("", adHandler.CreateAd)
		adRoutes.PUT("/:id", adHandler.UpdateAd)
		adRoutes.PATCH("/:id/toggle", adHandler.ToggleAd)
		adRoutes.DELETE("/:id", adHandler.DeleteAd)
	}

	// Маршруты для пользователей
	userRoutes := r.Group("/users")
	{
		userRoutes.POST("", userHandler.CreateUser)
		userRoutes.DELETE("/:id", userHandler.DeleteUser)
	}

	// Добавляем тестовые данные для категорий
	r.GET("/init-categories", func(c *gin.Context) {
		initCategories(c, database)
	})

	// Health check
	r.GET("/health", healthCheck)

	log.Println("Server started on :8080")
	log.Println("Swagger UI available at: http://localhost:8080/swagger/index.html")
	r.Run(":8080")
}
