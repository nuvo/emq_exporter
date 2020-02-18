package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/common/log"
)

const timeout = 5 * time.Second

var (
	targetsV2 = map[string]string{
		"monitoring_metrics": "/api/v2/monitoring/metrics/%s",
		"monitoring_stats":   "/api/v2/monitoring/stats/%s",
		"monitoring_nodes":   "/api/v2/monitoring/nodes/%s",
		"management_nodes":   "/api/v2/management/nodes/%s",
	}
	//scraping endpoints for EMQ v3 api version
	targetsV3 = map[string]string{
		"nodes_metrics": "/api/v3/nodes/%s/metrics/",
		"nodes_stats":   "/api/v3/nodes/%s/stats/",
		"nodes":         "/api/v3/nodes/%s",
	}
	targetsV4 = map[string]string{
		"nodes_metrics": "/api/v4/nodes/%s/metrics/",
		"nodes_stats":   "/api/v4/nodes/%s/stats/",
		"nodes":         "/api/v4/nodes/%s",
	}
)

type emqResponse struct {
	Code   float64                `json:"code,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"` //api v2 json key
	Data   map[string]interface{} `json:"data,omitempty"`   //api v3 json key
}

//Client manages communication with emq api
type Client struct {
	hc         *http.Client
	host       string
	node       string
	apiVersion string
	targets    map[string]string
	username   string
	password   string
}

//NewClient returns a new emq client
func NewClient(host, node, apiVersion, username, password string) *Client {

	c := &Client{
		hc:         &http.Client{Timeout: timeout},
		host:       host,
		node:       node,
		apiVersion: apiVersion,
		username:   username,
		password:   password,
	}

	switch apiVersion {
	case "v2":
		c.targets = targetsV2
	case "v3":
		c.targets = targetsV3
	case "v4":
		c.targets = targetsV4
	}

	return c
}

//Fetch gets all the metrics from the emq api listed in the targets map
//implements emq_exporter.Fetcher
func (c *Client) Fetch() (map[string]interface{}, error) {

	data := make(map[string]interface{})

	for name, path := range c.targets {

		res, err := c.get(path)
		if err != nil {
			return nil, err
		}

		for k, v := range res {
			mName := fmt.Sprintf("%s_%s", name, strings.Replace(k, "/", "_", -1))
			data[mName] = v
		}
	}

	return data, nil
}

//set the host name for the client (mostly for testing purposes)
func (c *Client) setHost(host string) {
	c.host = host
}

//get preforms an http GET call to the provided path and returns the response
func (c *Client) get(path string) (map[string]interface{}, error) {

	req, err := c.newRequest(path)
	if err != nil {
		return nil, err
	}

	er := &emqResponse{}
	data := make(map[string]interface{})

	res, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to get metrics: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received status code not ok %s, got %d", req.URL, res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(er); err != nil {
		return nil, fmt.Errorf("Error in json decoder %v", err)
	}

	if er.Code != 0 {
		return nil, fmt.Errorf("Recvied code != 0 from EMQ %f", er.Code)
	}

	//Print the returned response data for debuging
	log.Debugf("%#v", *er)

	if c.apiVersion == "v2" {
		data = er.Result
	} else {
		data = er.Data
	}

	return data, nil
}

//newRequest creates a new http request, setting the relevant headers
func (c *Client) newRequest(path string) (req *http.Request, err error) {

	u := c.host + fmt.Sprintf(path, c.node)

	if !strings.Contains(u, "://") {
		u = fmt.Sprintf("http://%s", u)
	}

	log.Debugln("Fetching from", u)

	req, err = http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		log.Debugf("Failed to create http request: %v", err)
		return req, fmt.Errorf("Failed to create http request: %v", err)
	}

	//set request headers
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/json")

	return
}
