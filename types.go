package MatomoTracking

// PathConfig specifies the tracking rules for a specific path.
type PathConfig struct {
	TrackingEnabled    *bool               `json:"trackingEnabled,omitempty"`
	IdSite             *int                `json:"idSite,omitempty"`
	ExcludedPaths      []string            `json:"excludedPaths,omitempty"`
	IncludedPaths      []string            `json:"includedPaths,omitempty"`
	ResponseConditions *ResponseConditions `json:"responseConditions,omitempty"`
}

// DomainConfig specifies the tracking rules for a specific domain.
type DomainConfig struct {
	TrackingEnabled    bool                  `json:"trackingEnabled,omitempty"`
	IdSite             int                   `json:"idSite,omitempty"`
	ExcludedPaths      []string              `json:"excludedPaths,omitempty"`
	IncludedPaths      []string              `json:"includedPaths,omitempty"`
	PathOverrides      map[string]PathConfig `json:"paths,omitempty"`
	ResponseConditions *ResponseConditions   `json:"responseConditions,omitempty"`
}

// Config represents the configuration for the MatomoTracking plugin.
type Config struct {
	MatomoURL string                  `json:"matomoURL,omitempty"`
	Domains   map[string]DomainConfig `json:"domains,omitempty"`
}

// CreateConfig returns the default configuration for the plugin.
func CreateConfig() *Config {
	return &Config{
		MatomoURL: "",
		Domains:   nil,
	}
}
