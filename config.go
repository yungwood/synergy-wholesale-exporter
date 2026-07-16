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
	envBackoffMin    = "SYNERGY_WHOLESALE_EXPORTER_API_ERROR_BACKOFF_MIN"
	envBackoffMax    = "SYNERGY_WHOLESALE_EXPORTER_API_ERROR_BACKOFF_MAX"
	envDebug         = "SYNERGY_WHOLESALE_EXPORTER_DEBUG"
	envJSON          = "SYNERGY_WHOLESALE_EXPORTER_JSON"
	envGolangMetrics = "SYNERGY_WHOLESALE_EXPORTER_GOLANG_METRICS"
	envDNSSECMetrics = "SYNERGY_WHOLESALE_EXPORTER_DNSSEC_METRICS"
)

func applyEnvConfig() error {
	return applyEnvConfigFromFlagSet(flag.CommandLine, envConfig{
		resellerID:          resellerID,
		apiKey:              apiKey,
		listenAddress:       listenAddress,
		cacheTTLSeconds:     cacheTTLSeconds,
		apiErrorBackoffMin:  apiErrorBackoffMin,
		apiErrorBackoffMax:  apiErrorBackoffMax,
		debugLogging:        debugLogging,
		jsonLogging:         jsonLogging,
		enableGolangMetrics: enableGolangMetrics,
		enableDNSSECMetrics: enableDNSSECMetrics,
	})
}

type envConfig struct {
	resellerID          *string
	apiKey              *string
	listenAddress       *string
	cacheTTLSeconds     *int64
	apiErrorBackoffMin  *int64
	apiErrorBackoffMax  *int64
	debugLogging        *bool
	jsonLogging         *bool
	enableGolangMetrics *bool
	enableDNSSECMetrics *bool
}

func applyEnvConfigFromFlagSet(flags *flag.FlagSet, config envConfig) error {
	setStringFromEnv(flags, "reseller-id", config.resellerID, envResellerID)
	setStringFromEnv(flags, "apikey", config.apiKey, envAPIKey)
	setStringFromEnv(flags, "address", config.listenAddress, envAddress)

	if err := setInt64FromEnv(flags, "cache-ttl", config.cacheTTLSeconds, envCacheTTL); err != nil {
		return err
	}
	if err := setInt64FromEnv(flags, "api-error-backoff-min", config.apiErrorBackoffMin, envBackoffMin); err != nil {
		return err
	}
	if err := setInt64FromEnv(flags, "api-error-backoff-max", config.apiErrorBackoffMax, envBackoffMax); err != nil {
		return err
	}
	if err := setBoolFromEnv(flags, "debug", config.debugLogging, envDebug); err != nil {
		return err
	}
	if err := setBoolFromEnv(flags, "json", config.jsonLogging, envJSON); err != nil {
		return err
	}
	if err := setBoolFromEnv(flags, "golang-metrics", config.enableGolangMetrics, envGolangMetrics); err != nil {
		return err
	}
	if err := setBoolFromEnv(flags, "dnssec-metrics", config.enableDNSSECMetrics, envDNSSECMetrics); err != nil {
		return err
	}

	return nil
}

func setStringFromEnv(flags *flag.FlagSet, flagName string, target *string, envName string) {
	if wasFlagSet(flags, flagName) {
		return
	}
	if value := os.Getenv(envName); value != "" {
		*target = value
	}
}

func setInt64FromEnv(flags *flag.FlagSet, flagName string, target *int64, envName string) error {
	if wasFlagSet(flags, flagName) {
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

func setBoolFromEnv(flags *flag.FlagSet, flagName string, target *bool, envName string) error {
	if wasFlagSet(flags, flagName) {
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

func wasFlagSet(flags *flag.FlagSet, name string) bool {
	wasSet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == name {
			wasSet = true
		}
	})
	return wasSet
}
