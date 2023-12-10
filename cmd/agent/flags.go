package main

import "flag"

var reportInterval int
var poolInterval int
var addressHttp string

func parseFlags() {
	flag.IntVar(&reportInterval, "r", 10, "report interval period in seconds")
	flag.IntVar(&poolInterval, "p", 2, "pool interval period in seconds")
	flag.StringVar(&addressHttp, "a", "localhost:8080", "HTTP address")
	flag.Parse()
}
