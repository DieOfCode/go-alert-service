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

type Service struct {
	logger *zerolog.Logger
	repo   Repository
}

type Repository interface {
	Load(mtype, mname string) *metrics.Metric
	LoadAll() metrics.Data
	Store(m metrics.Metric) bool
}

func New(l *zerolog.Logger, repo Repository) *Service {
	return &Service{
		logger: l,
		repo:   repo,
	}
}

func (s *Service) GetMetric(mtype, mname string) (*metrics.Metric, error) {
	m := s.repo.Load(mtype, mname)
	if m == nil {
		return nil, fmt.Errorf("failed to load metric %s", mname)
	}

	return m, nil
}

func (s *Service) GetMetrics() (metrics.Data, error) {
	m := s.repo.LoadAll()
	if m == nil {
		return nil, errors.New("failed to load metrics")
	}

	return m, nil
}

func (s *Service) SaveMetric(m metrics.Metric) error {
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
