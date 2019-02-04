package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"code.cloudfoundry.org/bytefmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	usernameEnv = "EMQ_USERNAME"
	passwordEnv = "EMQ_PASSWORD"
)

//Try to parse value from string to float64, return error on failure
func parseString(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)

	if err != nil {
		//try to convert to bytes
		u, err := bytefmt.ToBytes(s)
		if err != nil {
			log.Debugln("can't parse", s, err)
			return v, err
		}
		v = float64(u)
	}

	return v, nil
}

//newDesc returns a Prometheus description from a metric
func newDesc(m metric) *prometheus.Desc {
	return prometheus.NewDesc(m.name, m.help, nil, nil)
}

//neMetric returns a Prometheus metric from a metric
func newMetric(m metric) (prometheus.Metric, error) {
	return prometheus.NewConstMetric(newDesc(m), m.kind, m.value)
}

//findCreds tries to find credentials in the follwing precedence:
//1. Env vars - EMQ_USERNAME && EMQ_PASSWORD
//2. A file under the specified path
//returns the found username and password or error
func findCreds(path string) (u, p string, err error) {
	log.Debugln("Loading credentails")
	u, p, err = loadFromEnv()

	if err != nil {
		log.Debugln(err)
		return loadFromFile(path)
	}

	return
}

//loadFromEnv tries to find auth details in env vars
func loadFromEnv() (u, p string, err error) {
	log.Debugln("Trying to load credentails from environment")
	var ok bool

	u, ok = os.LookupEnv(usernameEnv)
	if !ok {
		err = fmt.Errorf("Can't find %s", usernameEnv)
		return
	}

	p, ok = os.LookupEnv(passwordEnv)
	if !ok {
		err = fmt.Errorf("Can't find %s", passwordEnv)
		return
	}

	return
}

//loadFromFile tries to load auth details from a file
func loadFromFile(path string) (u, p string, err error) {
	log.Debugln("Trying to load credentails from file")
	var data map[string]string

	absPath, ferr := filepath.Abs(path)
	if ferr != nil {
		log.Debugln(ferr)
		err = ferr
		return
	}

	f, rerr := ioutil.ReadFile(absPath)
	if rerr != nil {
		log.Debugln(rerr)
		err = rerr
		return
	}

	if jerr := json.Unmarshal(f, &data); jerr != nil {
		log.Debugln(jerr)
		err = jerr
		return
	}

	u = data["username"]
	if u == "" {
		err = fmt.Errorf("missing username in %s", path)
	}

	p = data["password"]
	if p == "" {
		err = fmt.Errorf("missing password in %s", path)
	}

	return
}
