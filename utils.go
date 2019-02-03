package main

import (
	"strconv"

	"code.cloudfoundry.org/bytefmt"
	"github.com/prometheus/common/log"
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
