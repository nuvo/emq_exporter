package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/prometheus/common/log"
)

const (
	usernameEnv = "EMQ_USERNAME"
	passwordEnv = "EMQ_PASSWORD"
)

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

//Try to find auth details in env vars
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

//Try to load auth details from file
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
