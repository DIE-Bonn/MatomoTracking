package matomo_tracking

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Config represents the configuration for the MatomoTracking plugin.
type Config struct {
	MatomoURL string                  `json:"matomoURL,omitempty"`
	Domains   map[string]DomainConfig `json:"domains,omitempty"`
}

// DomainConfig specifies the tracking rules for a specific domain.
type DomainConfig struct {
	TrackingEnabled bool     `json:"trackingEnabled,omitempty"`
	IdSite          int      `json:"idSite,omitempty"`
	ExcludedPaths   []string `json:"excludedPaths,omitempty"`
}

// CreateConfig returns the default configuration for the plugin.
func CreateConfig() *Config {
	return &Config{}
}

// MatomoTracking is the middleware that handles Matomo tracking.
type MatomoTracking struct {
	next     http.Handler
	name     string
	config   *Config
}

// New creates a new instance of the MatomoTracking middleware.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &MatomoTracking{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

// ServeHTTP is the main logic of the middleware. It processes the request,
// performs Matomo tracking if enabled for the domain, and forwards the request
// to the next handler in the chain.
func (m *MatomoTracking) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("Matomo Tracking")
	fmt.Println(m.config)

	for _, domain := range m.config.Domains {
		fmt.Println(domain)
		fmt.Println(domain.TrackingEnabled)
		fmt.Println(domain.IdSite)
		fmt.Println(domain.ExcludedPaths)
	}

	m.next.ServeHTTP(rw, req)

}
