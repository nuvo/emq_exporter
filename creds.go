package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/prometheus/common/log"
)

//LoadCreds tries to load credentials in the follwing precedence:
//1. Env vars - EMQ_USERNAME && EMQ_PASSWORD
//2. A file under the specified path
//returns the found username and password or error
func LoadCreds(path string) (string, string, error) {
	log.Debugln("Trying to load credentails from environment")
	u, p, err := loadFromEnv()

	if err != nil {
		log.Debugln(err)
		log.Debugln("Trying to load credentails from file")
		return loadFromFile(path)
	}

	return u, p, nil
}

//Try to find auth details in env vars
func loadFromEnv() (string, string, error) {
	var u, p string
	var ok bool

	u, ok = os.LookupEnv("EMQ_USERNAME")
	if !ok {
		return u, p, fmt.Errorf("Can't find EMQ_USERNAME")
	}

	p, ok = os.LookupEnv("EMQ_PASSWORD")
	if !ok {
		return u, p, fmt.Errorf("Can't find EMQ_PASSWORD")
	}

	return u, p, nil
}

//Try to load auth details from file
func loadFromFile(path string) (string, string, error) {
	type creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var err error

	c := &creds{}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return c.Username, c.Password, err
	}

	f, err := ioutil.ReadFile(absPath)
	if err != nil {
		return c.Username, c.Password, err
	}

	if err := json.Unmarshal(f, c); err != nil {
		return c.Username, c.Password, err
	}

	if c.Username == "" {
		err = fmt.Errorf("missing username in %s", path)
	}
	if c.Password == "" {
		err = fmt.Errorf("missing password in %s", path)
	}

	return c.Username, c.Password, err
}
