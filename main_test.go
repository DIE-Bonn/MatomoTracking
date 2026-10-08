package MatomoTracking

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

func TestPathMatchesPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path   string
		prefix string
		want   bool
	}{
		{"/test", "/test", true},
		{"/test/sub", "/test", true},
		{"/test2", "/test", false},
		{"/testing", "/test", false},
		{"/", "/", true},
		{"/a/b", "/a", true},
		{"/a", "/a/b", false},
	}

	for _, tt := range tests {
		if got := pathMatchesPrefix(tt.path, tt.prefix); got != tt.want {
			t.Fatalf("pathMatchesPrefix(%q, %q) = %v; want %v", tt.path, tt.prefix, got, tt.want)
		}
	}
}

func TestMergeConfigs(t *testing.T) {
	t.Parallel()

	base := DomainConfig{
		TrackingEnabled: true,
		IdSite:          1,
		ExcludedPaths:   []string{"/admin"},
		IncludedPaths:   []string{`\.php$`},
	}

	override := PathConfig{
		TrackingEnabled: boolPtr(false),
		IdSite:          intPtr(42),
		ExcludedPaths:   []string{"/private"},
		IncludedPaths:   []string{`\.aspx$`},
	}

	got := mergeConfigs(base, override)

	if got.TrackingEnabled != false {
		t.Fatalf("TrackingEnabled = %v; want false", got.TrackingEnabled)
	}
	if got.IdSite != 42 {
		t.Fatalf("IdSite = %d; want 42", got.IdSite)
	}
	if len(got.ExcludedPaths) != 1 || got.ExcludedPaths[0] != "/private" {
		t.Fatalf("ExcludedPaths = %#v; want [/private]", got.ExcludedPaths)
	}
	if len(got.IncludedPaths) != 1 || got.IncludedPaths[0] != `\.aspx$` {
		t.Fatalf("IncludedPaths = %#v; want [\\.aspx$]", got.IncludedPaths)
	}
}

func TestIsPathExcluded(t *testing.T) {
	t.Parallel()

	// Matches files with extensions (e.g., .css, .png) optionally followed by query
	excluded := []string{`\.\w{1,5}(\?.+)?$`}
	// Includes PHP and ASPX explicitly even if excluded matched
	included := []string{`\.php(\?.*)?$`, `\.aspx(\?.*)?$`}

	cases := []struct {
		path string
		want bool
	}{
		{"/index.php", false},      // included explicitly
		{"/INDEX.PHP", true},       // included regex patterns are case-sensitive -> not included, so excluded by ext
		{"/api/data", false},       // no excluded match
		{"/image.png", true},       // excluded by ext
		{"/style.CSS", true},       // excluded by ext
		{"/page.aspx?x=1", false},  // included explicitly
		{"/download.tar.gz", true}, // .gz matches (3 letters)
		{"/noext", false},          // no excluded match
	}

	for _, tc := range cases {
		got := isPathExcluded(tc.path, excluded, included)
		if got != tc.want {
			t.Fatalf("isPathExcluded(%q) = %v; want %v", tc.path, got, tc.want)
		}
	}
}

func TestTrackedMethodsConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config string
		method string
		want   bool
	}{
		{"omitted allows GET", `{}`, http.MethodGet, true},
		{"omitted allows HEAD", `{}`, http.MethodHead, true},
		{"omitted allows OPTIONS", `{}`, http.MethodOptions, true},
		{"listed GET", `{"trackedMethods":["GET"]}`, http.MethodGet, true},
		{"unlisted HEAD", `{"trackedMethods":["GET"]}`, http.MethodHead, false},
		{"unlisted OPTIONS", `{"trackedMethods":["GET"]}`, http.MethodOptions, false},
		{"multiple methods", `{"trackedMethods":["GET","POST"]}`, http.MethodPost, true},
		{"case insensitive", `{"trackedMethods":["post"]}`, http.MethodPost, true},
		{"empty allows none", `{"trackedMethods":[]}`, http.MethodGet, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config DomainConfig
			if err := json.Unmarshal([]byte(tt.config), &config); err != nil {
				t.Fatal(err)
			}
			if got := isMethodTracked(tt.method, config.TrackedMethods); got != tt.want {
				t.Fatalf("isMethodTracked(%q, %#v) = %v; want %v", tt.method, config.TrackedMethods, got, tt.want)
			}
		})
	}
}

func TestMergeConfigsTrackedMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		base     []string
		override []string
		want     []string
	}{
		{"omitted everywhere", nil, nil, nil},
		{"inherit domain list", []string{"GET"}, nil, []string{"GET"}},
		{"replace domain list", []string{"GET"}, []string{"POST"}, []string{"POST"}},
		{"empty override tracks none", []string{"GET"}, []string{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeConfigs(DomainConfig{TrackedMethods: tt.base}, PathConfig{TrackedMethods: tt.override})
			if !reflect.DeepEqual(got.TrackedMethods, tt.want) {
				t.Fatalf("TrackedMethods = %#v; want %#v", got.TrackedMethods, tt.want)
			}
		})
	}
}

func TestServeHTTPTrackedMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		trackedMethods []string
		override       *PathConfig
		wantTracking   bool
	}{
		{"omitted tracks HEAD", http.MethodHead, nil, nil, true},
		{"omitted tracks OPTIONS", http.MethodOptions, nil, nil, true},
		{"GET list tracks GET", http.MethodGet, []string{"GET"}, nil, true},
		{"GET list skips HEAD", http.MethodHead, []string{"GET"}, nil, false},
		{"GET list skips OPTIONS", http.MethodOptions, []string{"GET"}, nil, false},
		{"explicit POST", http.MethodPost, []string{"GET", "POST"}, nil, true},
		{"empty list skips GET", http.MethodGet, []string{}, nil, false},
		{"path inherits domain list", http.MethodHead, []string{"GET"}, &PathConfig{}, false},
		{"path replaces domain list", http.MethodPost, []string{"GET"}, &PathConfig{TrackedMethods: []string{"POST"}}, true},
		{"path replacement excludes GET", http.MethodGet, []string{"GET"}, &PathConfig{TrackedMethods: []string{"POST"}}, false},
		{"empty path override skips GET", http.MethodGet, []string{"GET"}, &PathConfig{TrackedMethods: []string{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trackingCalls := make(chan string, 1)
			matomo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trackingCalls <- r.Method
				w.WriteHeader(http.StatusNoContent)
			}))
			defer matomo.Close()

			domain := DomainConfig{TrackingEnabled: true, IdSite: 1, TrackedMethods: tt.trackedMethods}
			if tt.override != nil {
				domain.PathOverrides = map[string]PathConfig{"/api": *tt.override}
			}
			config := &Config{MatomoURL: matomo.URL, Domains: map[string]DomainConfig{"demo.localhost": domain}}
			req := httptest.NewRequest(tt.method, "http://demo.localhost/api/test", nil)
			req.RemoteAddr = "203.0.113.9:54321"
			forwarded := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				forwarded = true
				if r != req || r.Method != tt.method {
					t.Fatal("original request or method changed before forwarding")
				}
				w.WriteHeader(http.StatusAccepted)
			})
			handler, err := New(context.Background(), next, config, "test")
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if !forwarded || recorder.Code != http.StatusAccepted {
				t.Fatalf("request forwarding failed: forwarded=%v, status=%d", forwarded, recorder.Code)
			}

			timeout := 200 * time.Millisecond
			if tt.wantTracking {
				timeout = 2 * time.Second
			}
			select {
			case method := <-trackingCalls:
				if !tt.wantTracking {
					t.Fatal("unexpected Matomo tracking request")
				}
				if method != http.MethodGet {
					t.Fatalf("Matomo request method = %q; want GET", method)
				}
			case <-time.After(timeout):
				if tt.wantTracking {
					t.Fatal("expected Matomo tracking request was not received")
				}
			}
		})
	}
}
