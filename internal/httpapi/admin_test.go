package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminUIHandlerServesPageWithStrictCSP(t *testing.T) {
	handler := AdminUIHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp := request(t, server, http.MethodGet, "/admin", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	csp := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'self'") {
		t.Fatalf("expected a strict default-src in CSP, got %q", csp)
	}
	if strings.Contains(csp, "http://") || strings.Contains(csp, "https://") {
		t.Fatalf("CSP must not reference any external host, got %q", csp)
	}

	body := readAll(t, resp)
	for _, marker := range []string{"id=\"list-apps\"", "id=\"form-flag\"", "id=\"cp-status\"", "id=\"agent-status\""} {
		if !strings.Contains(body, marker) {
			t.Fatalf("expected admin page to contain %q", marker)
		}
	}
	// The page must not pull in any external script/style — only same-origin/inline.
	if strings.Contains(body, "unpkg.com") || strings.Contains(body, "cdn.") {
		t.Fatalf("admin page must not depend on a CDN")
	}
}

func readAll(t *testing.T, resp *http.Response) string {
	t.Helper()
	buf := make([]byte, 0, 64*1024)
	chunk := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(chunk)
		buf = append(buf, chunk[:n]...)
		if err != nil {
			break
		}
	}
	return string(buf)
}
