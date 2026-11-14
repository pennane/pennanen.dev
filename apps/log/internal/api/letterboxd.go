package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"slices"

	"github.com/pennane/pennanen.dev/apps/log/internal/letterboxd"
	"github.com/pennane/pennanen.dev/apps/log/internal/syncs"
	"github.com/pennane/pennanen.dev/apps/log/internal/tmdb"
)

const (
	LetterboxdExportSyncName = "letterboxd-export"
	maxLetterboxdExportBytes = 32 << 20
)

func (s *Server) importLetterboxdExport(w http.ResponseWriter, r *http.Request) {
	if s.Letterboxd == nil {
		writeError(w, http.StatusNotImplemented, "letterboxd is not configured")
		return
	}
	archive, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxLetterboxdExportBytes))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "the export is larger than 32 MB")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	export, err := letterboxd.ParseExport(archive)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	run, err := syncs.New(r.Context(), LetterboxdExportSyncName, func(ctx context.Context, p *syncs.Progress) error {
		return s.Letterboxd.Import(ctx, p, export)
	}, s.Store, s.changed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	report := run.Run(context.WithoutCancel(r.Context()))
	s.startTMDBMatch(r.Context())
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) startTMDBMatch(ctx context.Context) {
	i := slices.IndexFunc(s.Syncs, func(sync *syncs.Sync) bool { return sync.Name() == tmdb.SyncName })
	if i >= 0 {
		s.Syncs[i].Start(context.WithoutCancel(ctx))
	}
}
