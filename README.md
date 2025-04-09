<p align="center">
<img src="https://github.com/DIE-Bonn/MatomoTracking/raw/main/.assets/matomo_tracking_logo.png" 
alt="Matomo_Tracking_Logo" title="Matomo_Tracking_Logo" />
</p>

---

<h1 align="center">
<img alt="GitHub" src="https://img.shields.io/github/license/DIE-Bonn/MatomoTracking?color=blue">
<img alt="GitHub release (latest by date including pre-releases)" src="https://img.shields.io/github/v/release/DIE-Bonn/MatomoTracking?include_prereleases">
<img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/DIE-Bonn/MatomoTracking">
<img alt="GitHub issues" src="https://img.shields.io/github/issues/DIE-Bonn/MatomoTracking">
<img alt="GitHub last commit (branch)" src="https://img.shields.io/github/last-commit/DIE-Bonn/MatomoTracking/main">
</h1>

# Matomo Tracking

## Overview

This plugin, `MatomoTracking`, is designed for Traefik as middleware to handle the tracking of requests using the Matomo analytics platform. The plugin inspects incoming HTTP requests, checks if tracking is enabled for the requested domain, and sends tracking data to Matomo if required.

The main purpose of this plugin is to enhance the accuracy of visitor tracking by overcoming limitations associated with the traditional JavaScript-based tracking method used by Matomo. Standard tracking relies on JavaScript code running in the user's browser, which can be blocked by certain browser extensions or privacy tools. By capturing tracking data directly on the server side, this plugin ensures that visitor information is accurately recorded even when JavaScript is disabled or blocked, providing a more reliable and comprehensive analytics solution.

