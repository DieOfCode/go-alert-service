package configuration

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"ADDRESS"`
	ReportInterval  int    `env:"REPORT_INTERVAL"`
	PollInterval    int    `env:"POLL_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
}

func NewAgent() (*Config, error) {
	srvAddr, reportInt, pollInt := parseAgentFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = srvAddr
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = reportInt
	}
	if config.PollInterval == 0 {
		config.PollInterval = pollInt
	}


	return &config, nil
}

func NewServer() (*Config, error) {
	srvAddr, fileStoragePath, restore, storeInterval := parseServerFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = srvAddr
	}
	if config.Restore == nil {
		config.Restore = &restore
	}
	if config.StoreInterval == nil {
		config.StoreInterval = &storeInterval
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = fileStoragePath
	}

	return &config, nil
}

func parseAgentFlags() (string, int, int) {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 10, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 2, "interval to gather metrics (in seconds)")
	flag.Parse()

	return *serverAddress, *reportInterval, *pollInterval
}

func parseServerFlags() (string, string, bool, int) {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	restore := flag.Bool("r", true, "restore")
	storeInterval := flag.Int("i", 300, "interval")
	flag.Parse()

	return *serverAddress, *fileStoragePath, *restore, *storeInterval
}
