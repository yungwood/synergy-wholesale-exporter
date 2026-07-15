package main

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
)

var BuildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "build_info",
		Help: "Application build information",
	},
	[]string{"version", "goversion"},
)

type Collector struct {
	domainAutoRenew     *prometheus.Desc
	domainExpiry        *prometheus.Desc
	domainNameServer    *prometheus.Desc
	domainDNSSECKeyInfo *prometheus.Desc
}

func newCollector() *Collector {
	return &Collector{
		domainAutoRenew: prometheus.NewDesc("domain_auto_renew_enable",
			"Domain auto-renewal status",
			[]string{"domain"},
			nil,
		),
		domainDNSSECKeyInfo: prometheus.NewDesc("domain_dnssec_key_info",
			"Domain DNSSEC key info",
			[]string{"domain", "key_tag", "algorithm", "digest_type", "digest"},
			nil,
		),
		domainExpiry: prometheus.NewDesc("domain_expiry_timestamp_seconds",
			"Domain expiry timestamp in seconds",
			[]string{"domain", "status"},
			nil,
		),
		domainNameServer: prometheus.NewDesc("domain_name_server_info",
			"Domain name server info",
			[]string{"domain", "name_server"},
			nil,
		),
	}
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.domainAutoRenew
	ch <- collector.domainDNSSECKeyInfo
	ch <- collector.domainExpiry
	ch <- collector.domainNameServer
}

func (collector *Collector) Collect(channel chan<- prometheus.Metric) {
	listDomainsResp := getDomains()
	for _, domain := range listDomainsResp.Return.DomainList {
		// skip domains where api status != OK they are usually old/deleted
		if domain.Status != "OK" {
			slog.Debug(
				"Skipping domain with non-OK API status",
				"domain", domain.DomainName,
				"status", domain.Status,
				"error_message", domain.ErrorMessage,
			)
			continue
		}

		channel <- prometheus.MustNewConstMetric(
			collector.domainAutoRenew,
			prometheus.GaugeValue,
			float64(domain.AutoRenew),
			domain.DomainName,
		)
		for _, dnsSECKey := range domain.DNSSECKeys {
			channel <- prometheus.MustNewConstMetric(
				collector.domainDNSSECKeyInfo,
				prometheus.GaugeValue,
				float64(1),
				domain.DomainName, dnsSECKey.KeyTag, dnsSECKey.Algorithm, dnsSECKey.DigestType, dnsSECKey.Digest,
			)
		}
		channel <- prometheus.MustNewConstMetric(
			collector.domainExpiry,
			prometheus.GaugeValue,
			float64(domain.GetDomainExpiryTimestamp()),
			domain.DomainName, domain.DomainStatus,
		)

		for _, server := range domain.NameServers {
			channel <- prometheus.MustNewConstMetric(
				collector.domainNameServer,
				prometheus.GaugeValue,
				float64(1),
				domain.DomainName, server,
			)
		}
	}
}
