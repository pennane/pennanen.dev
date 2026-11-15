package syncs_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/db"
	"github.com/pennane/pennanen.dev/apps/log/internal/store"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
)

func openStore(t *testing.T) *store.Store {
	t.Helper()
	d, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "log.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return store.Open(d)
}

func TestReportsSurviveARestartAndFailuresKeepTheLastSuccess(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	outcome := error(nil)
	work := func(ctx context.Context, p *syncs.Progress) error {
		p.Add("seen", 3)
		p.Problem("one %s was odd", "thing")
		return outcome
	}

	first, err := syncs.New(ctx, "fake", work, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Last().State != syncs.NeverRan {
		t.Fatalf("fresh: %+v", first.Last())
	}
	succeeded := first.Run(ctx)
	if succeeded.State != syncs.Succeeded || succeeded.Counts["seen"] != 3 || len(succeeded.Problems) != 1 || succeeded.LastSucceededAt == "" {
		t.Fatalf("success: %+v", succeeded)
	}

	outcome = errors.New("site changed its markup")
	restarted, err := syncs.New(ctx, "fake", work, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Last().State != syncs.Succeeded {
		t.Fatalf("report lost on restart: %+v", restarted.Last())
	}
	failed := restarted.Run(ctx)
	if failed.State != syncs.Failed || failed.LastSucceededAt != succeeded.LastSucceededAt || failed.Problems[len(failed.Problems)-1] != "site changed its markup" {
		t.Fatalf("failure: %+v", failed)
	}
	if err := restarted.Job().Run(ctx); err == nil {
		t.Error("a failed sync is a failed job")
	}
}

func TestASyncCutShortByARestartShowsAsFailed(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	if err := st.SaveSyncReport(ctx, "fake", syncs.Report{Name: "fake", State: syncs.Running, Counts: map[string]int{}, Problems: []string{}}); err != nil {
		t.Fatal(err)
	}
	s, err := syncs.New(ctx, "fake", func(context.Context, *syncs.Progress) error { return nil }, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	if last := s.Last(); last.State != syncs.Failed || len(last.Problems) != 1 {
		t.Fatalf("interrupted: %+v", last)
	}
}

func TestOneRunAtATimeAndChangesAreAnnounced(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	release := make(chan struct{})
	done := make(chan struct{})
	announced := 0
	s, err := syncs.New(ctx, "fake", func(ctx context.Context, p *syncs.Progress) error {
		<-release
		p.ChangedData()
		return nil
	}, st, func() { announced++; close(done) })
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.Start(ctx); !errors.Is(err, syncs.ErrAlreadyRunning) {
		t.Errorf("second start: %v", err)
	}
	close(release)
	<-done
	if announced != 1 {
		t.Errorf("announced %d", announced)
	}
}
