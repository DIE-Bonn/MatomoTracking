package matomo_tracking

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"strconv"
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
	host := req.Host
        var requestedDomain string

        // Check if the host contains a colon (indicating a possible port)
        if strings.Contains(host, ":") {
        	// Try to split the host and port
		var err error
        	requestedDomain, _, err = net.SplitHostPort(host)
        	if err != nil {
            		// If there's an error, fallback to using the entire host as the domain
            		requestedDomain = host
        	}
        } else {
        	// No colon found, so host is already just the domain
        	requestedDomain = host
    	}

	fmt.Println("Matomo Tracking")
	fmt.Println("Requested Domain: ", requestedDomain)
	fmt.Println(m.config)

	// Iterate through the config to check if the requested domain shall be tracked
        for domainName, domainConfig := range m.config.Domains {
                fmt.Println(domainName)
                fmt.Println(domainConfig.TrackingEnabled)
                fmt.Println(domainConfig.IdSite)
                fmt.Println(domainConfig.ExcludedPaths)
                // Check if the requested domain exists in the config 
                if domainName == requestedDomain {
			// Check if tracking is enabled for this domain
                        if domainConfig.TrackingEnabled {
                                fmt.Println("Requested Domain exists and is enabled.")
                                // Check if the requested path matches the exclusion list, if not track the request
                                if !isPathExcluded(req.URL.Path, domainConfig.ExcludedPaths) {
                                        fmt.Println("Track the requested URL.")

                                        // Perform the tracking request asynchronously
                                        go m.sendTrackingRequest(req, domainConfig, requestedDomain)
                                }
                        }
                        break
                }
        }
	
	m.next.ServeHTTP(rw, req)

}

// sendTrackingRequest sends a tracking request to Matomo asynchronously.
func (m *MatomoTracking) sendTrackingRequest(req *http.Request, domainConfig DomainConfig, requestedDomain string) {
    // Build the Matomo URL
    matomoReqURL, err := url.Parse(m.config.MatomoURL)
    if err != nil {
        fmt.Println("Error parsing Matomo URL:", err)
        return
    }

    // Determine the scheme (http or https)
    scheme := "http"
    if req.TLS != nil {
        scheme = "https"
    }

    // Construct the full URL
    fullURL := scheme + "://" + requestedDomain + req.URL.RequestURI()
    query := matomoReqURL.Query()
    query.Set("url", fullURL)
    query.Set("rec", "1")
    query.Set("idsite", strconv.Itoa(domainConfig.IdSite))
    matomoReqURL.RawQuery = query.Encode()
    fmt.Println("Matomo query string:", matomoReqURL.RawQuery)

    // Create the Matomo request
    matomoReq, err := http.NewRequest("GET", matomoReqURL.String(), nil)
    if err != nil {
        fmt.Println("Error creating Matomo request:", err)
        return
    }
    matomoReq.Header.Set("User-Agent", req.Header.Get("User-Agent"))

    // Extract the client's IP address
    clientIP := req.Header.Get("X-Forwarded-For")
    if clientIP == "" {
        // If the X-Forwarded-For header is empty, use the remote address
        clientIP, _, _ = net.SplitHostPort(req.RemoteAddr)
    }

    // Set or update the X-Forwarded-For header
    if existingXFF := matomoReq.Header.Get("X-Forwarded-For"); existingXFF != "" {
        matomoReq.Header.Set("X-Forwarded-For", existingXFF+", "+clientIP)
    } else {
        matomoReq.Header.Set("X-Forwarded-For", clientIP)
    }

    fmt.Println("Matomo tracking request: ", matomoReq)

    // Send the request asynchronously
    resp, err := http.DefaultClient.Do(matomoReq)
    if err != nil {
        fmt.Println("Error sending Matomo request:", err)
    } else {
	//close the response body to ensure that resources such as network connections associated with the HTTP response body are released properly
        defer resp.Body.Close()
        fmt.Println("Matomo response status:", resp.Status)
    }
}

func isPathExcluded(path string, excludedPaths []string) bool {
    fmt.Println("Checking path:", path)

    for _, excludedPath := range excludedPaths {
        fmt.Println("Testing against excluded path pattern:", excludedPath)

        matches, err := regexp.MatchString(excludedPath, path)
        if err != nil {
            // Log the error and continue with the next pattern
            fmt.Println("Error matching regex:", err)
            continue
        }

        if matches {
            fmt.Println("Path matches excluded pattern:", excludedPath)
            return true
        }
    }

    fmt.Println("No matching excluded pattern found.")
    return false
}
