package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var version = "development"

var (
	resellerID           = flag.String("reseller-id", "", "Synergy Wholesale Reseller ID")
	apiKey               = flag.String("apikey", "", "Synergy Wholesale API Key")
	listenAddress        = flag.String("address", ":8080", "listening port for exporter")
	printVersion         = flag.Bool("version", false, "print version and exit")
	debugLogging         = flag.Bool("debug", false, "enable debug logging")
	jsonLogging          = flag.Bool("json", false, "output logs in JSON format")
	cacheTTLSeconds      = flag.Int64("cache-ttl", 3600, "cache TTL value in seconds for Synergy Wholesale API requests")
	disableGolangMetrics = flag.Bool("no-golang-metrics", false, "disable the default golang prometheus collectors")
)

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
	// enable default collectors
	if !*disableGolangMetrics {
		prometheusRegistry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}
	// add build_info metric
	BuildInfo.WithLabelValues(version, runtime.Version()).Set(1)
	collector := newCollector()
	prometheusRegistry.MustRegister(BuildInfo, collector)
	http.Handle("/metrics", promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{}))

	// add a readiness and liveness check endpoint (return blank 200 OK response)
	http.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {})

	// start webserver
	slog.Info(
		"Starting web server",
		"listen_address", *listenAddress,
		"cache_ttl_seconds", *cacheTTLSeconds,
		"golang_metrics_enabled", !*disableGolangMetrics,
		"debug_logging", *debugLogging,
		"json_logging", *jsonLogging,
	)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		slog.Error("Error starting web server", "error", err)
	}
}
