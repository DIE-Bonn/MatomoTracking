# Matomo Tracking

## Overview

This plugin, `matomoTracking`, is designed for Traefik as middleware to handle tracking of requests using the Matomo analytics platform. The plugin inspects incoming HTTP requests, checks if the tracking is enabled for the requested domain, and sends tracking data to Matomo if required.

## Structs and Configuration Explanation

1. `Config` Struct

    **Purpose**:

    The `Config` struct represents the configuration settings for the entire Matomo Tracking plugin. It defines the global settings that apply to all domains and the specific configurations for each domain that needs tracking.

    **Fields:**

    - `MatomoURL`
        - Type: **string**
        - Description: Specifies the base URL for the Matomo server endpoint where tracking data should be sent. Typically, this is the URL to the `matomo.php` file on the Matomo server, such as `https://matomo.example.com/matomo.php`.
        - Example: `"https://matomo.example.com/matomo.php"`
    - `Domains`:
        - Type: `map[string]DomainConfig`
        - Description: A map where each key is a domain name (as a `string`) and the corresponding value is a `DomainConfig` struct. This allows you to define tracking rules for multiple domains individually.
        - Example:
            ```
            domains:
              "www3.example.com":
                trackingEnabled: true
                idSite: 21
                excludedPaths:
                  - "/admin/*"
                  - "\\.(css|js|ico)$"
                  - "\\.\\w{1,5}(\\?.+)?$"
              "test.de":
                trackingEnabled: false
                idSite: 456
            ```

2. `DomainConfig` Struct

    **Purpose**:

    The `DomainConfig` struct defines the tracking settings for a specific domain. Each domain that should be tracked (or explicitly not tracked) has a corresponding `DomainConfig` entry.

    **Fields**:

    - `TrackingEnabled`:
        - Type: `bool`
        - Description: Indicates whether tracking is enabled for the given domain. If `true`, the plugin will attempt to send tracking data to Matomo. If `false`, tracking is disabled for that domain.
        - Example: true
    - `IdSite`:
        - Type: `int`
        - Description: The unique identifier for the website or domain in Matomo. Each domain has its own site ID that Matomo uses to differentiate between multiple websites being tracked on the same server.
        - Example: `21`
    - `ExcludedPaths`:
        - Type: `[]string` (Slice of strings)
        - Description: A list of regular expressions that define URL paths that should be excluded from tracking. If the requested path matches any of the regex patterns in this list, the request will not be tracked by Matomo.
        - Example:
            ```
            excludedPaths:
              - "/admin/*"
              - "\\.(css|js|ico)$"
              - "\\.\\w{1,5}(\\?.+)?$"
            ```

### Configuration Breakdown

Let's analyze how the configuration is used in the `dynamic.yml` file:

**Example Configuration**

```
matomo-tracking:
  plugin:
    matomoTracking:
      matomoURL: "https://matomo.example.com/matomo.php"
      domains:
        "www3.example.com":
          trackingEnabled: true
          idSite: 21
          excludedPaths:
            - "/admin/*"
            - "\\.(css|js|ico)$"
            - "\\.\\w{1,5}(\\?.+)?$"
        "test.de":
          trackingEnabled: false
          idSite: 456
```

**Explanation:**

- `matomo-tracking`: This is the name of the plugin instance in Traefik.
- `plugin`: Denotes that the following configuration is for a plugin.
- `matomoTracking`: The name of your custom plugin. It references the code logic that you provided in your Go plugin implementation.
- `matomoURL`: Specifies the Matomo server's URL where the tracking data should be sent.
- `domains`: Contains specific tracking configurations for multiple domains:
    - `"www3.example.com"`:
        - `trackingEnabled: true`: Enables tracking for `www3.example.com.`
        - `idSite: 21`: Uses `21` as the Matomo site ID.
        - `excludedPaths`: Specifies paths that should not be tracked. For example:
            - `/admin/*`: Excludes all paths under `/admin`.
            - `\\.(css|js|ico)$`: Excludes all files with `.css`, `.js`, or `.ico` extensions.
            - `\\.\\w{1,5}(\\?.+)?$`: Excludes files with extensions between 1 and 5 characters, and optionally followed by query parameters.
    - `"test.de"`:
        - `trackingEnabled: false`: Disables tracking for `test.de`.
        - `idSite: 456`: Uses `456` as the Matomo site ID.


## Code Documentation

### ServeHTTP Method

Main logic of the middleware:

1. Extracts the requested domain from the host.
2. Checks if tracking is enabled for the domain.
3. If enabled and not excluded, sends a tracking request to Matomo asynchronously.
4. Forwards the request to the next handler in the chain.

### sendTrackingRequest Method

Sends a tracking request to Matomo asynchronously:

1. Constructs the tracking URL with the appropriate query parameters.
2. Creates an HTTP GET request to Matomo.
3. Sets the `User-Agent` and `X-Forwarded-For` headers to identify the client.
4. Sends the request using a custom HTTP client.
5. Logs the response status.

### isPathExcluded Function

Checks if a given path matches any of the exclusion patterns:

1. Iterates through all exclusion regex patterns.
2. Returns `true` if the path matches any pattern, `false` otherwise.

