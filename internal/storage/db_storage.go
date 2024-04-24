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

	for _, metric := range metrics {
		err := storage.Store(metric)
		if err != nil {
			return err
		}
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

func (storage *DatabaseStorage) Store(m metrics.Metric) error {
	var query string
	var args []interface{}

	if m.MType == metrics.TypeCounter {
		query = `
            INSERT INTO metrics (id, type, delta) VALUES ($1, $2, $3)
            ON CONFLICT (id, type) DO UPDATE
            SET delta = metrics.delta + EXCLUDED.delta
            WHERE metrics.type = 'counter'
        `
		args = append(args, m.ID, m.MType, *m.Delta)
	} else if m.MType == metrics.TypeGauge {
		query = `
            INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3)
            ON CONFLICT (id, type) DO UPDATE
            SET value = EXCLUDED.value
            WHERE metrics.type = 'gauge'
        `
		args = append(args, m.ID, m.MType, *m.Value)
	}

	_, err := storage.db.Exec(query, args...)

	return err
}

func (storage *DatabaseStorage) RestoreFromFile() error {
	return errNotSupported
}

func (storage *DatabaseStorage) WriteToFile() error {
	return errNotSupported
}
