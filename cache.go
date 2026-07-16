package main

import (
	"log/slog"
	"sync"
	"time"

	api "github.com/yungwood/synergy-wholesale-exporter/synergywholesaleapi"
)

var (
	listDomainsResponse api.ListDomainsResponse
	cacheExpires        int64
	nextAPIRequestAfter int64
	apiErrorCount       int
	cacheMu             sync.Mutex
)

func getDomains() api.ListDomainsResponse {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	now := time.Now().Unix()
	if cacheExpires > now {
		slog.Debug("Using cached Synergy Wholesale domain response", "cache_expires", cacheExpires)
		return listDomainsResponse
	}
	if nextAPIRequestAfter > now {
		slog.Warn(
			"Skipping Synergy Wholesale API request due to error backoff",
			"next_api_request_after", nextAPIRequestAfter,
			"api_error_count", apiErrorCount,
		)
		return listDomainsResponse
	}

	slog.Info("Sending listDomains request to Synergy Wholesale API", "reseller_id", *resellerID)

	request := api.ListDomainsRequest{
		APIKey:     *apiKey,
		ResellerID: *resellerID,
	}

	response, err := api.Send(request)
	if err != nil {
		backoffSeconds := apiErrorBackoffSeconds(apiErrorCount)
		apiErrorCount++
		nextAPIRequestAfter = now + backoffSeconds
		slog.Error(
			"Error sending SOAP request",
			"error", err,
			"api_error_count", apiErrorCount,
			"backoff_seconds", backoffSeconds,
			"next_api_request_after", nextAPIRequestAfter,
		)
		return listDomainsResponse
	}

	listDomainsResponse = response
	now = time.Now().Unix()
	cacheExpires = now + *cacheTTLSeconds
	nextAPIRequestAfter = 0
	apiErrorCount = 0
	CacheLastSuccessfulRefreshTimestamp.Set(float64(now))
	slog.Info(
		"Updated Synergy Wholesale domain cache",
		"domain_count", len(response.Return.DomainList),
		"cache_expires", cacheExpires,
	)

	return response
}

func apiErrorBackoffSeconds(errorCount int) int64 {
	minBackoff := *apiErrorBackoffMin
	if minBackoff < 1 {
		minBackoff = 1
	}
	maxBackoff := *apiErrorBackoffMax
	if maxBackoff < minBackoff {
		maxBackoff = minBackoff
	}

	backoffSeconds := minBackoff
	for range errorCount {
		if backoffSeconds >= maxBackoff/2 {
			return maxBackoff
		}
		backoffSeconds *= 2
	}
	if backoffSeconds > maxBackoff {
		return maxBackoff
	}
	return backoffSeconds
}
