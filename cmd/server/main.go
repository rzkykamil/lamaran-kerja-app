package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"

	lamarankerja "lamarankerja"
	"lamarankerja/internal/config"
	"lamarankerja/internal/database"
	"lamarankerja/internal/store"
	"lamarankerja/internal/web"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if err := run(logger); err != nil {
		logger.Fatalf("aplikasi berhenti: %v", err)
	}
}

func run(logger *log.Logger) error {
	cfg, err := config.Load(".env")
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}

	st := store.New(pool)

	if n, err := st.JumlahUser(ctx); err == nil && n == 0 {
		logger.Println("PERHATIAN: belum ada akun. Jalankan: go run ./cmd/seedadmin")
	}

	sessions := scs.New()
	sessions.Store = pgxstore.New(pool)
	sessions.Lifetime = 7 * 24 * time.Hour
	sessions.Cookie.HttpOnly = true
	sessions.Cookie.SameSite = http.SameSiteLaxMode

	app, err := web.NewApp(st, sessions, lamarankerja.Assets, logger)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      app.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	matiSignal := make(chan os.Signal, 1)
	signal.Notify(matiSignal, os.Interrupt, syscall.SIGTERM)
	errChan := make(chan error, 1)

	go func() {
		logger.Printf("server jalan di http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-matiSignal:
		logger.Println("menutup server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
