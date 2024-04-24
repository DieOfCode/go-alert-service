package repository

import (
	"fmt"
	"time"

	"github.com/DieOfCode/go-alert-service/internal/metrics"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const maxRetries = 3

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
	Store(m metrics.Metric) error
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

	err := s.Retry(maxRetries, func() error {
		if err := s.repo.Store(m); err != nil {
			return err
		}
		return nil
	}, 1*time.Second, 3*time.Second, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Repository) SaveMetrics(m []metrics.Metric) error {
	err := s.Retry(maxRetries, func() error {
		if err := s.repo.StoreMetrics(m); err != nil {
			return err
		}
		return nil
	}, 1*time.Second, 3*time.Second, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}
	s.logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Repository) Retry(maxRetries int, fn func() error, intervals ...time.Duration) error {
	var err error
	err = fn()
	if err == nil {
		return nil
	}
	for i := 0; i < maxRetries; i++ {
		s.logger.Info().Msgf("Retrying... (Attempt %d)", i+1)
		time.Sleep(intervals[i])
		if err = fn(); err == nil {
			return nil
		}
	}
	s.logger.Error().Msg("Retrying... Failed")
	return err
}
