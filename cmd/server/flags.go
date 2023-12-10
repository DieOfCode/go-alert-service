package main

import "flag"

var serverAddress string

func parseFlags() {
	flag.StringVar(&serverAddress, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}
