package main

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type metric struct {
	kind  prometheus.ValueType
	value float64
	name  string
	help  string
}

type emqResponse struct {
	Code   float64                `json:"code,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"` //api v2 json key
	Data   map[string]interface{} `json:"data,omitempty"`   //api v3 json key
}

type config struct {
	host       string
	username   string
	password   string
	node       string
	apiVersion string
}

// Exporter collects EMQ stats from the given host and exports them using
// the prometheus metrics package.
type Exporter struct {
	config *config
	client *http.Client

	mu      *sync.Mutex
	metrics []*metric
}
