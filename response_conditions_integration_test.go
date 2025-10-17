package MatomoTracking

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_ResponseConditions(t *testing.T) {
	base := localMatomoURL()
	if _, ok := waitForMatomo(t, base, 5*time.Second); !ok {
		return
	}

	t.Run("tracks on 200 and header match", func(t *testing.T) {
		proxyURL, statusCh, closeProxy := startMatomoProbeProxy(t, base+"/matomo.php")
		defer closeProxy()

		cfg := &Config{
			MatomoURL: proxyURL,
			Domains: map[string]DomainConfig{
				"demo.localhost": {
					TrackingEnabled: true,
					IdSite:          1, // ensure site 1 exists in your local Matomo
					ResponseConditions: &ResponseConditions{
						TrackOnStatusCodes: []int{200},
						TrackWhenHeaders: map[string]string{
							"Content-Type": "text/html; charset=UTF-8",
						},
					},
				},
			},
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("<html>ok</html>"))
		})

		h, err := New(context.Background(), next, cfg, "test")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://demo.localhost/rc-test", nil)
		req.RemoteAddr = "203.0.113.9:54321"
		req.Header.Set("User-Agent", "RC-IT-UA")
		req.Header.Set("X-Forwarded-For", "203.0.112.8")

		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status from next: %d", rr.Code)
		}

		select {
		case code := <-statusCh:
			if code < 200 || code >= 300 {
				t.Fatalf("Matomo tracking failed: status %d", code)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("did not observe Matomo tracking request (expected due to matching conditions)")
		}
	})

	t.Run("skips on non-200", func(t *testing.T) {
		proxyURL, statusCh, closeProxy := startMatomoProbeProxy(t, base+"/matomo.php")
		defer closeProxy()

		cfg := &Config{
			MatomoURL: proxyURL,
			Domains: map[string]DomainConfig{
				"demo.localhost": {
					TrackingEnabled: true,
					IdSite:          1,
					ResponseConditions: &ResponseConditions{
						TrackOnStatusCodes: []int{200},
						TrackWhenHeaders: map[string]string{
							"Content-Type": "text/html; charset=UTF-8",
						},
					},
				},
			},
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
		})

		h, err := New(context.Background(), next, cfg, "test")
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://demo.localhost/rc-test-404", nil)
		req.RemoteAddr = "203.0.113.9:54321"
		req.Header.Set("User-Agent", "RC-IT-UA")
		req.Header.Set("X-Forwarded-For", "203.0.112.8")

		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status from next: %d", rr.Code)
		}

		select {
		case code := <-statusCh:
			t.Fatalf("unexpected Matomo call (status %d) despite non-matching conditions", code)
		case <-time.After(1 * time.Second):
			// expected: no call
			fmt.Println("no tracking call observed as expected")
		}
	})
}
