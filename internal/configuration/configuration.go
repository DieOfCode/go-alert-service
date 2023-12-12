package configuration

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ServerAddress  string `env:"SERVER_ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func AgentConfiguration() *Config {
	address, poolInterval, reportInterval := parseAgentFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	config.ServerAddress = orDefault(config.ServerAddress, address)
	config.ReportInterval = orDefaultInt(config.ReportInterval, reportInterval)
	config.PollInterval = orDefaultInt(config.PollInterval, poolInterval)

	return &config
}

func ServerConfiguration() *Config {
	serverAddress := parseServerFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		log.Fatal(err)
	}

	config.ServerAddress = orDefault(config.ServerAddress, serverAddress)

	return &config

}

func parseServerFlags() string {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	flag.Parse()
	return *serverAddress
}

func parseAgentFlags() (string, int, int) {
	reportInterval := flag.Int("r", 10, "report interval period in seconds")
	poolInterval := flag.Int("p", 2, "pool interval period in seconds")
	adress := flag.String("a", "localhost:8080", "HTTP address")
	flag.Parse()
	return *adress, *poolInterval, *reportInterval
}

func orDefault(currentValue, defaultValue string) string {
	if currentValue == "" {
		return defaultValue
	}
	return currentValue
}

func orDefaultInt(currentValue, defaultValue int) int {
	if currentValue == 0 {
		return defaultValue
	}
	return currentValue
}
