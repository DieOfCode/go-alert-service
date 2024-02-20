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

func (storage *DatabaseStorage) StoreMetrics(metrics []metrics.Metric) bool {
	var stored bool
	for _, metric := range metrics {
		stored = storage.Store(metric)
		if !stored {
			return false
		}
	}
	return true
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

func (storage *DatabaseStorage) Store(m metrics.Metric) bool {
	var mID, mType string
	var mDelta sql.NullInt64

	raw := storage.db.QueryRow("SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2", m.ID, m.MType)
	if err := raw.Scan(&mID, &mType, &mDelta); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false
	}

	if m.MType == metrics.TypeCounter {
		result, err := storage.db.Exec(
			"INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3) ON CONFLICT (id, type) DO UPDATE SET value = EXCLUDED.value",
			m.ID, m.MType, mDelta.Int64+*m.Delta,
		)
		if err != nil {
			return false
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return false
		}
		if affected != 1 {
			return false
		}
		return true
	}

	if mID != "" && m.MType == metrics.TypeGauge {
		result, err := storage.db.Exec(
			"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
			m.Value, m.ID, m.MType,
		)
		if err != nil {
			return false
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return false
		}
		if affected != 1 {
			return false
		}
		return true
	}

	var result sql.Result
	var err error
	if m.MType == metrics.TypeGauge {
		result, err = storage.db.Exec(
			"INSERT INTO metrics (id, type, value) VALUES ($1,$2,$3)",
			m.ID, m.MType, *m.Value,
		)
	}
	// else {
	// 	result, err = storage.db.Exec(
	// 		"INSERT INTO metrics (id, type, delta) VALUES ($1,$2,$3)",
	// 		m.ID, m.MType, *m.Delta,
	// 	)
	// }

	if err != nil {
		return false
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false
	}
	if affected != 1 {
		return false
	}

	return true
}

func (storage *DatabaseStorage) RestoreFromFile() error {
	return errNotSupported
}

func (storage *DatabaseStorage) WriteToFile() error {
	return errNotSupported
}
