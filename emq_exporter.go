package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/nuvo/emq_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	namespace = "emq"
)

var (
	up = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "up",
		Help:      "Was the last scrape of EMQ successful",
	})

	totalScrapes = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "exporter_total_scrapes",
		Help:      "Current total scrapes.",
	})

	//GitTag stands for a git tag, populated at build time
	GitTag string
	//GitCommit stands for a git commit hash, populated at build time
	GitCommit string
)

//Fetcher knows how to fetch metrics from emq
type Fetcher interface {
	Fetch() (map[string]interface{}, error)
}

//metric is an internal representation of a metric before being processed
//and sent to prometheus
type metric struct {
	kind  prometheus.ValueType
	value float64
	name  string
	help  string
}

// Exporter collects EMQ stats from the given host and exports them using
// the prometheus metrics package.
type Exporter struct {
	fetcher Fetcher
	mu      *sync.Mutex
	metrics []*metric
}

// NewExporter returns an initialized Exporter.
func NewExporter(fetcher Fetcher) *Exporter {
	return &Exporter{
		fetcher: fetcher,
		mu:      &sync.Mutex{},
	}
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	var err error

	if err = e.scrape(); err != nil {
		log.Warnln(err)
	}

	//Send the metrics to the channel
	e.mu.Lock()

	if err != nil {
		up.Set(0)
	} else {
		up.Set(1)
	}
	ch <- up

	totalScrapes.Inc()
	ch <- totalScrapes

	metricList := make([]metric, 0, len(e.metrics))
	for _, i := range e.metrics {
		metricList = append(metricList, *i)
	}
	e.mu.Unlock()

	for _, i := range metricList {
		i.name = strings.Replace(i.name, ".", "_", -1)
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
	ch <- up.Desc()
	ch <- totalScrapes.Desc()
}

// get the json responses from the targets map, process them and
// insert them into exporter.metrics array
func (e *Exporter) scrape() error {
	data, err := e.fetcher.Fetch()
	if err != nil {
		return err
	}

	for k, v := range data {
		fqName := fmt.Sprintf("%s_%s", namespace, k)
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

	return nil
}

//add adds a metric to the exporter.metrics array
func (e *Exporter) add(fqName, help string, value float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	//check if the metric with a given fqName exists
	for _, v := range e.metrics {
		if strings.Contains(newDesc(*v).String(), fqName) {
			v.value = value
			return
		}
	}

	//append it to the e.metrics array
	e.metrics = append(e.metrics, &metric{
		kind:  prometheus.GaugeValue,
		name:  fqName,
		help:  help,
		value: value,
	})

	return
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9540").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		emqURI        = kingpin.Flag("emq.uri", "HTTP API address of the EMQ node.").Default("http://127.0.0.1:18083").Short('u').String()
		emqCreds      = kingpin.Flag("emq.creds-file", "Path to json file containing emq credentials").Default("./auth.json").Short('f').String()
		emqNodeName   = kingpin.Flag("emq.node", "Node name of the emq node to scrape.").Default("emq@127.0.0.1").Short('n').String()
		emqAPIVersion = kingpin.Flag("emq.api-version", "The API version used by EMQ. Valid values: [v2, v3, v4]").Default("v3").Enum("v2", "v3", "v4")
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

	if *emqAPIVersion == "v2" {
		log.Warnln("v2 api version is deprecated and will be removed in future versions")
	}

	c := client.NewClient(*emqURI, *emqNodeName, *emqAPIVersion, username, password)

	exporter := NewExporter(c)

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
