// Package config provides environment-driven application configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
)

const (
	defaultPort            = 8080
	defaultSpannerProject  = "cars-project"
	defaultSpannerInstance = "cars-instance"
	defaultSpannerDatabase = "cars-db"
)

// Config holds all runtime configuration for the Cars API server.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port int

	// SpannerProject is the GCP project ID hosting the Spanner instance.
	SpannerProject string

	// SpannerInstance is the Spanner instance ID.
	SpannerInstance string

	// SpannerDatabase is the Spanner database name.
	SpannerDatabase string

	// SpannerEmulatorHost, when non-empty, redirects the Spanner client to the
	// local emulator (e.g. "localhost:9010"). Mirrors the SPANNER_EMULATOR_HOST
	// environment variable that the Go client library also reads automatically.
	SpannerEmulatorHost string
}

// Load reads configuration from environment variables, falling back to defaults
// for any unset value.
func Load() Config {
	return Config{
		Port:                parsePort(getenv("PORT", strconv.Itoa(defaultPort))),
		SpannerProject:      getenv("SPANNER_PROJECT", defaultSpannerProject),
		SpannerInstance:     getenv("SPANNER_INSTANCE", defaultSpannerInstance),
		SpannerDatabase:     getenv("SPANNER_DATABASE", defaultSpannerDatabase),
		SpannerEmulatorHost: os.Getenv("SPANNER_EMULATOR_HOST"),
	}
}

// DatabasePath returns the fully-qualified Spanner database path required by
// the Go Spanner client: projects/{project}/instances/{instance}/databases/{db}.
func (c Config) DatabasePath() string {
	return fmt.Sprintf(
		"projects/%s/instances/%s/databases/%s",
		c.SpannerProject,
		c.SpannerInstance,
		c.SpannerDatabase,
	)
}

// getenv returns the environment variable named by key, or fallback if unset or empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parsePort converts s to an int port, returning defaultPort on any parse error.
func parsePort(s string) int {
	p, err := strconv.Atoi(s)
	if err != nil || p <= 0 || p > 65535 {
		return defaultPort
	}
	return p
}
