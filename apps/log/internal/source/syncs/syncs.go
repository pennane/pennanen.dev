package syncs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/platform/jobs"
)

type State string

const (
	NeverRan  State = "never-ran"
	Running   State = "running"
	Succeeded State = "succeeded"
	Failed    State = "failed"
)

type Report struct {
	Name            string         `json:"name"`
	State           State          `json:"state"`
	StartedAt       string         `json:"startedAt,omitempty"`
	FinishedAt      string         `json:"finishedAt,omitempty"`
	LastSucceededAt string         `json:"lastSucceededAt,omitempty"`
	Counts          map[string]int `json:"counts"`
	Problems        []string       `json:"problems"`
}

type Reports interface {
	SaveSyncReport(ctx context.Context, name string, report any) error
	SyncReport(ctx context.Context, name string, into any) (bool, error)
}

type Work func(ctx context.Context, p *Progress) error

type Sync struct {
	name     string
	work     Work
	reports  Reports
	onChange func()

	mu   sync.Mutex
	last Report
}

const interruptedByRestart = "the service stopped while this ran"

func New(ctx context.Context, name string, work Work, reports Reports, onChange func()) (*Sync, error) {
	last := Report{Name: name, State: NeverRan, Counts: map[string]int{}, Problems: []string{}}
	if _, err := reports.SyncReport(ctx, name, &last); err != nil {
		return nil, fmt.Errorf("sync %s: %w", name, err)
	}
	if last.State == Running {
		last.State = Failed
		last.Problems = append(last.Problems, interruptedByRestart)
	}
	return &Sync{name: name, work: work, reports: reports, onChange: onChange, last: last}, nil
}

func (s *Sync) Name() string { return s.name }

func (s *Sync) Last() Report {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.last
}

var ErrAlreadyRunning = errors.New("already running")

func (s *Sync) Start(ctx context.Context) error {
	started, ok := s.begin(ctx)
	if !ok {
		return ErrAlreadyRunning
	}
	go s.run(ctx, started)
	return nil
}

func (s *Sync) Run(ctx context.Context) Report {
	if started, ok := s.begin(ctx); ok {
		s.run(ctx, started)
	}
	return s.Last()
}

func (s *Sync) Job() jobs.Job {
	return jobs.Job{Name: s.name, Run: func(ctx context.Context) error {
		report := s.Run(ctx)
		if report.State == Failed {
			return errors.New(report.Problems[len(report.Problems)-1])
		}
		return nil
	}}
}

func (s *Sync) begin(ctx context.Context) (Report, bool) {
	s.mu.Lock()
	if s.last.State == Running {
		s.mu.Unlock()
		return Report{}, false
	}
	s.last = Report{
		Name:            s.name,
		State:           Running,
		StartedAt:       now(),
		LastSucceededAt: s.last.LastSucceededAt,
		Counts:          map[string]int{},
		Problems:        []string{},
	}
	started := s.last
	s.mu.Unlock()
	s.save(ctx, started)
	return started, true
}

func (s *Sync) run(ctx context.Context, started Report) {
	p := &Progress{sync: s, report: started}
	err := s.work(ctx, p)
	report := p.snapshot()
	report.FinishedAt = now()
	report.State = Succeeded
	report.LastSucceededAt = report.FinishedAt
	if err != nil {
		report.State = Failed
		report.LastSucceededAt = started.LastSucceededAt
		report.Problems = append(report.Problems, err.Error())
	}
	slog.Info("sync finished", "sync", s.name, "state", report.State, "counts", report.Counts, "problems", len(report.Problems))
	s.publish(report)
	s.save(context.WithoutCancel(ctx), report)
	if p.changedData() && s.onChange != nil {
		s.onChange()
	}
}

func (s *Sync) publish(report Report) {
	s.mu.Lock()
	s.last = report
	s.mu.Unlock()
}

func (s *Sync) save(ctx context.Context, report Report) {
	if err := s.reports.SaveSyncReport(ctx, s.name, report); err != nil {
		slog.Error("saving sync report failed", "sync", s.name, "err", err)
	}
}

type DataChange int

const (
	NoDataChanged DataChange = iota
	DataChanged
)

type Progress struct {
	sync   *Sync
	mu     sync.Mutex
	report Report
	change DataChange
}

func (p *Progress) Add(count string, n int) {
	p.update(func(r *Report) { r.Counts[count] += n })
}

func (p *Progress) Problem(format string, args ...any) {
	p.update(func(r *Report) { r.Problems = append(r.Problems, fmt.Sprintf(format, args...)) })
}

func (p *Progress) ChangedData() {
	p.mu.Lock()
	p.change = DataChanged
	p.mu.Unlock()
}

func (p *Progress) update(change func(r *Report)) {
	p.mu.Lock()
	change(&p.report)
	snapshot := p.copyOfReport()
	p.mu.Unlock()
	p.sync.publish(snapshot)
}

func (p *Progress) snapshot() Report {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.copyOfReport()
}

func (p *Progress) copyOfReport() Report {
	r := p.report
	r.Counts = maps.Clone(p.report.Counts)
	r.Problems = append([]string{}, p.report.Problems...)
	return r
}

func (p *Progress) changedData() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.change == DataChanged
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }
