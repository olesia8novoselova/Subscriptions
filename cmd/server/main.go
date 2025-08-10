package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/olesia8novoselova/Subscriptions/internal/config"
	"github.com/olesia8novoselova/Subscriptions/internal/controller"
	"github.com/olesia8novoselova/Subscriptions/internal/repository/postgres"
	"github.com/olesia8novoselova/Subscriptions/internal/service"
	"github.com/olesia8novoselova/Subscriptions/pkg/logging"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/olesia8novoselova/Subscriptions/internal/docs"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title Subscriptions API (Swagger)
// @version  1.0
// @description REST-сервис для управления онлайн-подписками пользователей.
// @contact.name Subscriptions API
// @BasePath /
// @schemes http
// @host localhost:8080
func main() {
	_ = godotenv.Load()

	// Логгер
	logger := logging.New()
	slog.SetDefault(logger)

	// Конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return
	}

	// БД (GORM)
	db, err := initDB(cfg)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		return
	}
	sqlDB, _ := db.DB()
	defer func() {
		_ = sqlDB.Close()
		logger.Info("database connection closed")
	}()

	repo := postgres.New(db, logger)
	svc := service.NewSubscriptionService(repo, logger)
	h := controller.NewSubscriptionHandler(svc, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /api/subscriptions", h.CreateSubscription)
	mux.HandleFunc("GET /api/subscriptions/", h.GetSubscription)
	mux.HandleFunc("GET /api/subscriptions", h.ListSubscriptions)
	mux.HandleFunc("DELETE /api/subscriptions/", h.DeleteSubscription)
	mux.HandleFunc("PATCH /api/subscriptions/", h.PatchSubscription)
	mux.HandleFunc("GET /api/subscriptions/total", h.GetTotalCost)

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// Оборачиваем middleware логирования
	handler := logging.HTTPMiddleware(logger, mux)

	// HTTP Server с таймаутами
	srv := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return context.Background()
		},
	}

	logger.Info("starting server", "port", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		return
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort,
	)

	db, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
