package MatomoTracking

import (
	"context"
	"fmt"
	"io"
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
		t.Fatalf("Invalid MATOMO_URL %q: %v", base, err)
		return "", false
	}
	u.Path = "/matomo.php"
	q := u.Query()
	u.RawQuery = q.Encode()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(u.String())
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
				return base, true
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Local Matomo not reachable at %s within %s", u.String(), timeout)
	return "", false
}

// startMatomoProbeProxy forwards requests to target (e.g., http://127.0.0.1:8082/matomo.php)
// and sends the upstream Matomo status code on statusCh.
func startMatomoProbeProxy(t *testing.T, target string) (proxyURL string, statusCh <-chan int, closeFn func()) {
	t.Helper()

	ch := make(chan int, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Forward to real Matomo
		forwardURL := target
		if rq := r.URL.RawQuery; rq != "" {
			forwardURL += "?" + rq
		}
		req, _ := http.NewRequestWithContext(r.Context(), r.Method, forwardURL, nil)
		req.Header = r.Header.Clone()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// Treat as 599 on network error
			select {
			case ch <- 599:
			default:
			}
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Report status back to test
		select {
		case ch <- resp.StatusCode:
		default:
		}

		// Mirror upstream status/body (body usually empty for matomo.php)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(io.Discard, resp.Body)
	}))

	return srv.URL, ch, func() { srv.Close() }
}

func TestIntegration_LocalMatomo_DirectProbe(t *testing.T) {
	base := localMatomoURL()
	if _, ok := waitForMatomo(t, base, 5*time.Second); !ok {
		return
	}
	// If we got here, Matomo is up enough to accept tracker requests.
}

func TestIntegration_LocalMatomo_MiddlewareSendsRequest(t *testing.T) {
	base := localMatomoURL()
	if _, ok := waitForMatomo(t, base, 5*time.Second); !ok {
		return
	}

	// Route middleware traffic through a lightweight in-memory proxy.
	// This proxy does *not* send any requests by itself — it just waits for the middleware
	// to make a tracking call. When that happens, the proxy forwards the request once
	// to the real Matomo instance and reports Matomo’s HTTP status code back on a channel.
	// This lets the test verify that a real tracking hit occurred and was accepted
	// without modifying or double-sending the request.
	proxyURL, statusCh, closeProxy := startMatomoProbeProxy(t, base+"/matomo.php")
	defer closeProxy()

	cfg := &Config{
		MatomoURL: proxyURL, // middleware appends query to this URL
		Domains: map[string]DomainConfig{
			"demo.localhost": {
				TrackingEnabled: true,
				IdSite:          1, // ensure this site exists or expect failure
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

	// Assert the tracking call happened and Matomo accepted it
	select {
	case code := <-statusCh:
		if code < 200 || code >= 300 {
			t.Fatalf("Matomo tracking request failed: status %d (check idsite and config)", code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("did not observe Matomo tracking request")
	}
}
