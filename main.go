package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"merceria/internal/auth"
	"merceria/internal/config"
	handler "merceria/internal/handler"
	"merceria/internal/middleware"
	"merceria/internal/spreadsheets"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-quit:
			log.Println("Received shutdown signal")
			cancel()
		case <-ctx.Done():
		}
		return nil
	})

	grp.Go(func() error {
		return Run(ctx, grp)
	})

	err := grp.Wait()
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
	log.Println("Goodbye!")
}

func Run(ctx context.Context, grp *errgroup.Group) error {
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	rauth := auth.NewAuthorizerFactory(cfg)
	spreadsheets, err := spreadsheets.New(ctx, cfg.ServiceAccount)
	if err != nil {
		return fmt.Errorf("failed to create spreadsheet operator: %w", err)
	}

	static, err := os.OpenRoot("./static")
	if err != nil {
		return fmt.Errorf("opening static files: %w", err)
	}

	mux := http.NewServeMux()
	mw := middleware.From(middleware.Logging, middleware.Recover(), middleware.CORS(cfg.CORSOrigins))
	mux.Handle("/~/", middleware.Apply(http.FileServerFS(static.FS()), mw, middleware.StripPrefix("/~/")))

	mw = middleware.From(mw, middleware.RateLimit(2, 8))
	mux.Handle("/", middleware.Apply(http.RedirectHandler("/form", http.StatusFound), mw))
	mux.Handle("GET /health", middleware.ApplyFunc(handler.Health, mw))
	mux.Handle("GET /auth/logout", middleware.ApplyFunc(handler.LogoutHandler(ctx, static), mw))
	mux.Handle("GET /auth/google/login", middleware.ApplyFunc(handler.LoginHandler(rauth), mw))
	mux.Handle("GET /auth/google/callback", middleware.ApplyFunc(handler.LoginCallbackHandler(rauth), mw))
	if cfg.Development {
		mux.Handle("GET /auth/dev/login", middleware.ApplyFunc(handler.DevLoginHandler(rauth), mw))
	}

	mw = middleware.From(mw, middleware.Auth(rauth))
	mux.Handle("GET /pick", middleware.ApplyFunc(handler.PickHandler(ctx, rauth, static, cfg.GoogleAPIKey), mw))
	mux.Handle("POST /pick", middleware.Apply(handler.PickCallback(ctx, rauth), mw))

	mw = middleware.From(mw, middleware.WithSpreadsheetId(rauth))
	mux.Handle("GET /form", middleware.Apply(handler.CreateRowForm(ctx, static), mw))
	mux.Handle("POST /form", middleware.Apply(handler.CreateRow(spreadsheets), mw))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	grp.Go(func() error {
		log.Printf("Starting server on port %s", cfg.Port)
		if cfg.Development {
			log.Println("WARNING: Running in development mode. Do not use this in production!")
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	})

	grp.Go(func() error {
		<-ctx.Done()
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server forced to shutdown: %w", err)
		}
		return nil
	})

	return nil
}
