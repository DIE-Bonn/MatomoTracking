package MatomoTracking

import (
	"fmt"
	"regexp"
	"strings"
)

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

func mergeConfigs(base DomainConfig, override PathConfig) DomainConfig {
	merged := base // Start with the domain-level config

	if override.TrackingEnabled != nil {
		merged.TrackingEnabled = *override.TrackingEnabled
	}

	if override.IdSite != nil {
		merged.IdSite = *override.IdSite
	}

	// For slice overrides, we completely replace the base slices
	if override.ExcludedPaths != nil {
		merged.ExcludedPaths = override.ExcludedPaths
	}

	if override.IncludedPaths != nil {
		merged.IncludedPaths = override.IncludedPaths
	}

	if override.ResponseConditions != nil {
		merged.ResponseConditions = override.ResponseConditions
	}
	return merged
}

func pathMatchesPrefix(path, prefix string) bool {
	// Exact match or subpath with trailing slash
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
