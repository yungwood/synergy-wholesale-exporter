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
	t.Setenv(envDebug, "true")
	t.Setenv(envJSON, "true")
	t.Setenv(envGolangMetrics, "true")
	t.Setenv(envDNSSECMetrics, "true")

	mustSetFlag(t, flags, "reseller-id", "flag-reseller")
	mustSetFlag(t, flags, "apikey", "flag-api-key")
	mustSetFlag(t, flags, "address", ":8081")
	mustSetFlag(t, flags, "cache-ttl", "300")
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
	var debugLogging bool
	var jsonLogging bool
	var enableGolangMetrics bool
	var enableDNSSECMetrics bool

	flags.StringVar(&resellerID, "reseller-id", resellerID, "")
	flags.StringVar(&apiKey, "apikey", apiKey, "")
	flags.StringVar(&listenAddress, "address", listenAddress, "")
	flags.Int64Var(&cacheTTLSeconds, "cache-ttl", cacheTTLSeconds, "")
	flags.BoolVar(&debugLogging, "debug", debugLogging, "")
	flags.BoolVar(&jsonLogging, "json", jsonLogging, "")
	flags.BoolVar(&enableGolangMetrics, "golang-metrics", enableGolangMetrics, "")
	flags.BoolVar(&enableDNSSECMetrics, "dnssec-metrics", enableDNSSECMetrics, "")

	return flags, envConfig{
		resellerID:          &resellerID,
		apiKey:              &apiKey,
		listenAddress:       &listenAddress,
		cacheTTLSeconds:     &cacheTTLSeconds,
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
