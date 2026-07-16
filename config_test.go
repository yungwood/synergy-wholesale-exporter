package main

import (
	"flag"
	"strings"
	"testing"
)

func TestApplyEnvConfigFromFlagSetUsesEnvironmentForUnsetFlags(t *testing.T) {
	flags, config := newTestEnvConfig()
	t.Setenv(envResellerID, "env-reseller")
	t.Setenv(envAPIKey, "env-api-key")
	t.Setenv(envAddress, ":9090")
	t.Setenv(envCacheTTL, "120")
	t.Setenv(envBackoffMin, "30")
	t.Setenv(envBackoffMax, "600")
	t.Setenv(envDebug, "true")
	t.Setenv(envJSON, "true")
	t.Setenv(envGolangMetrics, "true")
	t.Setenv(envDNSSECMetrics, "true")

	if err := applyEnvConfigFromFlagSet(flags, config); err != nil {
		t.Fatalf("apply env config: %v", err)
	}

	if *config.resellerID != "env-reseller" {
		t.Errorf("reseller ID = %q, want env-reseller", *config.resellerID)
	}
	if *config.apiKey != "env-api-key" {
		t.Errorf("API key = %q, want env-api-key", *config.apiKey)
	}
	if *config.listenAddress != ":9090" {
		t.Errorf("listen address = %q, want :9090", *config.listenAddress)
	}
	if *config.cacheTTLSeconds != 120 {
		t.Errorf("cache TTL = %d, want 120", *config.cacheTTLSeconds)
	}
	if *config.apiErrorBackoffMin != 30 {
		t.Errorf("API error backoff min = %d, want 30", *config.apiErrorBackoffMin)
	}
	if *config.apiErrorBackoffMax != 600 {
		t.Errorf("API error backoff max = %d, want 600", *config.apiErrorBackoffMax)
	}
	if !*config.debugLogging {
		t.Error("debug logging = false, want true")
	}
	if !*config.jsonLogging {
		t.Error("JSON logging = false, want true")
	}
	if !*config.enableGolangMetrics {
		t.Error("golang metrics = false, want true")
	}
	if !*config.enableDNSSECMetrics {
		t.Error("DNSSEC metrics = false, want true")
	}
}

func TestApplyEnvConfigFromFlagSetKeepsExplicitFlags(t *testing.T) {
	flags, config := newTestEnvConfig()
	t.Setenv(envResellerID, "env-reseller")
	t.Setenv(envAPIKey, "env-api-key")
	t.Setenv(envAddress, ":9090")
	t.Setenv(envCacheTTL, "120")
	t.Setenv(envBackoffMin, "30")
	t.Setenv(envBackoffMax, "600")
	t.Setenv(envDebug, "true")
	t.Setenv(envJSON, "true")
	t.Setenv(envGolangMetrics, "true")
	t.Setenv(envDNSSECMetrics, "true")

	mustSetFlag(t, flags, "reseller-id", "flag-reseller")
	mustSetFlag(t, flags, "apikey", "flag-api-key")
	mustSetFlag(t, flags, "address", ":8081")
	mustSetFlag(t, flags, "cache-ttl", "300")
	mustSetFlag(t, flags, "api-error-backoff-min", "45")
	mustSetFlag(t, flags, "api-error-backoff-max", "900")
	mustSetFlag(t, flags, "debug", "false")
	mustSetFlag(t, flags, "json", "false")
	mustSetFlag(t, flags, "golang-metrics", "false")
	mustSetFlag(t, flags, "dnssec-metrics", "false")

	if err := applyEnvConfigFromFlagSet(flags, config); err != nil {
		t.Fatalf("apply env config: %v", err)
	}

	if *config.resellerID != "flag-reseller" {
		t.Errorf("reseller ID = %q, want flag-reseller", *config.resellerID)
	}
	if *config.apiKey != "flag-api-key" {
		t.Errorf("API key = %q, want flag-api-key", *config.apiKey)
	}
	if *config.listenAddress != ":8081" {
		t.Errorf("listen address = %q, want :8081", *config.listenAddress)
	}
	if *config.cacheTTLSeconds != 300 {
		t.Errorf("cache TTL = %d, want 300", *config.cacheTTLSeconds)
	}
	if *config.apiErrorBackoffMin != 45 {
		t.Errorf("API error backoff min = %d, want 45", *config.apiErrorBackoffMin)
	}
	if *config.apiErrorBackoffMax != 900 {
		t.Errorf("API error backoff max = %d, want 900", *config.apiErrorBackoffMax)
	}
	if *config.debugLogging {
		t.Error("debug logging = true, want false")
	}
	if *config.jsonLogging {
		t.Error("JSON logging = true, want false")
	}
	if *config.enableGolangMetrics {
		t.Error("golang metrics = true, want false")
	}
	if *config.enableDNSSECMetrics {
		t.Error("DNSSEC metrics = true, want false")
	}
}

