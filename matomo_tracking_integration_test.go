package MatomoTracking

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

func localMatomoURL() string {
	if v := os.Getenv("MATOMO_URL"); v != "" {
		return v
	}
	// Default to your compose port mapping
	return "http://localhost:8082"
}

func waitForMatomo(t *testing.T, base string, timeout time.Duration) (string, bool) {
	t.Helper()
	u, err := url.Parse(base)
	if err != nil {
		t.Skipf("Invalid MATOMO_URL %q: %v", base, err)
		return "", false
	}
	// Probe tracking endpoint
	u.Path = "/matomo.php"
	q := u.Query()
	// q.Set("ping", "1")
	u.RawQuery = q.Encode()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(u.String())
		if err == nil {
			_ = resp.Body.Close()
			// Accept common statuses from the tracker
			if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
				return base, true
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("Local Matomo not reachable at %s within %s", u.String(), timeout)
	return "", false
}

func TestIntegration_LocalMatomo_DirectProbe(t *testing.T) {
	base := localMatomoURL()
	if _, ok := waitForMatomo(t, base, 30*time.Second); !ok {
		return
	}
	// If we got here, Matomo is up enough to accept tracker requests.
}

func TestIntegration_LocalMatomo_MiddlewareSendsRequest(t *testing.T) {
	base := localMatomoURL()
	if _, ok := waitForMatomo(t, base, 5*time.Second); !ok {
		return
	}

	cfg := &Config{
		MatomoURL: base + "/matomo.php",
		Domains: map[string]DomainConfig{
			"demo.localhost": {
				TrackingEnabled: true,
				IdSite:          2, // assumes site id 1 exists after setup
			},
		},
	}

	// Next handler to verify normal flow
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	h, err := New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://demo.localhost/integration/test", nil)
	req.RemoteAddr = "203.0.113.9:54321"
	req.Header.Set("User-Agent", "Integration-Test-UA")
	req.Header.Set("X-Forwarded-For", "203.0.112.8")

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%q", rr.Code, rr.Body.String())
	}

	// Give the async tracker a moment (best-effort)
	time.Sleep(300 * time.Millisecond)
}
