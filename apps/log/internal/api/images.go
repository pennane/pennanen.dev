package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pennane/pennanen.dev/apps/log/internal/images"
)

const uploadTicketLife = 10 * time.Minute

type UploadTicket struct {
	Ticket string `json:"ticket"`
	Path   string `json:"path"`
}

type UploadedImage struct {
	Name string `json:"name"`
}

func (s *Server) uploadTicket(w http.ResponseWriter, r *http.Request) {
	if s.Images == nil {
		writeError(w, http.StatusServiceUnavailable, "images are not configured")
		return
	}
	expires := time.Now().Add(uploadTicketLife).Unix()
	writeJSON(w, http.StatusOK, UploadTicket{Ticket: s.ticketFor(expires), Path: "/images"})
}

func (s *Server) ticketFor(expires int64) string {
	mac := hmac.New(sha256.New, []byte(s.AdminToken))
	mac.Write([]byte("upload-image:" + strconv.FormatInt(expires, 10)))
	return strconv.FormatInt(expires, 10) + "." + hex.EncodeToString(mac.Sum(nil))
}

func (s *Server) admitsTicket(presented string, now time.Time) bool {
	expiresText, _, found := strings.Cut(presented, ".")
	expires, err := strconv.ParseInt(expiresText, 10, 64)
	if !found || err != nil || s.AdminToken == "" || now.Unix() > expires {
		return false
	}
	return hmac.Equal([]byte(presented), []byte(s.ticketFor(expires)))
}

func allowUploadsFromAnyOrigin(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
}

func (s *Server) uploadPreflight(w http.ResponseWriter, r *http.Request) {
	allowUploadsFromAnyOrigin(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) uploadImage(w http.ResponseWriter, r *http.Request) {
	allowUploadsFromAnyOrigin(w)
	if s.Images == nil {
		writeError(w, http.StatusServiceUnavailable, "images are not configured")
		return
	}
	if !s.admitsTicket(strings.TrimPrefix(r.Header.Get("Authorization"), "Ticket "), time.Now()) {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, images.MaxBytes))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, "images can be at most 10 MB")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	name, err := s.Images.Save(r.Context(), body)
	if errors.Is(err, images.ErrNotAnImage) {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if err != nil {
		respond(w, 0, nil, err)
		return
	}
	writeJSON(w, http.StatusCreated, UploadedImage{Name: name})
}

func (s *Server) serveImage(w http.ResponseWriter, r *http.Request) {
	if s.Images == nil {
		http.NotFound(w, r)
		return
	}
	s.Images.Serve(w, r)
}
