package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/Ultra-Smork/Subscription-service/docs"
	"github.com/Ultra-Smork/Subscription-service/internals/config"
	"github.com/Ultra-Smork/Subscription-service/internals/handler"
	"github.com/Ultra-Smork/Subscription-service/internals/repository/postgres"
	"github.com/Ultra-Smork/Subscription-service/internals/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type App struct {
	cfg    *config.Config
	server *http.Server
	dbPool *pgxpool.Pool
}

func Run(ctx context.Context, cfg *config.Config) error {
	app := &App{cfg: cfg}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	slog.Info("initializing database", "layer", "app", "host", cfg.DBHost, "port", cfg.DBPort, "database", cfg.DBName)
	if err := app.initDB(ctx); err != nil {
		slog.Error("database initialization failed", "layer", "app", "error", err)
		return fmt.Errorf("init db: %w", err)
	}
	slog.Info("database connected", "layer", "app")
	defer app.dbPool.Close()

	repo := postgres.NewSubscriptionRepository(app.dbPool, logger)
	svc := service.NewSubscriptionService(repo, logger)
	ctrl := handler.NewSubscriptionHandler(svc, logger)
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/swagger/*", httpSwagger.Handler())

	router.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", ctrl.Create)
		r.Get("/", ctrl.List)
		r.Get("/{id}", ctrl.GetByID)
		r.Put("/{id}", ctrl.Update)
		r.Delete("/{id}", ctrl.Delete)
		r.Get("/cost", ctrl.GetTotalCost)
	})

	app.server = &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("starting http server", "layer", "app", "port", cfg.ServerPort)
		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "layer", "app", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server", "layer", "app")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}

func (a *App) initDB(ctx context.Context) error {
	slog.Debug("creating database connection pool", "layer", "app")

	pool, err := pgxpool.New(ctx, a.cfg.DSN())
	if err != nil {
		slog.Error("failed to create connection pool", "layer", "app", "error", err)
		return err
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("database ping failed", "layer", "app", "error", err)
		return err
	}
	slog.Debug("database connection pool created", "layer", "app")
	a.dbPool = pool
	return nil
}
