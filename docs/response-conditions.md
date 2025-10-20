# Response-based tracking conditions

This feature defers the tracking decision until after the response has been generated. You can restrict tracking to specific HTTP status codes and/or response headers.

Summary
- Tracking is evaluated post-response.
- Conditions can be configured per domain and overridden per path.
- Backward compatible: if no conditions are set, behavior stays unchanged.

Configuration schema
- DomainConfig.responseConditions
- PathConfig.responseConditions
- ResponseConditions:
  - trackOnStatusCodes: list of allowed final status codes (empty = allow any)
  - trackWhenHeaders: required response headers (exact key/value matches; header names are case-insensitive)

Evaluation order
1) Domain enabled (trackingEnabled).
2) Path include/exclude rules.
3) Response conditions after the handler writes the response.
   - All headers in trackWhenHeaders must be present with exactly matching values.
   - If trackOnStatusCodes is set, status must be one of the listed values.

Traefik dynamic config (YAML)
```yaml
http:
  middlewares:
    matomo-tracking:
      plugin:
        matomoTracking:
          matomoURL: "http://matomo-local/matomo.php"
          domains:
            "demo.localhost":
              trackingEnabled: true
              idSite: 1
              # Track only successful HTML pages
              responseConditions:
                trackOnStatusCodes: [200]
                trackWhenHeaders:
                  Content-Type: "text/html; charset=UTF-8"
```

Notes and limitations
- Conditions are ANDed (status AND all headers must match).
- Multi-value headers pass if any value equals the configured one.
- Header names are case-insensitive; values match exactly (no regex).
- Tracking is sent after response completion; long-running responses delay the send.

Testing
- Unit tests: response_conditions_unit_test.go
- Integration tests: response_conditions_integration_test.go (requires local Matomo)
  - Run: go test -v ./...