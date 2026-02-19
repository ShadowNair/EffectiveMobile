package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test_task/internal/config"
	"test_task/internal/connections"

	handlersub "test_task/internal/handlers/subscription"
	logmid "test_task/internal/middleware/loger_middleware"
	reposub "test_task/internal/repository/postgres/subscription"
	usecasesub "test_task/internal/usecase/subscription"
)

func main() {
	cfg := config.GetConfig()

	log := logmid.NewLogger(os.Getenv("LOG_LEVEL"))

	conn, err := connections.New(cfg)
	if err != nil {
		log.Error("connections init failed", slog.Any("err", err))
		os.Exit(1)
	}

	repo := reposub.New(conn.PostgresSQL)
	usecase := usecasesub.New(repo)
	handler := handlersub.New(log, usecase)
	router := handlersub.Router(log, handler)

	addr := net.JoinHostPort(cfg.AppConfig.Host, cfg.AppConfig.Port)
	if cfg.AppConfig.Host == "" || cfg.AppConfig.Port == "" {
		addr = "0.0.0.0:8080"
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("server started",
			slog.String("addr", srv.Addr),
			slog.String("swagger", "http://"+srv.Addr+"/swagger/"),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.Any("err", err))
			serverErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-stop:
		log.Info("received shutdown signal")
	case err := <-serverErrors:
		log.Error("server failed", slog.Any("err", err))
	}

	log.Info("shutting down")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error("shutdown error", slog.Any("err", err))
		cancel()
		conn.CloseAll()
		os.Exit(1)
	}
	cancel()

	log.Info("bye")
}
