package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/yourname/stockbit-appsec/internal/delivery/http/v1_secure"
	"github.com/yourname/stockbit-appsec/internal/delivery/http/v1_vulnerable"
	"github.com/yourname/stockbit-appsec/internal/middleware"
	"github.com/yourname/stockbit-appsec/internal/repository"
	"github.com/yourname/stockbit-appsec/internal/usecase"
	"github.com/yourname/stockbit-appsec/pkg/utils"
)

func main() {
	// 1. Setup Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Database Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/stockbit_db?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		slog.Error("Failed to open database connection", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Connected to database")

	// 2.5 Seed Data (For Demo Purposes)
	seedData(db)

	// 3. Setup Dependencies
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	jwtSecret := "super-secret-key" // In prod, use environment variable
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtSecret)
	orderUsecase := usecase.NewOrderUsecase(orderRepo)

	// Handlers
	vulnerableOrderHandler := v1_vulnerable.NewOrderHandler(orderUsecase)
	secureOrderHandler := v1_secure.NewOrderHandler(orderUsecase)

	// 4. Setup Gin Router
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// Public Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/register", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := authUsecase.Register(c.Request.Context(), req.Email, req.Password, req.Role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "User registered"})
	})

	r.POST("/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		token, err := authUsecase.Login(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	// Protected Routes
	api := r.Group("/api")
	api.Use(middleware.Auth(jwtSecret))
	{
		api.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			c.JSON(http.StatusOK, gin.H{"message": "Profile data", "user_id": userID})
		})

		// Vulnerable Endpoint
		api.GET("/v1/vuln/orders/:id", vulnerableOrderHandler.GetOrder)

		// Secure Endpoint
		api.GET("/v1/secure/orders/:id", secureOrderHandler.GetOrder)
	}

	// 5. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen", slog.String("error", err.Error()))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}

	slog.Info("Server exiting")
}

func seedData(db *sql.DB) {
	// Create Tables
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL
		);
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id),
			stock_symbol VARCHAR(10) NOT NULL,
			amount INT NOT NULL,
			type VARCHAR(10) NOT NULL
		);
	`)
	if err != nil {
		slog.Error("Failed to create tables", slog.String("error", err.Error()))
		return
	}

	// Helper to hash password
	hash := func(p string) string {
		h, _ := utils.HashPassword(p)
		return h
	}

	// Seed Users: Attacker (ID 1) and Victim (ID 2)
	// Check if users exist
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if count == 0 {
		slog.Info("Seeding data...")
		// Attacker
		_, err := db.Exec("INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)", "attacker@example.com", hash("password"), "user")
		if err != nil { slog.Error("Failed to seed attacker", "error", err) }
		
		// Victim
		_, err = db.Exec("INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)", "victim@example.com", hash("password"), "user")
		if err != nil { slog.Error("Failed to seed victim", "error", err) }

		// Seed Order for Victim (ID 2). Assume IDs are serial 1 and 2.
		// Use nested query to be safe about IDs
		// Also manually insert ID 2 for victim order to be predictable for testing if needed, but serial is fine.
		_, err = db.Exec(`
			INSERT INTO orders (user_id, stock_symbol, amount, type) 
			VALUES ((SELECT id FROM users WHERE email='victim@example.com'), 'BBCA', 100, 'buy')
		`)
		if err != nil { slog.Error("Failed to seed order", "error", err) }
		
		slog.Info("Data seeded successfully")
	}
}
