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
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func NewAgent() (*Config, error) {
	flags := parseAgentFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = flags.ServerAddress
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = flags.ReportInterval
	}
	if config.PollInterval == 0 {
		config.PollInterval = flags.PollInterval
	}

	return &config, nil
}

func NewServer() (Config, error) {
	flags := parseServerFlags()

	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}

	if config.ServerAddress == "" {
		config.ServerAddress = flags.ServerAddress
	}
	if config.Restore == nil {
		config.Restore = flags.Restore
	}
	if config.StoreInterval == nil {
		config.StoreInterval = flags.StoreInterval
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = flags.FileStoragePath
	}
	if config.DatabaseDSN == "" {
		config.DatabaseDSN = flags.DatabaseDSN
	}
	if config.DatabaseDSN != "" {
		config.FileStoragePath = ""
		*config.StoreInterval = -1
		*config.Restore = false
	}

	return config, nil
}

func parseAgentFlags() Config {
	serverAddress := flag.String("a", "localhost:8080", "HTTP server endpoint address")
	reportInterval := flag.Int("r", 10, "report interval to the server (in seconds)")
	pollInterval := flag.Int("p", 2, "interval to gather metrics (in seconds)")
	flag.Parse()
	return Config{
		ServerAddress:  *serverAddress,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
	}
}

func parseServerFlags() Config {
	serverAddress := flag.String("a", "localhost:8080", "address and port to run server")
	fileStoragePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	restore := flag.Bool("r", true, "restore")
	storeInterval := flag.Int("i", 300, "interval")
	databaseDSN := flag.String("d", "", "database DSN")
	flag.Parse()

	return Config{
		ServerAddress:   *serverAddress,
		FileStoragePath: *fileStoragePath,
		DatabaseDSN:     *databaseDSN,
		Restore:         restore,
		StoreInterval:   storeInterval,
	}
}
