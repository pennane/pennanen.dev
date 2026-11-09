package revalidate

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Notifier struct {
	URL      string
	Secret   string
	Tags     []string
	Debounce time.Duration
	HTTP     *http.Client

	mu      sync.Mutex
	pending *time.Timer
}

func (n *Notifier) Changed() {
	if n == nil || n.URL == "" {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.pending != nil {
		n.pending.Stop()
	}
	n.pending = time.AfterFunc(n.Debounce, n.send)
}

func (n *Notifier) send() {
	body, _ := json.Marshal(map[string]any{"tags": n.Tags})
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL, bytes.NewReader(body))
	if err != nil {
		slog.Error("revalidate request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.Secret)
	client := n.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		slog.Error("revalidate", "err", err)
		return
	}
	res.Body.Close()
	if res.StatusCode/100 != 2 {
		slog.Error("revalidate", "status", res.StatusCode)
	}
}
