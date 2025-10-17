package MatomoTracking

import "net/http"

// ResponseConditions define when to track based on the final response.
type ResponseConditions struct {
	// Track only if the final status code is one of these. Empty = allow any.
	TrackOnStatusCodes []int `json:"trackOnStatusCodes,omitempty"`
	// All headers must be present and equal to the given value (exact match, case-insensitive keys).
	TrackWhenHeaders map[string]string `json:"trackWhenHeaders,omitempty"`
}

// statusRecorder captures the final status while delegating to the real ResponseWriter.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	// Default to 200 unless WriteHeader is called explicitly.
	return &statusRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// matchesResponseConditions returns true if rc is nil or all conditions match.
func matchesResponseConditions(status int, hdr http.Header, rc *ResponseConditions) bool {
	if rc == nil {
		return true
	}
	// Status filter
	if len(rc.TrackOnStatusCodes) > 0 {
		ok := false
		for _, s := range rc.TrackOnStatusCodes {
			if s == status {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	// Header filters (exact value match)
	for k, want := range rc.TrackWhenHeaders {
		found := false
		for _, v := range hdr.Values(k) {
			if v == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
