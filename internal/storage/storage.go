package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

type MemStorage struct {
	mu              sync.RWMutex
	logger          *zerolog.Logger
	data            metrics.Data
	interval        int
	storageFileName string
}

func NewMemStorage(logger *zerolog.Logger, interval int, file string) *MemStorage {
	return &MemStorage{
		logger:          logger,
		interval:        interval,
		storageFileName: file,
		data:            make(metrics.Data),
	}
}

func (s *MemStorage) RestoreFromFile() error {
	_, err := os.Stat(s.storageFileName)
	if errors.Is(err, os.ErrNotExist) {
		return os.ErrNotExist
	}
	if err != nil {
		return err
	}
	b, err := os.ReadFile(s.storageFileName)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &s.data); err != nil {
		return err
	}
	s.logger.Info().Msgf("RestoreFromFile: %+v", s.data)
	return nil
}

func (s *MemStorage) WriteToFile() error {
	file, err := os.OpenFile(s.storageFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	s.logger.Info().Msg("File successfully opened")

	b, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	s.logger.Info().Msgf("Data successfully marshalled: %v", string(b))

	n, err := file.Write(b)
	if err != nil {
		return err
	}
	s.logger.Info().Msgf("%d bytes were written to the file", n)
	return nil
}

func (s *MemStorage) LoadAll() metrics.Data {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *MemStorage) Load(mtype, mname string) *metrics.Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, ok := s.data[mtype]
	if !ok {
		s.logger.Info().Msgf("Metric type %s doesn't exist", mtype)
		return nil
	}

	mvalue, ok := metrics[mname]
	if !ok {
		s.logger.Info().Msgf("Metric %v doesn't exist", mvalue)
		return nil
	}

	return &mvalue
}

func (s *MemStorage) Store(m metrics.Metric) bool {
	s.logger.Info().Interface("Start store", s.data).Send()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.interval == 0 {
		defer func() {
			if err := s.WriteToFile(); err != nil {
				s.logger.Error().Err(err).Msg("Failed to write storage content to file")
			}
		}()
	}

	metric, ok := s.data[m.MType]
	if !ok {
		s.data[m.MType] = map[string]metrics.Metric{
			m.ID: {ID: m.ID, MType: m.MType, Value: m.Value, Delta: m.Delta},
		}
		return true
	}

	switch m.MType {
	case metrics.TypeGauge:
		metric[m.ID] = metrics.Metric{ID: m.ID, MType: m.MType, Value: m.Value}
	case metrics.TypeCounter:
		selectedMetric, ok := metric[m.ID]
		if !ok {
			metric[m.ID] = metrics.Metric{ID: m.ID, MType: m.MType, Delta: m.Delta}
			return true
		}
		*selectedMetric.Delta += *m.Delta
		metric[m.ID] = metrics.Metric{ID: m.ID, MType: m.MType, Delta: selectedMetric.Delta}
	}
	s.logger.Info().Interface("Storage content", s.data).Send()

	return true
}
