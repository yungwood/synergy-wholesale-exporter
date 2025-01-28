package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	api "github.com/yungwood/synergy-wholesale-exporter/synergywholesaleapi"
)

var version = "development"

var (
	listDomainsResponse api.ListDomainsResponse
	cacheExpires        int64
)

var (
	resellerID      = flag.String("reseller-id", "", "Synergy Wholesale Reseller ID")
	apiKey          = flag.String("apikey", "", "Synergy Wholesale API Key")
	listenAddress   = flag.String("address", ":8080", "listening port for api")
	printVersion    = flag.Bool("version", false, "print version and exit")
	debugLogging    = flag.Bool("debug", false, "enable debug logging")
	jsonLogging     = flag.Bool("json", false, "output logging in JSON format")
	cacheTTLSeconds = flag.Int64("ttl", 3600, "cache TTL value in seconds")
)

var (
	BuildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "build_info",
			Help: "Application build information",
		},
		[]string{"version", "goversion"},
	)
)

type Collector struct {
	domainAutoRenew      *prometheus.Desc
	domainExpiry         *prometheus.Desc
	domainNameServer     *prometheus.Desc
	domainDNSSECKeyCount *prometheus.Desc
}

func newCollector() *Collector {
	return &Collector{
		domainAutoRenew: prometheus.NewDesc("domain_auto_renew_enable",
			"Domain auto-renewal status",
			[]string{"domain"},
			nil,
		),
		domainDNSSECKeyCount: prometheus.NewDesc("domain_dnssec_key_count",
			"Number of DNSSEC keys per domain",
			[]string{"domain"},
			nil,
		),
		domainExpiry: prometheus.NewDesc("domain_expiry_timestamp_seconds",
			"Domain expiry timestamp in seconds",
			[]string{"domain", "status"},
			nil,
		),
		domainNameServer: prometheus.NewDesc("domain_name_server_info",
			"Domain name server info",
			[]string{"domain", "name_server_info"},
			nil,
		),
	}
}

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.domainAutoRenew
	ch <- collector.domainExpiry
	ch <- collector.domainNameServer
}

func (collector *Collector) Collect(ch chan<- prometheus.Metric) {

	listDomainsResp := getDomains()
	for _, domain := range listDomainsResp.Return.DomainList {

		// skip domains where api status != OK they are usually old/deleted
		if domain.Status != "OK" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			collector.domainAutoRenew,
			prometheus.GaugeValue,
			float64(domain.AutoRenew),
			domain.DomainName,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.domainDNSSECKeyCount,
			prometheus.GaugeValue,
			float64(len(domain.DNSSECKeys)),
			domain.DomainName,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.domainExpiry,
			prometheus.GaugeValue,
			float64(domain.GetDomainExpiryTimestamp()),
			domain.DomainName, domain.DomainStatus,
		)

		for _, server := range domain.NameServers {
			ch <- prometheus.MustNewConstMetric(
				collector.domainNameServer,
				prometheus.GaugeValue,
				float64(1),
				domain.DomainName, server,
			)
		}
	}
}

func getDomains() api.ListDomainsResponse {
	if cacheExpires > time.Now().Unix() {
		return listDomainsResponse
	}

	slog.Info("Sending listDomains request to Synergy Wholesale API", "reseller_id", *resellerID)

	request := api.ListDomainsRequest{
		ApiKey:     *apiKey,
		ResellerID: *resellerID,
	}

	data, err := api.SendSOAPRequest(request)
	if err != nil {
		fmt.Printf("Error sending SOAP request: %v\n", err)
		return api.ListDomainsResponse{}
	}

	// Prepare the response struct
	var response api.ListDomainsResponse

	// Unmarshal the response
	err2 := api.UnmarshalSOAPResponse(data, &response)
	if err2 != nil {
		fmt.Printf("Error: %v\n", err2)
		return api.ListDomainsResponse{}
	}

	listDomainsResponse = response
	cacheExpires = time.Now().Unix() + *cacheTTLSeconds

	return response
}

func main() {

	// parse command-line args
	flag.Parse()

	// print version and exit
	if *printVersion {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	// setup logging options
	loggingLevel := slog.LevelInfo // default loglevel
	if *debugLogging {
		loggingLevel = slog.LevelDebug // debug logging enabled
	}
	opts := &slog.HandlerOptions{
		Level: loggingLevel,
	}

	// create json or text logger based on args
	var logger *slog.Logger
	if *jsonLogging {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	slog.SetDefault(logger)
	if *resellerID == "" {
		*resellerID = os.Getenv("SYNERGY_WHOLESALE_RESELLER_ID")
	}

	// check for required parameters
	if *resellerID == "" {
		slog.Error("Reseller ID not set!")
		os.Exit(1)
	}
	if *apiKey == "" {
		*apiKey = os.Getenv("SYNERGY_WHOLESALE_API_KEY")
	}
	if *apiKey == "" {
		slog.Error("API Key not set!")
		os.Exit(1)
	}

	// setup exporter
	prometheusRegistry := prometheus.NewRegistry()
	BuildInfo.WithLabelValues(version, runtime.Version()).Set(1)
	collector := newCollector()
	prometheusRegistry.MustRegister(BuildInfo, collector)
	http.Handle("/metrics", promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{}))

	// add a readiness and liveness check endpoint (return blank 200 OK response)
	http.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {})

	// start webserver
	slog.Info("Starting web server", "listen_address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		slog.Error("Error starting web server", "error", err)
	}
}
