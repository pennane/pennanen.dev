package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/api"
	"github.com/pennane/pennanen.dev/apps/log/internal/backup"
	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/igdb"
	"github.com/pennane/pennanen.dev/apps/log/internal/images"
	"github.com/pennane/pennanen.dev/apps/log/internal/jobs"
	"github.com/pennane/pennanen.dev/apps/log/internal/revalidate"
	"github.com/pennane/pennanen.dev/apps/log/internal/steam"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("log stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPath := env("LOG_DB", "log.db")
	database, err := db.Open(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()
	st := store.New(database)

	notifier := &revalidate.Notifier{
		URL:      os.Getenv("REVALIDATE_URL"),
		Secret:   os.Getenv("REVALIDATE_SECRET"),
		Tags:     []string{"log"},
		Debounce: 5 * time.Second,
	}

	imageDir := env("LOG_IMAGES", filepath.Join(filepath.Dir(dbPath), "images"))
	if err := os.MkdirAll(imageDir, 0o755); err != nil {
		return fmt.Errorf("image directory: %w", err)
	}
	pictures := &images.Store{Dir: imageDir}

	server := &api.Server{Store: st, AdminToken: os.Getenv("LOG_ADMIN_TOKEN"), Changed: notifier.Changed, Images: pictures}
	if server.AdminToken == "" {
		slog.Warn("LOG_ADMIN_TOKEN is empty; admin endpoints will refuse every request")
	}

	var daily []jobs.Job
	if key, id := os.Getenv("STEAM_API_KEY"), os.Getenv("STEAM_ID"); key != "" && id != "" {
		server.Steam = steam.NewSyncer(st, steam.NewClient(key, id), notifier.Changed)
		daily = append(daily, jobs.Job{Name: "steam-sync", Run: func(ctx context.Context) error {
			if report := server.Steam.Run(ctx); report.State == steam.SyncFailed {
				return errors.New(report.Problems[len(report.Problems)-1])
			}
			return nil
		}})
	} else {
		slog.Warn("STEAM_API_KEY or STEAM_ID missing; steam sync disabled")
	}

	if id, secret := os.Getenv("IGDB_CLIENT_ID"), os.Getenv("IGDB_CLIENT_SECRET"); id != "" && secret != "" {
		server.IGDB = igdb.NewClient(id, secret)
		server.IGDBMatch = igdb.NewMatcher(st, server.IGDB, notifier.Changed)
		daily = append(daily, jobs.Job{Name: "igdb-match", Run: func(ctx context.Context) error {
			if report := server.IGDBMatch.Run(ctx); report.State == igdb.MatchFailed {
				return errors.New(report.Problems[len(report.Problems)-1])
			}
			return nil
		}})
	} else {
		slog.Warn("IGDB_CLIENT_ID or IGDB_CLIENT_SECRET missing; igdb disabled")
	}

	if target, configured := r2FromEnv(); configured {
		b := &backup.Backup{DB: database, ScratchDir: filepath.Dir(dbPath), Target: target}
		server.Backup = b.Run
		pictures.Mirror = target.Put
		daily = append(daily, jobs.Job{Name: "backup", Run: func(ctx context.Context) error {
			key, err := b.Run(ctx)
			if err == nil {
				slog.Info("backup uploaded", "key", key)
			}
			return err
		}})
	} else {
		slog.Warn("R2 settings missing; backups disabled")
	}

	go jobs.RunDaily(ctx, 4, 0, daily)

	httpServer := &http.Server{
		Addr:              env("LOG_ADDR", "127.0.0.1:3005"),
		Handler:           server.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdown)
	}()
	slog.Info("log listening", "addr", httpServer.Addr, "db", dbPath)
	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func r2FromEnv() (backup.R2, bool) {
	r := backup.R2{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		Bucket:          os.Getenv("R2_BUCKET"),
	}
	return r, r.AccountID != "" && r.AccessKeyID != "" && r.SecretAccessKey != "" && r.Bucket != ""
}

func env(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
