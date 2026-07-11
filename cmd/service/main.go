package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/peymanahmadi/payment-risk-guard/internal/adapter/httpapi"
	"github.com/peymanahmadi/payment-risk-guard/internal/adapter/postgres"
	"github.com/peymanahmadi/payment-risk-guard/internal/config"
	"github.com/peymanahmadi/payment-risk-guard/internal/risk"
	"github.com/peymanahmadi/payment-risk-guard/internal/usecase"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Open(ctx, postgres.Config{
		DSN:             cfg.PostgresDSN,
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	})
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := postgres.Migrate(ctx, db); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	txRepo := postgres.NewTransactionRepository(db)
	alertRepo := postgres.NewAlertRepository(db)

	engine := risk.NewEngine(
		cfg.RiskThreshold,
		risk.NewVelocityRule(txRepo, cfg.VelocityWindow, cfg.VelocityMaxAllowed),
		risk.NewAmountSpikeRule(txRepo, cfg.AmountSpikeLookback, cfg.AmountSpikeMultiplier, cfg.AmountSpikeMinHistory),
		risk.NewGeoMismatchRule(txRepo, cfg.GeoLookbackWindow, cfg.GeoMinGap),
	)
	process := usecase.NewProcessTransaction(txRepo, alertRepo, engine, log)

	handler := httpapi.NewHandler(process, alertRepo, txRepo)
	router := httpapi.NewRouter(handler, log)
	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		log.Error("component failed, shutting down", "error", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown error", "error", err)
	}
}