![Matomo Tracking diagram](https://github.com/DIE-Bonn/MatomoTracking/raw/main/.assets/matomo_tracking_network_flow.png)

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
            ```yaml
            domains:
              "www3.example.com":
                trackingEnabled: true
                idSite: 21
                excludedPaths:
                  - "/admin/*"
                  - "\\.\\w{1,5}(\\?.+)?$"
                includedPaths:
                  - "\\.(php|aspx)(\\?.*)?$"
                pathOverrides:
                  "/subdir":
                    trackingEnabled: true
                    idSite: 24
                    excludedPaths:
                      - "/test"
                      - "/test2"
                  "/subdir2":
                    trackingEnabled: false
                    idSite: 24
              "test.de":
                trackingEnabled: false
                idSite: 456
            ```

2. `DomainConfig` Struct

    **Purpose**:

    The `DomainConfig` struct defines the tracking settings for a specific domain. Each domain that should be tracked (or explicitly not tracked) has a corresponding `DomainConfig` entry.

    **Fields**:

    - `TrackingEnabled`:
        - **Type**: `bool`
        - **Description**: Indicates whether tracking is enabled for the given domain. If `true`, the plugin will attempt to send tracking data to Matomo. If `false`, tracking is disabled for that domain.
        - **Example**: true
    - `IdSite`:
        - **Type**: `int`
        - **Description**: The unique identifier for the website or domain in Matomo. Each domain has its own site ID that Matomo uses to differentiate between multiple websites being tracked on the same server.
        - **Example**: `21`
    - `ExcludedPaths`:
        - **Type**: `[]string` (Slice of strings)
        - **Description**: A list of regular expressions that define URL paths that should be excluded from tracking. If the requested path matches any of the regex patterns in this list, the request will not be tracked by Matomo.
        - **Example**:
            ```yaml
            excludedPaths:
              - "/admin/*"
              - "\\.\\w{1,5}(\\?.+)?$"
            ```
    - `IncludedPaths`:
        - **Type**: `[]string` (Slice of strings)
        - **Description**: A list of regular expressions that define URL paths that should be explicitly included for tracking. If a requested path matches any of the regex patterns in this list, the request will be tracked by Matomo, even if it matches an exclusion pattern.
        - **Example**:
            ```yaml
            includedPaths:
              - "\\.(php|aspx)(\\?.*)?$"
            ```
    - `PathOverrides`:
        - **Type**: `map[string]PathConfig`
        - **Description**: A map of path-specific configuration overrides that apply only to requests matching those paths. Each key is a path prefix (e.g., `/api`, `/special`) and its corresponding value is a `PathConfig` block. This feature allows more granular control over tracking behavior within a domain.
        Path overrides support the same fields as the domain-level configuration: `trackingEnabled`, `idSite`, `excludedPaths`, and `includedPaths`. If a path override is defined, it will **override** the corresponding settings from the parent domain **only for requests matching that path**.
        Matching is done using **prefix matching with boundary awareness**. This means:
          - `/test` matches `/test` and `/test/something`
          - `/test` does not match `/test2` or `/testing`

### Configuration Breakdown

Let's analyze how the configuration is used in the `dynamic.yml` file:

**Example Configuration**

```yaml
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
            - "\\.\\w{1,5}(\\?.+)?$"
          includedPaths:
            - "\\.(php|aspx)(\\?.*)?$"
          pathOverrides:
            "/subdir":
              trackingEnabled: true
              idSite: 24
              excludedPaths:
                - "/test"
                - "/test2"
            "/subdir2":
              trackingEnabled: false
              idSite: 24
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
            - `\\.\\w{1,5}(\\?.+)?$`: Excludes files with extensions between 1 and 5 characters, and optionally followed by query parameters.
        - `includedPaths`: Specifies paths that should be tracked, even if they are excluded. 
        For example:
            - `\\.(php|aspx)(\\?.*)?$`: Includes files with extensions `.php` and `.aspx`, and optionally followed by query parameters
        - `pathOverrides`: The pathOverrides block allows you to define more specific tracking settings for particular path prefixes under the domain. These overrides take precedence over the domain-level settings, but only for requests that match the specified paths.
        Matching is based on prefix matching with path boundary awareness, meaning:
          - `/subdir` matches `/subdir` and `/subdir/test`, but not `/subdir2`.
        
          For example:
          - `"/subdir"`:
            - `trackingEnabled: true`: Enables tracking for paths under `/subdir`, even if other paths are excluded.
            - `idSite: 24`: Overrides the site ID used for this specific path.
            - `excludedPaths`: Within `/subdir`, `/subdir/test` and `/subdir/test2` are excluded from tracking.
          - `"/subdir2"`:
            - `trackingEnabled: false`: Disables tracking entirely for any request under `/subdir2`.
            - `idSite: 24`: Still defines an ID, but it is unused since tracking is disabled here.

    - `"test.de"`:
        - `trackingEnabled: false`: Disables tracking for `test.de`.
        - `idSite: 456`: Uses `456` as the Matomo site ID.


## Code Documentation

### ServeHTTP Method

Main logic of the middleware:

1. Extracts the requested domain from the host.
2. Checks if tracking is enabled for the domain.
3. If `pathOverrides` are defined, the middleware:
    - Searches for the most specific matching path override (using longest prefix match with boundary awareness).
    - Merges the override settings with the domain-level config using `mergeConfigs`.
4. Uses the resulting (effective) config to evaluate `excludedPaths` and `includedPaths`.
5. If tracking is still enabled and the path is not excluded, sends a tracking request to Matomo asynchronously.
6. Forwards the request to the next handler in the chain.

### mergeConfigs Function

Merges a domain-level configuration with a path-level override.

- Any field explicitly set in the path override replaces the value from the base domain config.
- This function ensures that only the overridden fields change, while other inherited values remain intact.

### sendTrackingRequest Method

Sends a tracking request to Matomo asynchronously:

1. Constructs the tracking URL with the appropriate query parameters.
2. Creates an HTTP GET request to Matomo.
3. Sets the `User-Agent` and `X-Forwarded-For` headers to identify the client.
4. Sends the request using a custom HTTP client.
5. Logs the response status.

### isPathExcluded Function

Checks if a given path matches any of the exclusion patterns:

1. Iterates through all exclusion regex patterns provided in the configuration in `excludedPaths`.
2. For each pattern:
    - Attempts to match the pattern against the given path.
    - Logs any errors encountered while processing the regex.
    - If a match is found, sets a flag indicating the path is excluded and stops further checks.
3. If no match is found in the exclusion patterns, returns `false` immediately, indicating the path is not excluded
4. Proceeds to check against inclusion patterns if an exclusion match is found.
5. If a match is found in `includedPaths`, the function returns `false`, indicating that the path should not be excluded, as it is explicitly included.


## Setup instructions

Step 1: **Load/import the plugin into traefik**

1. Edit your Traefik static configuration file (e.g., traefik.yml or traefik.toml), and add the plugin's Github repository:

    Example: `traefik.yml`:
    ```yaml
    experimental:
      plugins:
        matomoTracking:
          moduleName: "github.com/DIE-Bonn/MatomoTracking"
          version: "v1.0.5"
    ```
**Ensure to use the current version tag.**

Step 2: **Configure Dynamic Configuration**

1. Create a new or use an already existing dynamic configuration file (e.g., dynamic.yml) that defines how the plugin should behave:

    Example `dynamic.yml`:
    ```yaml
    http:
      middlewares:
        matomo-tracking:
          plugin:
            matomoTracking:
              matomoURL: "https://matomo-staging.die-bonn.de/matomo.php"
              domains:
                "www3.die-bonn.de":
                  trackingEnabled: true
                  idSite: 21
                  excludedPaths:
                    - "/admin/*"
                    - "\\.\\w{1,5}(\\?.+)?$"
                  includedPaths:
                    - "\\.(php|aspx)(\\?.*)?$"
                  pathOverrides:
                    "/subdir":
                      trackingEnabled: true
                      idSite: 24
                      excludedPaths:
                        - "/test"
                        - "/test2"
                    "/subdir2":
                      trackingEnabled: false
                      idSite: 24
                "kansas-suche.de":
                  trackingEnabled: false
                  idSite: 456
    ```

    - This configuration defines the global rules for the `matomo-tracking` middleware, consisting of domain names with their individual tracking configuration.

Step 3: **Associate the middleware plugin to the entrypoint**

1. Edit your Traefik static configuration file `traefik.yml`:

    Example `traefik.yml`:

    ```yaml
    entryPoints:
      webinsecure:
        address: ":80"
        http:
          middlewares:
            - matomo-tracking@file
    ```

    - This configuration ensures that the `matomo-tracking` plugin can analyze all incoming requests to decide which requests will be sent to the matomo server for tracking purposes.

Step 4: **Restart Traefik**

1. Start or restart traefik to load the plugin and apply the new configuration

    ```bash
    docker compose down && docker compose up -d
    ```

