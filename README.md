# Matomo Tracking

## Overview

This plugin, `matomoTracking`, is designed for Traefik as middleware to handle the tracking of requests using the Matomo analytics platform. The plugin inspects incoming HTTP requests, checks if tracking is enabled for the requested domain, and sends tracking data to Matomo if required.

The main purpose of this plugin is to enhance the accuracy of visitor tracking by overcoming limitations associated with the traditional JavaScript-based tracking method used by Matomo. Standard tracking relies on JavaScript code running in the user's browser, which can be blocked by certain browser extensions or privacy tools. By capturing tracking data directly on the server side, this plugin ensures that visitor information is accurately recorded even when JavaScript is disabled or blocked, providing a more reliable and comprehensive analytics solution.

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
                  - "\\.\\w{1,5}(\\?.+)?$"
                includedPaths:
                  - "\\.(php|aspx)(\\?.*)?$"
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
            ```
            excludedPaths:
              - "/admin/*"
              - "\\.\\w{1,5}(\\?.+)?$"
            ```
    - `IncludedPaths`:
        - **Type**: `[]string` (Slice of strings)
        - **Description**: A list of regular expressions that define URL paths that should be explicitly included for tracking. If a requested path matches any of the regex patterns in this list, the request will be tracked by Matomo, even if it matches an exclusion pattern.
        - **Example**:
            ```
            includedPaths:
              - "\\.(php|aspx)(\\?.*)?$"
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
            - "\\.\\w{1,5}(\\?.+)?$"
          includedPaths:
            - "\\.(php|aspx)(\\?.*)?$"
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

1. Iterates through all exclusion regex patterns provided in the configuration in `excludedPaths`.
2. For each pattern:
    - Attempts to match the pattern against the given path.
    - Logs any errors encountered while processing the regex.
    - If a match is found, sets a flag indicating the path is excluded and stops further checks.
3. If no match is found in the exclusion patterns, returns `false` immediately, indicating the path is not excluded
4. Proceeds to check against inclusion patterns if an exclusion match is found.
5. If a match is found in `includedPaths`, the function returns `false`, indicating that the path should not be excluded, as it is explicitly included.


## Setup instructions

Step 1: **Create the Plugin**

1. Create a directory for your plugin, for example: `traefik/plugins-local/src/matomo_tracking/`.

2. Place the plugin’s Go source code files in this directory:

    - main.go (contains the plugin logic).
    - .traefik.yml (meta file for loading the plugin)
    - Other necessary Go files (if any).

    Here's an example structure

    ```
    traefik/
    ├── plugins-local/
    │   └── src/
    │       └── matomo_tracking/ 
    │           ├── .traefik.yml
    │           ├── go.mod
    │           └── main.go
    ```

Step 2: **Configure Traefik for Local Plugins**

1. Edit your Traefik static configuration file (e.g., traefik.yml or traefik.toml), and enable experimental local plugins:

    Example: `traefik.yml`:
    ```
    experimental:
      localPlugins:
        matomoTracking:
          moduleName: matomo_tracking
    ```

2. Edit your Traefik `docker-compose.yml` file and create a bind-mount for the `plugins-local` directory in order to make it available for the traefik container.

    Example: `docker-compose.yml`:
    ```
    services:
      edgerouter:
        image: traefik:2.11
        container_name: traefik
        security_opt:
          - no-new-privileges:true
        ports:
          - 80:80
          - 443:443
          - 11111:11111
        restart: "always"
        volumes:
          - /etc/localtime:/etc/localtime:ro
          - /var/run/docker.sock:/var/run/docker.sock:ro
          - ./config:/etc/traefik
          - ./plugins-local:/plugins-local
          - ./letsencrypt:/letsencrypt
          - /certs:/certs
        networks:
          - edgerouter

    networks:
      edgerouter:
        external: true
    ```

Step 3: **Configure Dynamic Configuration**

1. Create a dynamic configuration file (e.g., dynamic.yml) that defines how the plugin should behave:

    Example `dynamic.yml`:
    ```
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
                "kansas-suche.de":
                  trackingEnabled: false
                  idSite: 456
    ```

    - This configuration defines the global rules for the `matomo-tracking` middleware, consisting of domain names with their individual tracking configuration.

Step 4: **Associate the middleware plugin to the entrypoint**

1. Edit your Traefik static configuration file `traefik.yml`:

    Example `traefik.yml`:

    ```
    entryPoints:
      webinsecure:
        address: ":80"
        http:
          middlewares:
            - matomo-tracking@file
    ```

    - This configuration ensures that the `matomo-tracking` plugin can analyze all incoming requests to decide which requests to send to the matomo server for tracking purposes.

Step 5: **Restart Traefik**

1. Start or restart traefik to load the plugin and apply the new configuration

    ```bash
    docker compose down && docker compose up -d
    ```

