package jobs

import (
	"context"
	"log/slog"
	"time"
)

type Job struct {
	Name string
	Run  func(ctx context.Context) error
}

func RunDaily(ctx context.Context, hour, minute int, jobs []Job) {
	for {
		wait := time.Until(nextOccurrence(time.Now(), hour, minute))
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		for _, job := range jobs {
			started := time.Now()
			if err := job.Run(ctx); err != nil {
				slog.Error("daily job failed", "job", job.Name, "err", err)
				continue
			}
			slog.Info("daily job done", "job", job.Name, "took", time.Since(started).Round(time.Second))
		}
	}
}

func nextOccurrence(from time.Time, hour, minute int) time.Time {
	next := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, from.Location())
	if !next.After(from) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
