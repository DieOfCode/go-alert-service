package storage

import (
	"database/sql"
	"errors"

	"github.com/DieOfCode/go-alert-service/internal/metrics"
	"github.com/rs/zerolog"
)

var errNotSupported = errors.New("not supported when DB is enabled")

type DatabaseStorage struct {
	db     *sql.DB
	logger *zerolog.Logger
}

func NewDatabaseStorage(logger *zerolog.Logger, db *sql.DB) *DatabaseStorage {

	return &DatabaseStorage{
		db:     db,
		logger: logger,
	}
}

func (storage *DatabaseStorage) LoadAll() metrics.Data {

	rows, err := storage.db.Query("SELECT id,type,value,delta FROM metrics")
	if err != nil {
		return nil
	}
	defer rows.Close()

	result := make(metrics.Data)
	for rows.Next() {
		var mID, mType string
		var mValue sql.NullFloat64
		var mDelta sql.NullInt64

		if err := rows.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
			return nil
		}
		_, ok := result[mType]
		if !ok {
			result[mType] = map[string]metrics.Metric{
				mID: {
					ID:    mID,
					MType: mType,
					Delta: parseDelta(mDelta),
					Value: parseValue(mValue),
				},
			}
			continue
		}
		result[mType][mID] = metrics.Metric{
			ID:    mID,
			MType: mType,
			Delta: parseDelta(mDelta),
			Value: parseValue(mValue),
		}
	}
	if err := rows.Err(); err != nil {
		return nil
	}

	return result
}

func (storage *DatabaseStorage) StoreMetrics(metrics []metrics.Metric) error {
	tx, err := storage.db.Begin()
	if err != nil {
		storage.logger.Error().Err(err).Msg("StoreMetrics: begin transaction error")
		return err
	}

	for _, metric := range metrics {
		err = storage.store(tx, metric)
		if err != nil {
			storage.logger.Error().Err(err).Msg("StoreMetrics: store data error")
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		storage.logger.Error().Err(err).Msg("StoreMetrics: commit transaction error")
		return err
	}

	return nil
}

func (storage *DatabaseStorage) Load(mtype, mname string) *metrics.Metric {
	var mID, mType string
	var mValue sql.NullFloat64
	var mDelta sql.NullInt64

	row := storage.db.QueryRow("SELECT id, type, value, delta FROM metrics WHERE type = $1 AND id = $2", mtype, mname)
	if err := row.Scan(&mID, &mType, &mValue, &mDelta); err != nil {
		return nil
	}
	return &metrics.Metric{
		MType: mType,
		ID:    mID,
		Value: parseValue(mValue),
		Delta: parseDelta(mDelta),
	}
}

func parseDelta(mDelta sql.NullInt64) *int64 {
	if mDelta.Valid {
		return &mDelta.Int64
	}
	return nil
}

func parseValue(mValue sql.NullFloat64) *float64 {
	if mValue.Valid {
		return &mValue.Float64
	}
	return nil
}

func (storage *DatabaseStorage) store(tx *sql.Tx, m metrics.Metric) error {
	storage.logger.Info().Any("metric", m).Send()

	var mID, mType string
	var mDelta sql.NullInt64

	raw := tx.QueryRow("SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2", m.ID, m.MType)
	err := raw.Scan(&mID, &mType, &mDelta)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		storage.logger.Error().Err(err).Msg("store: scan row error")
		return err
	}

	if m.MType == metrics.TypeCounter {
		switch {
		case mID != "":
			_, err = tx.Exec(
				"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
				mDelta.Int64+*m.Delta, m.ID, m.MType,
			)
		default:
			_, err = tx.Exec(
				"INSERT INTO metrics  (id, type, delta) VALUES ($1,$2,$3) ON CONFLICT (id, type) DO UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
				m.ID, m.MType, *m.Delta,
			)
		}
		if err != nil {
			storage.logger.Error().Err(err).Msg("store: error to store counter")
			return err
		}
	}

	if m.MType == metrics.TypeGauge {
		switch {
		case mID != "":
			_, err = tx.Exec(
				"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
				m.Value, m.ID, m.MType,
			)
		default:
			_, err = tx.Exec(
				"INSERT INTO metrics (id, type, value) VALUES ($1,$2,$3)",
				m.ID, m.MType, *m.Value,
			)
		}
		if err != nil {
			storage.logger.Error().Err(err).Msg("store: error to store gauge")
			return err
		}
	}

	return nil
}

func (storage *DatabaseStorage) StoreMetric(m metrics.Metric) error {
	tx, err := storage.db.Begin()
	if err != nil {
		storage.logger.Error().Err(err).Msg("Store: begin transaction error")
		return err
	}
	if err := storage.store(tx, m); err != nil {
		storage.logger.Error().Err(err).Msg("Store: store data error")
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		storage.logger.Error().Err(err).Msg("Store: commit transaction error")
		return err
	}
	return nil
}

func (storage *DatabaseStorage) RestoreFromFile() error {
	return errNotSupported
}

func (storage *DatabaseStorage) WriteToFile() error {
	return errNotSupported
}
