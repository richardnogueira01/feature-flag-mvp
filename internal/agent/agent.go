package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/richardnogueira01/feature-flag-mvp/internal/snapshot"
)

type Agent struct {
	upstream string
	client   *http.Client
	store    *snapshot.Store
	etag     atomic.Value
}

type wireFlag struct {
	Key      string          `json:"key"`
	Enabled  bool            `json:"enabled"`
	Value    json.RawMessage `json:"value"`
	Revision uint64          `json:"revision"`
}
type wireResponse struct {
	Flags []wireFlag `json:"flags"`
}

func New(upstream string) *Agent {
	return &Agent{upstream: strings.TrimRight(upstream, "/"), client: &http.Client{Timeout: 10 * time.Second}, store: snapshot.NewStore(nil)}
}
func (a *Agent) Store() *snapshot.Store { return a.store }
func (a *Agent) Poll(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.upstream+"/v1/flags", nil)
	if err != nil {
		return err
	}
	if v, ok := a.etag.Load().(string); ok && v != "" {
		req.Header.Set("If-None-Match", v)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotModified {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream returned %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var payload wireResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	flags := make(map[string]snapshot.Flag, len(payload.Flags))
	var revision uint64
	for _, f := range payload.Flags {
		flags[f.Key] = snapshot.Flag{Key: f.Key, Enabled: f.Enabled, Value: append(json.RawMessage(nil), f.Value...)}
		if f.Revision > revision {
			revision = f.Revision
		}
	}
	a.store.Publish(&snapshot.Snapshot{Revision: revision, Flags: flags})
	if v := resp.Header.Get("ETag"); v != "" {
		a.etag.Store(v)
	}
	return nil
}
func (a *Agent) Run(ctx context.Context, interval time.Duration) {
	for {
		if err := a.Poll(ctx); err != nil {
			log.Printf("agent poll failed; serving last snapshot: %v", err)
		} else {
			log.Printf("agent snapshot revision=%d", a.store.Revision())
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
func ParseDuration(value string, fallback time.Duration) time.Duration {
	if value == "" {
		return fallback
	}
	if d, err := time.ParseDuration(value); err == nil && d > 0 {
		return d
	}
	return fallback
}
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func ParsePort(value string, fallback string) string {
	if _, err := strconv.Atoi(value); err == nil {
		return ":" + value
	}
	return fallback
}
