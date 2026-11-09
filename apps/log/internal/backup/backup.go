package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type R2 struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	HTTP            *http.Client
}

type Backup struct {
	DB         *sql.DB
	ScratchDir string
	Target     R2
}

func (b *Backup) Run(ctx context.Context) (string, error) {
	snapshot := filepath.Join(b.ScratchDir, fmt.Sprintf("log-snapshot-%d.db", time.Now().UnixNano()))
	defer os.Remove(snapshot)
	if _, err := b.DB.ExecContext(ctx, `VACUUM INTO ?`, snapshot); err != nil {
		return "", fmt.Errorf("snapshot: %w", err)
	}
	compressed, err := gzipFile(snapshot)
	if err != nil {
		return "", err
	}
	key := "backups/log-" + time.Now().UTC().Format("2006-01-02T150405Z") + ".db.gz"
	return key, b.Target.put(ctx, key, compressed)
}

func gzipFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := io.Copy(zw, f); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r R2) put(ctx context.Context, key string, body []byte) error {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", r.AccountID, r.Bucket, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/gzip")
	sum := sha256.Sum256(body)
	Sign(req, hex.EncodeToString(sum[:]), Credentials{
		AccessKeyID:     r.AccessKeyID,
		SecretAccessKey: r.SecretAccessKey,
		Region:          "auto",
		Service:         "s3",
	}, time.Now())
	client := r.HTTP
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("upload: status %d: %s", res.StatusCode, detail)
	}
	return nil
}
