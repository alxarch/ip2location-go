package ip2location

import (
	"bytes"
	"io/ioutil"
	"math/big"
)

const api_version string = "8.0.3"

// get api version
func ApiVersion() string {
	return api_version
}

// initialize the component with the database path
func OpenDB(dbpath string) (*DB, error) {
	data, err := ioutil.ReadFile(dbpath)
	if err != nil {
		return nil, err
	}
	return NewDB(bytes.NewReader(data))
}

var bigOne = big.NewInt(1)

type IPType int

const (
	IPv4 IPType = 4
	IPv6 IPType = 6
)
