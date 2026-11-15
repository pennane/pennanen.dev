package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pennane/pennanen.dev/apps/log/internal/source/spotify"
)

type spotifyConnection string

const spotifyNotConfigured spotifyConnection = "not-configured"

type spotifyConnectRequest struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirectUri"`
}

func (s *Server) spotifyConfigured(w http.ResponseWriter) bool {
	if s.Spotify == nil {
		writeError(w, http.StatusNotImplemented, "spotify is not configured")
		return false
	}
	return true
}

func (s *Server) spotifyConnection(w http.ResponseWriter, r *http.Request) {
	if s.Spotify == nil {
		writeJSON(w, http.StatusOK, map[string]spotifyConnection{"connection": spotifyNotConfigured})
		return
	}
	connection, err := s.Spotify.Connection(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"connection": string(connection)})
}

func (s *Server) spotifyAuthorize(w http.ResponseWriter, r *http.Request) {
	if !s.spotifyConfigured(w) {
		return
	}
	redirectURI := strings.TrimSpace(r.URL.Query().Get("redirect_uri"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if redirectURI == "" || state == "" {
		writeError(w, http.StatusBadRequest, "redirect_uri and state are required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": s.Spotify.AuthorizeURL(redirectURI, state)})
}

func (s *Server) spotifyConnect(w http.ResponseWriter, r *http.Request) {
	if !s.spotifyConfigured(w) {
		return
	}
	var request spotifyConnectRequest
	if !readJSON(w, r, &request) {
		return
	}
	if request.Code == "" || request.RedirectURI == "" {
		writeError(w, http.StatusBadRequest, "code and redirectUri are required")
		return
	}
	if err := s.Spotify.Connect(r.Context(), request.Code, request.RedirectURI); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, spotify.ErrCodeRefused) {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
