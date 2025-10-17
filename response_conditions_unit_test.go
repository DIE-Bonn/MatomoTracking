package MatomoTracking

import (
	"net/http"
	"testing"
)

func TestMatchesResponseConditions_Nil(t *testing.T) {
	t.Parallel()
	h := http.Header{}
	if !matchesResponseConditions(404, h, nil) {
		t.Fatalf("nil conditions should allow any status/headers")
	}
}

func TestMatchesResponseConditions_StatusOnly(t *testing.T) {
	t.Parallel()
	rc := &ResponseConditions{TrackOnStatusCodes: []int{200, 201}}
	h := http.Header{}

	if !matchesResponseConditions(200, h, rc) {
		t.Fatalf("expected status 200 to match")
	}
	if matchesResponseConditions(404, h, rc) {
		t.Fatalf("expected status 404 to NOT match")
	}
}

func TestMatchesResponseConditions_HeadersOnly(t *testing.T) {
	t.Parallel()

	// Exact value required; header name case-insensitive
	rc := &ResponseConditions{
		TrackWhenHeaders: map[string]string{
			"Content-Type": "text/html; charset=UTF-8",
		},
	}

	// Match exact value
	h := http.Header{}
	h.Set("content-type", "text/html; charset=UTF-8") // mixed case set
	if !matchesResponseConditions(200, h, rc) {
		t.Fatalf("expected header to match")
	}

	// Non-matching value
	h2 := http.Header{}
	h2.Set("Content-Type", "application/json")
	if matchesResponseConditions(200, h2, rc) {
		t.Fatalf("expected header to NOT match")
	}

	// Multiple values where one matches
	h3 := http.Header{}
	h3.Add("Content-Type", "application/json")
	h3.Add("Content-Type", "text/html; charset=UTF-8")
	if !matchesResponseConditions(200, h3, rc) {
		t.Fatalf("expected one-of multiple header values to match")
	}
}

func TestMatchesResponseConditions_BothStatusAndHeaders(t *testing.T) {
	t.Parallel()

	rc := &ResponseConditions{
		TrackOnStatusCodes: []int{200},
		TrackWhenHeaders: map[string]string{
			"X-App":        "web",
			"Content-Type": "text/html",
		},
	}

	// All match
	h := http.Header{}
	h.Set("X-App", "web")
	h.Set("Content-Type", "text/html")
	if !matchesResponseConditions(200, h, rc) {
		t.Fatalf("expected combined conditions to match")
	}

	// Status mismatch
	if matchesResponseConditions(404, h, rc) {
		t.Fatalf("expected status mismatch to fail")
	}

	// Header missing
	hMissing := http.Header{}
	hMissing.Set("X-App", "web")
	if matchesResponseConditions(200, hMissing, rc) {
		t.Fatalf("expected missing header to fail")
	}

	// Header present but value different
	hDiff := http.Header{}
	hDiff.Set("X-App", "api")
	hDiff.Set("Content-Type", "text/html")
	if matchesResponseConditions(200, hDiff, rc) {
		t.Fatalf("expected header value mismatch to fail")
	}
}
