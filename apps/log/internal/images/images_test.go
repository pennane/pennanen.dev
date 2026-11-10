package images_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pennane/pennanen.dev/apps/log/internal/images"
)

func pngBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSavedImagesAreNamedByContentAndMirroredOnce(t *testing.T) {
	var mirrored []string
	s := &images.Store{Dir: t.TempDir(), Mirror: func(_ context.Context, key string, _ []byte, contentType string) error {
		mirrored = append(mirrored, key+" "+contentType)
		return nil
	}}
	body := pngBytes(t)
	first, err := s.Save(context.Background(), body)
	if err != nil {
		t.Fatal(err)
	}
	second, _ := s.Save(context.Background(), body)
	if first != second || len(first) != 68 {
		t.Errorf("names %q and %q", first, second)
	}
	if len(mirrored) != 1 || mirrored[0] != "images/"+first+" image/png" {
		t.Errorf("mirrored %v", mirrored)
	}

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/images/"+first, nil)
	req.SetPathValue("name", first)
	s.Serve(res, req)
	if res.Code != http.StatusOK || !bytes.Equal(res.Body.Bytes(), body) || res.Header().Get("Content-Type") != "image/png" {
		t.Errorf("serve: %d %s", res.Code, res.Header())
	}
}

func TestFailedMirrorKeepsNothing(t *testing.T) {
	s := &images.Store{Dir: t.TempDir(), Mirror: func(context.Context, string, []byte, string) error {
		return errors.New("r2 down")
	}}
	body := pngBytes(t)
	if _, err := s.Save(context.Background(), body); err == nil {
		t.Fatal("saved without a mirror copy")
	}
	s.Mirror = nil
	name, err := s.Save(context.Background(), body)
	if err != nil || name == "" {
		t.Errorf("retry: %q %v", name, err)
	}
}

func TestOnlyImagesAreSaved(t *testing.T) {
	s := &images.Store{Dir: t.TempDir()}
	if _, err := s.Save(context.Background(), []byte("<script>alert(1)</script>")); !errors.Is(err, images.ErrNotAnImage) {
		t.Errorf("err = %v", err)
	}
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/images/../log.db", nil)
	req.SetPathValue("name", "../log.db")
	s.Serve(res, req)
	if res.Code != http.StatusNotFound {
		t.Errorf("served %d", res.Code)
	}
}
