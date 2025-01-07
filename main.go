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
	"time"
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
	IncludedPaths	[]string `json:"includedPaths,omitempty"`
}

// CreateConfig returns the default configuration for the plugin.
func CreateConfig() *Config {
	return &Config{
		MatomoURL: "",
		Domains: nil,
	}
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
	fmt.Println("Plugin: Matomo Tracking")
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

	fmt.Println("Requested Domain:", requestedDomain)
	fmt.Println("Matomo Tracking config: ", m.config)

	// Iterate through the config to check if the requested domain shall be tracked
        for domainName, domainConfig := range m.config.Domains {
                // Check if the requested domain exists in the config 
                if domainName == requestedDomain {
			// Check if tracking is enabled for this domain
			fmt.Printf("Requested domain %s found in config with the settings: %t, %d, %s", requestedDomain, domainConfig.TrackingEnabled, domainConfig.IdSite, domainConfig.ExcludedPaths)
                        if domainConfig.TrackingEnabled {
                                fmt.Println("Requested Domain exists and is enabled.")
                                // Check if the requested path matches the exclusion list, if not track the request
                                if !isPathExcluded(req.URL.Path, domainConfig.ExcludedPaths, domainConfig.IncludedPaths) {
                                        fmt.Println("Tracking the requested URL.")

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

    requestURI := req.URL.RequestURI()
    // Parse the URI
    parsedURI, err := url.Parse(requestURI)
    if err != nil {
        fmt.Println("Error parsing URI:", err)
        return
    }
    // Convert the path to lowercase
    parsedURI.Path = strings.ToLower(parsedURI.Path)

    // Reconstruct the URI with the lowercase path
    requestURI = parsedURI.String()

    // Construct the full URL
    fullURL := scheme + "://" + requestedDomain + requestURI
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

    // Create a custom HTTP client with timeouts
    var customClient = &http.Client{
        Timeout: 10 * time.Second, // Set a global timeout for requests
    }

    // Use this client to send requests
    resp, err := customClient.Do(matomoReq)
    if err != nil {
        fmt.Println("Error sending Matomo request:", err)
        return
    }
    //close the response body to ensure that resources such as network connections associated with the HTTP response body are released properly
    defer resp.Body.Close()

    // Process the response
    fmt.Println("Matomo response status:", resp.Status)
}

func isPathExcluded(path string, excludedPaths, includedPaths []string) bool {
      fmt.Println("Checking path:", path)

      // First, check if the path matches any of the excluded patterns
      excludedMatch := false
      for _, excludedPath := range excludedPaths {
            fmt.Println("Testing against excluded path pattern:", excludedPath)

            matches, err := regexp.MatchString(excludedPath, path)
            if err != nil {
                  // Log the error and continue with the next pattern
                  fmt.Println("Error matching regex for excluded path:", err)
                  continue
            }

            if matches {
                  fmt.Println("Path matches excluded pattern:", excludedPath)
                  excludedMatch = true
                  break
            }
      }

      // If there's no match in excluded paths, the path is not excluded
      if !excludedMatch {
            fmt.Println("No match found in excluded paths; path is not excluded.")
            return false // Do not exclude
      }

      // Now check if the path matches any of the included patterns
      for _, includedPath := range includedPaths {
            fmt.Println("Testing against included path pattern:", includedPath)

            matches, err := regexp.MatchString(includedPath, path)
            if err != nil {
                  // Log the error and continue with the next pattern
                  fmt.Println("Error matching regex for included path:", err)
                  continue
            }

            if matches {
                  fmt.Println("Path matches included pattern:", includedPath)
                  return false // Path should be included, so not excluded
            }
      }

      // If it matched an excluded path but not an included path, exclude it
      fmt.Println("Path is excluded due to no matching included pattern.")
      return true
}
