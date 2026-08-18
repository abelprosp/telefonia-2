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
	"github.com/luxus-connect/telefonia/api/internal/observability"
	"github.com/luxus-connect/telefonia/api/internal/services"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
	"github.com/luxus-connect/telefonia/api/internal/storage"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"github.com/luxus-connect/telefonia/api/internal/zapsign"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Load()

	if cfg.MonitoringTestEnabled {
		logger.Error("MONITORING_TEST_ENABLED=true — API encerrada de propósito (crash loop para teste de monitoramento)")
		os.Exit(1)
	}

	if cfg.DatabaseURL == "" {
		logger.Error("DATABASE_URL or CONNECTION_STRING is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("store connect failed", "error", err)
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

	zapClient := zapsign.NewClient(zapsign.Config{
		APIToken: cfg.ZapSignAPIToken,
		BaseURL:  cfg.ZapSignBaseURL,
		Sandbox:  cfg.ZapSignSandbox,
	})

	svc := &services.Service{
		Store:        st,
		Publisher:    publisher,
		Keycloak:     kcAdmin,
		Mailer:       email.NewSender(cfg),
		Sicredi:      sicredi.NewClient(sicredi.ConfigFrom(cfg)),
		ZapSign:      zapClient,
		StateMachine: statemachine.NewEngine(st),
	}
	if zapClient.Enabled() {
		logger.Info("zapsign integration enabled")
	} else {
		logger.Info("zapsign integration ready (configure ZAPSIGN_API_TOKEN to enable electronic signing)")
	}
	if svc.Mailer.Enabled() {
		logger.Info("smtp mailer enabled", "host", cfg.SMTPHost)
	} else {
		logger.Warn("smtp mailer disabled — configure SMTP_HOST to enable billing email")
	}
	if svc.Sicredi != nil && svc.Sicredi.Enabled() {
		logger.Info("sicredi boleto integration enabled", "sandbox", cfg.SicrediSandbox)
		if cfg.SicrediAutoRegisterWebhook {
			go func() {
				time.Sleep(3 * time.Second)
				resp, err := svc.SetupSicrediProduction(context.Background(), nil)
				if err != nil {
					logger.Warn("sicredi production setup failed", "error", err)
				} else if resp != nil && resp.Success {
					logger.Info("sicredi production setup complete", "message", resp.Message)
				} else if resp != nil {
					logger.Warn("sicredi production setup incomplete", "message", resp.Message)
				}
			}()
		}
		go func() {
			ticker := time.NewTicker(15 * time.Minute)
			defer ticker.Stop()
			svc.RunSicrediPaymentSyncAllOrgs(ctx, 7)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					svc.RunSicrediPaymentSyncAllOrgs(ctx, 7)
				}
			}
		}()
	} else if cfg.SicrediEnabled {
		logger.Warn("sicredi enabled but missing credentials — set SICREDI_API_KEY, SICREDI_PASSWORD, etc.")
	}

	var presigned *services.PresignedService
	if cfg.ObjectStorageServiceURL != "" {
		s3Client, err := storage.NewClient(cfg)
		if err != nil {
			logger.Warn("object storage unavailable", "error", err)
		} else {
			presigned = &services.PresignedService{Storage: s3Client}
			processor := &importservice.Processor{Store: st, Storage: s3Client, Log: logger, SM: svc.StateMachine}
			svc.Processor = processor

			if cfg.RabbitMQURL != "" {
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

	healthChecker := observability.NewHealthChecker()
	healthChecker.Register("postgres", func(c context.Context) observability.ComponentHealth {
		start := time.Now()
		err := st.Pool().Ping(c)
		observability.Observe("db.ping", time.Since(start), err == nil)
		if err != nil {
			return observability.ComponentHealth{Status: observability.StatusDown, Message: err.Error()}
		}
		return observability.ComponentHealth{Status: observability.StatusUp}
	})
	healthChecker.Register("rabbitmq", func(_ context.Context) observability.ComponentHealth {
		if cfg.RabbitMQURL == "" {
			return observability.ComponentHealth{Status: observability.StatusDisabled, Message: "RabbitMQ not configured"}
		}
		if publisher != nil {
			return observability.ComponentHealth{Status: observability.StatusUp}
		}
		return observability.ComponentHealth{Status: observability.StatusDown, Message: "Publisher unavailable"}
	})
	healthChecker.Register("sicredi", func(c context.Context) observability.ComponentHealth {
		if !cfg.SicrediEnabled || svc.Sicredi == nil || !svc.Sicredi.Enabled() {
			return observability.ComponentHealth{Status: observability.StatusDisabled, Message: "Sicredi disabled"}
		}
		if err := svc.Sicredi.Ping(c); err != nil {
			return observability.ComponentHealth{Status: observability.StatusDegraded, Message: err.Error()}
		}
		return observability.ComponentHealth{Status: observability.StatusUp}
	})

	r := chi.NewRouter()
	r.Use(observability.CorrelationMiddleware)
	r.Use(middleware.RealIP)
	r.Use(observability.StructuredLoggerMiddleware(logger))
	r.Use(middleware.Recoverer)

	if len(cfg.CORSOrigins) > 0 {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.CORSOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Correlation-ID", "X-Request-ID"},
			ExposedHeaders:   []string{"X-Correlation-ID", "X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	r.Get("/health", healthChecker.ReadinessHandler())
	r.Get("/health/live", observability.LivenessHandler())
	r.Get("/metrics/operations", observability.MetricsHandler())

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
