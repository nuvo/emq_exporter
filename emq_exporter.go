package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/nuvo/emq_exporter/internal/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
		log.Warn().Msg(err.Error())
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
			log.Error().Msg("newMetric: " + err.Error())
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
			log.Debug().Msg(k + " is of type I don't know how to handle")
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

	emqAPIVersion := flag.String("emq.api-version", "v3", "The API version used by EMQ. Valid values: [v2, v3, v4]")
	emqCreds := flag.String("emq.creds-file", "./auth.json", "Path to json file containing emq credentials")
	emqNodeName := flag.String("emq.node", "emq@127.0.0.1", "Node name of the emq node to scrape")
	emqURI := flag.String("emq.uri", "http://127.0.0.1:18083", "HTTP API address of the EMQ node")
	debug := flag.Bool("debug", false, "sets log level to debug")
	webListenAddress := flag.String("web.listen-address", ":9540", "Address to listen on for web interface and telemetry")
	webMetricsPath := flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics")

	flag.Parse()

	//log configs
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.With().Caller().Logger()

	log.Info().Msg("Loading authentication credentials")

	username, password, err := findCreds(*emqCreds)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load credentials:")
	}

	log.Info().Msg("Starting emq_exporter")
	log.Info().Msgf("Version %s (git-%s)", GitTag, GitCommit)

	switch *emqAPIVersion {
	case "v2":
		log.Warn().Msg("v2 api version is deprecated and will be removed in future versions")
	case "v3", "v4":
	default:
		log.Fatal().Err(errors.New("unsupported api version")).Msg("unsupported api version: " + *emqAPIVersion)
	}

	c := client.NewClient(*emqURI, *emqNodeName, *emqAPIVersion, username, password)

	exporter := NewExporter(c)

	prometheus.MustRegister(exporter)

	log.Info().Msg("Listening on " + *webListenAddress)

	http.Handle(*webMetricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>EMQ Exporter</title></head>
             <body>
             <h1>EMQ Exporter</h1>
             <p><a href='` + *webMetricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if err := http.ListenAndServe(*webListenAddress, nil); err != nil {
		log.Fatal().Err(err).Msg("startup failed")
	}

}
