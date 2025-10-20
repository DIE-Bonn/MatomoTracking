package MatomoTracking

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MatomoTracking is the middleware that handles Matomo tracking.
type MatomoTracking struct {
	next   http.Handler
	name   string
	config *Config
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

	// Extract the domain name from the request host (without port)
	host := req.Host
	var requestedDomain string
	if strings.Contains(host, ":") {
		var err error
		requestedDomain, _, err = net.SplitHostPort(host)
		if err != nil {
			requestedDomain = host
		}
	} else {
		requestedDomain = host
	}

	fmt.Println("Requested Domain:", requestedDomain)

	// Retrieve domain configuration
	domainConfig, ok := m.config.Domains[requestedDomain]
	if !ok {
		fmt.Println("No config found for domain:", requestedDomain)
		m.next.ServeHTTP(rw, req)
		return
	}

	// If domain-wide tracking is disabled, skip
	if !domainConfig.TrackingEnabled {
		fmt.Println("Tracking disabled at domain level.")
		m.next.ServeHTTP(rw, req)
		return
	}

	// Start with the base (domain-level) config
	effectiveConfig := domainConfig
	requestPath := req.URL.Path

	// Look for the best matching path override (longest prefix match)
	if domainConfig.PathOverrides != nil {
		var bestMatch string
		for prefix := range domainConfig.PathOverrides {
			if pathMatchesPrefix(requestPath, prefix) && len(prefix) > len(bestMatch) {
				bestMatch = prefix
			}
		}

		// Apply path-level override if found
		if bestMatch != "" {
			override := domainConfig.PathOverrides[bestMatch]
			fmt.Printf("Applying path override for prefix: %s\n", bestMatch)
			effectiveConfig = mergeConfigs(domainConfig, override)
		}
	}

	// Invoke next and capture final status/headers
	rec := newStatusRecorder(rw)
	m.next.ServeHTTP(rec, req)

	// Decide post-response whether to track
	shouldTrack := effectiveConfig.TrackingEnabled &&
		!isPathExcluded(requestPath, effectiveConfig.ExcludedPaths, effectiveConfig.IncludedPaths) &&
		matchesResponseConditions(rec.status, rec.Header(), effectiveConfig.ResponseConditions)

	if shouldTrack {
		fmt.Println("Tracking the request...")
		go m.sendTrackingRequest(req, effectiveConfig, requestedDomain)
	} else {
		fmt.Println("Tracking skipped (disabled, excluded, or response conditions not met).")
	}
}

// sendTrackingRequest sends a tracking request to Matomo asynchronously.
func (m *MatomoTracking) sendTrackingRequest(req *http.Request, domainConfig DomainConfig, requestedDomain string) {
	// Get IP address of requesting client/proxy
	clientIP, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		fmt.Println("Error reading client IP:", err)
		return
	}

	// fmt.Println("Client Request: ", req)
	fmt.Println("Client Remote Address: ", req.RemoteAddr)

	// Build the Matomo URL
	matomoReqURL, err := url.Parse(m.config.MatomoURL)
	if err != nil {
		fmt.Println("Error parsing Matomo URL:", err)
		return
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

	// Determine the scheme (http or https)
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}

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

	// Set matomo request headers
	matomoReq.Header.Set("User-Agent", req.Header.Get("User-Agent"))

	// Set or append to the X-Forwarded-For header to preserve the client IP chain for Matomo tracking.
	// The first entry is the original client ip
	// The last entry is the ip of the last/previous client/proxy in the chain
	// Matomo should be configured to use the first entry in the X-Forwarded-For header for tracking
	if existingXFF := req.Header.Get("X-Forwarded-For"); existingXFF != "" {
		fmt.Println("Existing XFF: ", existingXFF)
		matomoReq.Header.Set("X-Forwarded-For", existingXFF+","+clientIP)
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
