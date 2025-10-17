package MatomoTracking

import (
	"testing"
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
