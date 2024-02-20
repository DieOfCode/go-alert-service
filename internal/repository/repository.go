package repository

import (
	"fmt"

	"github.com/DieOfCode/go-alert-service/internal/metrics"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

var (
	ErrParseMetric = errors.New("failed to parse metric: wrong type")
	ErrStoreData   = errors.New("failed to store data")
)

type Repository struct {
	logger *zerolog.Logger
	repo   Storage
}

type Storage interface {
	Load(mtype, mname string) *metrics.Metric
	LoadAll() metrics.Data
	StoreMetric(m metrics.Metric) error
	StoreMetrics(m []metrics.Metric) error
	RestoreFromFile() error
	WriteToFile() error
}

func New(l *zerolog.Logger, repo Storage) *Repository {
	return &Repository{
		logger: l,
		repo:   repo,
	}
}

func (s *Repository) GetMetric(mtype, mname string) (*metrics.Metric, error) {
	m := s.repo.Load(mtype, mname)

	if m == nil {
		return nil, fmt.Errorf("failed to load metric %s", mname)
	}

	return m, nil
}

func (s *Repository) GetMetrics() (metrics.Data, error) {
	m := s.repo.LoadAll()
	if m == nil {
		return nil, errors.New("failed to load metrics")
	}

	return m, nil
}

func (s *Repository) SaveMetric(m metrics.Metric) error {
	logger := s.logger.With().
		Str("type", m.MType).
		Str("name", m.ID).
		Logger()
	print("Metric try to store")
	if err := s.repo.StoreMetric(m); err != nil {
		return ErrStoreData
	}
	logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Repository) SaveMetrics(m []metrics.Metric) error {
	if err := s.repo.StoreMetrics(m); err != nil {
		return ErrStoreData
	}
	s.logger.Info().Msg("Metric is stored")

	return nil
}
