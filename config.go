package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	envResellerID    = "SYNERGY_WHOLESALE_RESELLER_ID"
	envAPIKey        = "SYNERGY_WHOLESALE_API_KEY"
	envAddress       = "SYNERGY_WHOLESALE_EXPORTER_ADDRESS"
	envCacheTTL      = "SYNERGY_WHOLESALE_EXPORTER_CACHE_TTL"
	envDebug         = "SYNERGY_WHOLESALE_EXPORTER_DEBUG"
	envJSON          = "SYNERGY_WHOLESALE_EXPORTER_JSON"
	envGolangMetrics = "SYNERGY_WHOLESALE_EXPORTER_GOLANG_METRICS"
)

func applyEnvConfig() error {
	setStringFromEnv("reseller-id", resellerID, envResellerID)
	setStringFromEnv("apikey", apiKey, envAPIKey)
	setStringFromEnv("address", listenAddress, envAddress)

	if err := setInt64FromEnv("cache-ttl", cacheTTLSeconds, envCacheTTL); err != nil {
		return err
	}
	if err := setBoolFromEnv("debug", debugLogging, envDebug); err != nil {
		return err
	}
	if err := setBoolFromEnv("json", jsonLogging, envJSON); err != nil {
		return err
	}
	if err := setBoolFromEnv("golang-metrics", enableGolangMetrics, envGolangMetrics); err != nil {
		return err
	}

	return nil
}

func setStringFromEnv(flagName string, target *string, envName string) {
	if wasFlagSet(flagName) {
		return
	}
	if value := os.Getenv(envName); value != "" {
		*target = value
	}
}

func setInt64FromEnv(flagName string, target *int64, envName string) error {
	if wasFlagSet(flagName) {
		return nil
	}
	value := os.Getenv(envName)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid %s value %q: %w", envName, value, err)
	}
	*target = parsed
	return nil
}

func setBoolFromEnv(flagName string, target *bool, envName string) error {
	if wasFlagSet(flagName) {
		return nil
	}
	value := os.Getenv(envName)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid %s value %q: %w", envName, value, err)
	}
	*target = parsed
	return nil
}

func wasFlagSet(name string) bool {
	wasSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			wasSet = true
		}
	})
	return wasSet
}
