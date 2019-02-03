package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "emq"
)

var (
	//scraping endpoints for EMQ v2 api version
	targetsV2 = map[string]string{
		"monitoring_metrics": "/api/v2/monitoring/metrics/%s",
		"monitoring_stats":   "/api/v2/monitoring/stats/%s",
		"monitoring_nodes":   "/api/v2/monitoring/nodes/%s",
		"management_nodes":   "/api/v2/management/nodes/%s",
	}
	//scraping endpoints for EMQ v3 api version
	targetsV3 = map[string]string{
		"node_metrics": "/api/v3/nodes/%s/metrics/",
		"node_stats":   "/api/v3/nodes/%s/stats/",
		"nodes":        "/api/v3/nodes/%s",
	}

	//GitTag stands for a git tag, populated at build time
	GitTag string
	//GitCommit stands for a git commit hash populated at build time
	GitCommit string
)

//newDesc converts one e.metric to a Prometheus description
func newDesc(m metric) *prometheus.Desc {
	return prometheus.NewDesc(m.name, m.help, nil, nil)
}

//neMetric converts one e.metric to a Prometheus metric
func newMetric(m metric) (prometheus.Metric, error) {
	return prometheus.NewConstMetric(newDesc(m), m.kind, m.value)
}

// NewExporter returns an initialized Exporter.
func NewExporter(c *config, timeout time.Duration) *Exporter {

	return &Exporter{
		config: c,
		client: &http.Client{
			Timeout: timeout,
		},
		mu: &sync.Mutex{},
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of EMQ successful",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total scrapes.",
		}),
	}

}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	var up float64

	if err := e.scrape(); err != nil {
		log.Warnln(err)
	} else {
		e.mu.Lock()
		up = 1
		e.mu.Unlock()
	}

	//Send the metrics to the channel
	e.mu.Lock()

	e.up.Set(up)
	e.totalScrapes.Inc()

	metricList := make([]metric, 0, len(e.metrics))
	for _, i := range e.metrics {
		metricList = append(metricList, *i)
	}
	e.mu.Unlock()

	ch <- e.up
	ch <- e.totalScrapes

	for _, i := range metricList {
		m, err := newMetric(i)
		if err != nil {
			log.Errorf("newMetric: %v", err)
			continue
		}
		ch <- m
	}

}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
}

// get the json responses from the targets map, process them and
// insert them into exporter.metrics array
func (e *Exporter) scrape() error {

	var targets = make(map[string]string)

	if e.config.apiVersion == "v2" {
		targets = targetsV2
	} else {
		targets = targetsV3
	}

	for name, path := range targets {

		data, err := e.fetch(path)
		if err != nil {
			return err
		}

		for k, v := range data {
			fqName := fmt.Sprintf("%s_%s_%s", namespace, name, strings.Replace(k, "/", "_", -1))
			switch vv := v.(type) {
			case string:
				val, err := parseString(vv)
				if err != nil {
					break
				}
				e.add(fqName, k, val)
			case float64:
				e.add(fqName, k, vv)
			default:
				log.Debugln(k, "is of type I don't know how to handle")
			}
		}
	}

	return nil
}

//addMetric adds a metric to the exporter.Metric array
func (e *Exporter) add(fqName, help string, value float64) {
	//check if the metric with a given fqName exists, and update its value
	for _, v := range e.metrics {
		if strings.Contains(newDesc(*v).String(), fqName) {
			e.mu.Lock()
			v.value = value
			e.mu.Unlock()
			return
		}
	}

	//append a new metric to the metrics array
	e.metrics = append(e.metrics, &metric{
		kind:  prometheus.GaugeValue,
		name:  fqName,
		help:  help,
		value: value,
	})

}

//get the response from the provided target url
func (e *Exporter) fetch(target string) (map[string]interface{}, error) {

	data := &emqResponse{}
	response := make(map[string]interface{})

	u := e.config.host + fmt.Sprintf(target, e.config.node)

	log.Debugln("fetching from", u)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return response, fmt.Errorf("Failed to create http request: %v", err)
	}

	req.SetBasicAuth(e.config.username, e.config.password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "emp_exporter/"+GitTag)

	res, err := e.client.Do(req)
	if err != nil {
		return response, fmt.Errorf("Failed to get metrics from %s: %v", u, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return response, fmt.Errorf("Received status code not ok %s", u)
	}

	if err := json.NewDecoder(res.Body).Decode(data); err != nil {
		return response, fmt.Errorf("Error in json decoder %v", err)
	}

	if data.Code != 0 {
		return response, fmt.Errorf("Recvied code != 0 from EMQ %f", data.Code)
	}

	//Print the returned response data for debuging
	log.Debugf("%#v", *data)

	if e.config.apiVersion == "v2" {
		response = data.Result
	} else {
		response = data.Data
	}

	return response, nil
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9540").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		emqURI        = kingpin.Flag("emq.uri", "HTTP API address of the EMQ node.").Default("http://127.0.0.1:18083").Short('u').String()
		emqCreds      = kingpin.Flag("emq.creds-file", "Path to json file containing emq credentials").Default("./auth.json").Short('f').String()
		emqNodeName   = kingpin.Flag("emq.node", "Node name of the emq node to scrape.").Default("emq@127.0.0.1").Short('n').String()
		emqTimeout    = kingpin.Flag("emq.timeout", "Timeout for trying to get stats from emq").Default("5s").Duration()
		emqAPIVersion = kingpin.Flag("emq.api-version", "The API version used by EMQ. Valid values: [v2, v3]").Default("v2").Enum("v2", "v3")
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(fmt.Sprintf("Version %s (git-%s)", GitTag, GitCommit))
	kingpin.CommandLine.HelpFlag.Short('h')

	kingpin.Parse()

	log.Infoln("Loading authentication credentials")

	username, password, err := findCreds(*emqCreds)
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}

	log.Infoln("Starting emq_exporter")
	log.Infof("Version %s (git-%s)", GitTag, GitCommit)

	//common config for use in the exporter
	conf := &config{
		host:       *emqURI,
		username:   username,
		password:   password,
		node:       *emqNodeName,
		apiVersion: *emqAPIVersion,
	}

	exporter := NewExporter(conf, *emqTimeout)

	prometheus.MustRegister(exporter)

	log.Infoln("Listening on", *listenAddress)
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>EMQ Exporter</title></head>
             <body>
             <h1>EMQ Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
