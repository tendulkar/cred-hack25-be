package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cred.com/hack25/backend/internal/config"
	"cred.com/hack25/backend/internal/handlers"
	"cred.com/hack25/backend/internal/middleware"
	"cred.com/hack25/backend/internal/repository"
	"cred.com/hack25/backend/internal/service"
	"cred.com/hack25/backend/pkg/auth"
	"cred.com/hack25/backend/pkg/database"
	"cred.com/hack25/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Setup database connection
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := db.InitSchema(); err != nil {
		logger.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Conn)

	// Initialize JWT service
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.SigningAlgorithm,
	)

	// Initialize services
	userService := service.NewUserService(userRepo, jwtService)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup router
	router := gin.Default()

	// Add global middleware
	router.Use(middleware.Cors())
	router.Use(middleware.RequestLogger())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// Protected routes
		user := api.Group("/user")
		user.Use(authMiddleware.RequireAuth())
		{
			user.GET("/profile", userHandler.GetProfile)
			user.PUT("/profile", userHandler.UpdateProfile)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(authMiddleware.RequireAuth(), authMiddleware.RequireRole("admin"))
		{
			admin.GET("/users", userHandler.ListUsers)
		}
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Server started on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Shutdown with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
