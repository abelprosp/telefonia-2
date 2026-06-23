package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/email"
	"github.com/luxus-connect/telefonia/api/internal/handlers"
	"github.com/luxus-connect/telefonia/api/internal/importservice"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/messaging"
	"github.com/luxus-connect/telefonia/api/internal/services"
	"github.com/luxus-connect/telefonia/api/internal/storage"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL or CONNECTION_STRING is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer st.Close()

	kcAdmin := keycloak.NewAdminClient(cfg)

	authMW, err := auth.NewMiddleware(cfg, logger, kcAdmin)
	if err != nil {
		logger.Error("auth middleware init failed", "error", err)
		os.Exit(1)
	}
	authMW.StartJWKSRefresh(ctx)

	var publisher services.EventPublisher
	if cfg.RabbitMQURL != "" {
		pub, err := messaging.NewPublisher(cfg.RabbitMQURL, logger)
		if err != nil {
			logger.Warn("rabbitmq publisher unavailable", "error", err)
		} else {
			publisher = pub
			defer pub.Close()
		}
	}

	svc := &services.Service{Store: st, Publisher: publisher, Keycloak: kcAdmin, Mailer: email.NewSender(cfg)}
	if svc.Mailer.Enabled() {
		logger.Info("smtp mailer enabled", "host", cfg.SMTPHost)
	} else {
		logger.Warn("smtp mailer disabled — configure SMTP_HOST to enable billing email")
	}

	var presigned *services.PresignedService
	if cfg.ObjectStorageServiceURL != "" {
		s3Client, err := storage.NewClient(cfg)
		if err != nil {
			logger.Warn("object storage unavailable", "error", err)
		} else {
			presigned = &services.PresignedService{Storage: s3Client}

			if cfg.RabbitMQURL != "" {
				processor := &importservice.Processor{Store: st, Storage: s3Client, Log: logger}
				consumer, err := messaging.NewConsumer(cfg.RabbitMQURL, processor, logger)
				if err != nil {
					logger.Warn("rabbitmq consumer unavailable", "error", err)
				} else {
					defer consumer.Close()
					if err := consumer.Start(ctx); err != nil {
						logger.Warn("rabbitmq consumer start failed", "error", err)
					} else {
						logger.Info("rabbitmq consumer started", "queue", messaging.QueueName)
					}
				}
			}
		}
	}
	if presigned == nil {
		presigned = &services.PresignedService{}
	}

	h := &handlers.Handler{Svc: svc, Presigned: presigned}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	if len(cfg.CORSOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.CORSOrigins,
			AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	h.RegisterRoutes(r, authMW.Authenticate, authMW.RequireOperational, authMW.RequireFinancialAccess, authMW.RequireMaster, authMW.RequirePartner)

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{Addr: addr, Handler: r, ReadHeaderTimeout: 10 * time.Second}

	go func() {
		logger.Info("server starting", "addr", addr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
	logger.Info("server stopped")
}
