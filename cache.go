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
	cacheMu             sync.Mutex
)

func getDomains() api.ListDomainsResponse {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cacheExpires > time.Now().Unix() {
		slog.Debug("Using cached Synergy Wholesale domain response", "cache_expires", cacheExpires)
		return listDomainsResponse
	}

	slog.Info("Sending listDomains request to Synergy Wholesale API", "reseller_id", *resellerID)

	request := api.ListDomainsRequest{
		APIKey:     *apiKey,
		ResellerID: *resellerID,
	}

	response, err := api.Send(request)
	if err != nil {
		slog.Error("Error sending SOAP request", "error", err)
		return api.ListDomainsResponse{}
	}

	listDomainsResponse = response
	now := time.Now().Unix()
	cacheExpires = now + *cacheTTLSeconds
	CacheLastSuccessfulRefreshTimestamp.Set(float64(now))
	slog.Info(
		"Updated Synergy Wholesale domain cache",
		"domain_count", len(response.Return.DomainList),
		"cache_expires", cacheExpires,
	)

	return response
}
