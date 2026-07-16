package synergywholesaleapi

import "github.com/prometheus/client_golang/prometheus"

const metricNamespace = "synergy_wholesale"

var SynergyWholesaleAPIRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: metricNamespace,
		Name:      "api_requests_total",
		Help:      "Total number of HTTP requests sent to the Synergy Wholesale API.",
	},
	[]string{"code", "result"},
)
