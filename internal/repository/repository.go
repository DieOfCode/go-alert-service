package repository

import (
	"fmt"
	"time"

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
	Store(m metrics.Metric) bool
	StoreMetrics(m []metrics.Metric) bool
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

	err := s.Retry(3, func() bool {
		m = s.repo.Load(mtype, mname)
		if m != nil {
			return true
		}
		return false
	}, 1*time.Second, 3*time.Second, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to load metric %s", mname)
	}

	return m, nil
}

func (s *Repository) GetMetrics() (metrics.Data, error) {
	var m metrics.Data
	err := s.Retry(3, func() bool {
		m = s.repo.LoadAll()
		if m != nil {
			return true
		}
		return false
	}, 1*time.Second, 3*time.Second, 5*time.Second)
	if err != nil {
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
	if ok := s.repo.Store(m); !ok {
		return ErrStoreData
	}
	logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Repository) SaveMetrics(m []metrics.Metric) error {
	if ok := s.repo.StoreMetrics(m); !ok {
		return ErrStoreData
	}
	s.logger.Info().Msg("Metric is stored")

	return nil
}

func (s *Repository) Retry(maxRetries int, fn func() bool, intervals ...time.Duration) error {
	var ok bool
	ok = fn()
	if ok {
		return nil
	}
	for i := 0; i < maxRetries; i++ {
		s.logger.Info().Msgf("Retrying... (Attempt %d)", i+1)
		time.Sleep(intervals[i])
		if ok = fn(); ok {
			return nil
		}
	}
	s.logger.Error().Msg("Retrying... Failed")
	return errors.New("err")
}