func TestApplyEnvConfigFromFlagSetRejectsInvalidValues(t *testing.T) {
	t.Run("cache TTL", func(t *testing.T) {
		flags, config := newTestEnvConfig()
		t.Setenv(envCacheTTL, "invalid")

		err := applyEnvConfigFromFlagSet(flags, config)
		if err == nil {
			t.Fatal("apply env config error = nil, want error")
		}
		if !strings.Contains(err.Error(), envCacheTTL) {
			t.Errorf("error = %q, want %s", err.Error(), envCacheTTL)
		}
	})

	t.Run("debug", func(t *testing.T) {
		flags, config := newTestEnvConfig()
		t.Setenv(envDebug, "invalid")

		err := applyEnvConfigFromFlagSet(flags, config)
		if err == nil {
			t.Fatal("apply env config error = nil, want error")
		}
		if !strings.Contains(err.Error(), envDebug) {
			t.Errorf("error = %q, want %s", err.Error(), envDebug)
		}
	})
}

func newTestEnvConfig() (*flag.FlagSet, envConfig) {
	flags := flag.NewFlagSet("test", flag.ContinueOnError)

	var resellerID string
	var apiKey string
	var listenAddress = ":8080"
	var cacheTTLSeconds int64 = 3600
	var apiErrorBackoffMin int64 = 60
	var apiErrorBackoffMax int64 = 3600
	var debugLogging bool
	var jsonLogging bool
	var enableGolangMetrics bool
	var enableDNSSECMetrics bool

	flags.StringVar(&resellerID, "reseller-id", resellerID, "")
	flags.StringVar(&apiKey, "apikey", apiKey, "")
	flags.StringVar(&listenAddress, "address", listenAddress, "")
	flags.Int64Var(&cacheTTLSeconds, "cache-ttl", cacheTTLSeconds, "")
	flags.Int64Var(&apiErrorBackoffMin, "api-error-backoff-min", apiErrorBackoffMin, "")
	flags.Int64Var(&apiErrorBackoffMax, "api-error-backoff-max", apiErrorBackoffMax, "")
	flags.BoolVar(&debugLogging, "debug", debugLogging, "")
	flags.BoolVar(&jsonLogging, "json", jsonLogging, "")
	flags.BoolVar(&enableGolangMetrics, "golang-metrics", enableGolangMetrics, "")
	flags.BoolVar(&enableDNSSECMetrics, "dnssec-metrics", enableDNSSECMetrics, "")

	return flags, envConfig{
		resellerID:          &resellerID,
		apiKey:              &apiKey,
		listenAddress:       &listenAddress,
		cacheTTLSeconds:     &cacheTTLSeconds,
		apiErrorBackoffMin:  &apiErrorBackoffMin,
		apiErrorBackoffMax:  &apiErrorBackoffMax,
		debugLogging:        &debugLogging,
		jsonLogging:         &jsonLogging,
		enableGolangMetrics: &enableGolangMetrics,
		enableDNSSECMetrics: &enableDNSSECMetrics,
	}
}

func mustSetFlag(t *testing.T, flags *flag.FlagSet, name string, value string) {
	t.Helper()
	if err := flags.Set(name, value); err != nil {
		t.Fatalf("set flag %s: %v", name, err)
	}
}
