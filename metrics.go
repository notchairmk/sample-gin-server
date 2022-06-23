package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var totalRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "http_egress_requests_total",
		Help: "Number of external requests made",
	},
)
