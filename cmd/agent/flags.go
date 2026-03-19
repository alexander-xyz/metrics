package main

import (
	"flag"
	"strings"
)

var serverAddr string
var poolInterval int64
var reportInterval int64

func parseFlags() {
	var flagRunAddr string

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Int64Var(&reportInterval, "r", 10, "report interval in seconds")
	flag.Int64Var(&poolInterval, "p", 2, "update pool interval in seconds")
	flag.Parse()

	if !strings.HasPrefix(flagRunAddr, "http") {
		serverAddr = "http://" + flagRunAddr
	} else {
		serverAddr = flagRunAddr
	}
}
