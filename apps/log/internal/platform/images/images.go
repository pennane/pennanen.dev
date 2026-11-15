package images

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MaxBytes = 10 << 20

var ErrNotAnImage = errors.New("not a jpg, png, webp or gif")

var extensionOf = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/webp": "webp",
	"image/gif":  "gif",
}

var namePattern = regexp.MustCompile(`^[0-9a-f]{64}\.(jpg|png|webp|gif)$`)

type Mirror func(ctx context.Context, key string, body []byte, contentType string) error

type Store struct {
	Dir    string
	Mirror Mirror
}

func (s *Store) Save(ctx context.Context, body []byte) (string, error) {
	contentType := http.DetectContentType(body)
	extension, known := extensionOf[contentType]
	if !known {
		return "", ErrNotAnImage
	}
	sum := sha256.Sum256(body)
	name := hex.EncodeToString(sum[:]) + "." + extension
	path := filepath.Join(s.Dir, name)
	if _, err := os.Stat(path); err == nil {
		return name, nil
	}
	temporary, err := os.CreateTemp(s.Dir, "upload-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(body); err != nil {
		temporary.Close()
		return "", err
	}
	if err := temporary.Close(); err != nil {
		return "", err
	}
	if s.Mirror != nil {
		if err := s.Mirror(ctx, "images/"+name, body, contentType); err != nil {
			return "", fmt.Errorf("copy to r2: %w", err)
		}
	}
	return name, os.Rename(temporary.Name(), path)
}

func (s *Store) Serve(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !namePattern.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	for contentType, extension := range extensionOf {
		if strings.HasSuffix(name, "."+extension) {
			w.Header().Set("Content-Type", contentType)
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filepath.Join(s.Dir, name))
}
