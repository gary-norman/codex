package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gary-norman/forum/internal/models"
)

// LoggingModel provides database operations for all logging tables
type LoggingModel struct {
	DB *sql.DB
}

// InsertRequestLog stores an HTTP request log entry
func (m *LoggingModel) InsertRequestLog(ctx context.Context, log models.RequestLog) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for InsertRequestLog: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `INSERT INTO RequestLogs
		(Timestamp, Method, Path, StatusCode, Duration, UserID, IPAddress, UserAgent, Referer, BytesSent)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(
		ctx,
		query,
		log.Timestamp,
		log.Method,
		log.Path,
		log.StatusCode,
		log.Duration,
		log.UserID,
		log.IPAddress,
		log.UserAgent,
		log.Referer,
		log.BytesSent,
	)

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction for InsertRequestLog: %w", commitErr)
	}
	return err
}

// InsertErrorLog stores an application error log entry
func (m *LoggingModel) InsertErrorLog(ctx context.Context, log models.ErrorLog) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for InsertErrorLog: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `INSERT INTO ErrorLogs
		(Timestamp, Level, Message, StackTrace, RequestPath, UserID, Context)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(
		ctx,
		query,
		log.Timestamp,
		log.Level,
		log.Message,
		log.StackTrace,
		log.RequestPath,
		log.UserID,
		log.Context,
	)

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction for InsertErrorLog: %w", commitErr)
	}

	return err
}

// InsertSystemMetric stores a system performance/health metric
func (m *LoggingModel) InsertSystemMetric(ctx context.Context, metric models.SystemMetric) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for InsertSystemMetric: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `INSERT INTO SystemMetrics
		(Timestamp, MetricType, MetricName, MetricValue, Unit, Details)
		VALUES (?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(
		ctx,
		query,
		metric.Timestamp,
		metric.MetricType,
		metric.MetricName,
		metric.MetricValue,
		metric.Unit,
		metric.Details,
	)

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction for InsertSystemMetric: %w", commitErr)
	}

	return err
}

// GetRequestLogsSince retrieves request logs after a given timestamp
func (m *LoggingModel) GetRequestLogsSince(ctx context.Context, since string, limit int) ([]models.RequestLog, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetRequestLogsSince: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `SELECT ID, Timestamp, Method, Path, StatusCode, Duration, UserID, IPAddress, UserAgent, Referer, BytesSent
		FROM RequestLogs
		WHERE Timestamp >= ?
		ORDER BY Timestamp DESC
		LIMIT ?`

	rows, err := m.DB.Query(query, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.RequestLog
	for rows.Next() {
		var log models.RequestLog
		err = rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Method,
			&log.Path,
			&log.StatusCode,
			&log.Duration,
			&log.UserID,
			&log.IPAddress,
			&log.UserAgent,
			&log.Referer,
			&log.BytesSent,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction for InsertSystemMetric: %w", commitErr)
	}

	return logs, rows.Err()
}

// GetErrorLogsSince retrieves error logs after a given timestamp
func (m *LoggingModel) GetErrorLogsSince(ctx context.Context, since string, limit int) ([]models.ErrorLog, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetErrorLogsSince: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `SELECT ID, Timestamp, Level, Message, StackTrace, RequestPath, UserID, Context
		FROM ErrorLogs
		WHERE Timestamp >= ?
		ORDER BY Timestamp DESC
		LIMIT ?`

	rows, err := m.DB.Query(query, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ErrorLog
	for rows.Next() {
		var log models.ErrorLog
		err = rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.Level,
			&log.Message,
			&log.StackTrace,
			&log.RequestPath,
			&log.UserID,
			&log.Context,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetErrorLogsSince: %w", commitErr)
	}

	return logs, rows.Err()
}

// GetSystemMetricsSince retrieves system metrics after a given timestamp
func (m *LoggingModel) GetSystemMetricsSince(ctx context.Context, since string, limit int) ([]models.SystemMetric, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetSystemMetricsSince: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `SELECT ID, Timestamp, MetricType, MetricName, MetricValue, Unit, Details
		FROM SystemMetrics
		WHERE Timestamp >= ?
		ORDER BY Timestamp DESC
		LIMIT ?`

	rows, err := m.DB.Query(query, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.SystemMetric
	for rows.Next() {
		var metric models.SystemMetric
		err = rows.Scan(
			&metric.ID,
			&metric.Timestamp,
			&metric.MetricType,
			&metric.MetricName,
			&metric.MetricValue,
			&metric.Unit,
			&metric.Details,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetSystemMetricsSince: %w", commitErr)
	}

	return metrics, rows.Err()
}

// GetRequestStats retrieves aggregated request statistics
type RequestStats struct {
	TotalRequests   int64
	AvgDuration     float64
	ErrorRate       float64 // Percentage of 4xx/5xx responses
	UniqueUsers     int64
	RequestsPerPath map[string]int64
}

func (m *LoggingModel) GetRequestStats(ctx context.Context, since string) (*RequestStats, error) {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for GetRequestStats: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	stats := &RequestStats{
		RequestsPerPath: make(map[string]int64),
	}

	// Total requests and average duration
	err = m.DB.QueryRow(`
		SELECT COUNT(*), AVG(Duration)
		FROM RequestLogs
		WHERE Timestamp >= ?`, since).Scan(&stats.TotalRequests, &stats.AvgDuration)
	if err != nil {
		return nil, err
	}

	// Error rate
	var errorCount int64
	err = m.DB.QueryRow(`
		SELECT COUNT(*)
		FROM RequestLogs
		WHERE Timestamp >= ? AND StatusCode >= 400`, since).Scan(&errorCount)
	if err != nil {
		return nil, err
	}
	if stats.TotalRequests > 0 {
		stats.ErrorRate = float64(errorCount) / float64(stats.TotalRequests) * 100
	}

	// Unique users
	err = m.DB.QueryRow(`
		SELECT COUNT(DISTINCT UserID)
		FROM RequestLogs
		WHERE Timestamp >= ? AND UserID IS NOT NULL`, since).Scan(&stats.UniqueUsers)
	if err != nil {
		return nil, err
	}

	// Requests per path
	rows, err := m.DB.Query(`
		SELECT Path, COUNT(*)
		FROM RequestLogs
		WHERE Timestamp >= ?
		GROUP BY Path
		ORDER BY COUNT(*) DESC
		LIMIT 20`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		var count int64
		if err := rows.Scan(&path, &count); err != nil {
			return nil, err
		}
		stats.RequestsPerPath[path] = count
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, fmt.Errorf("failed to commit transaction for GetRequestStats: %w", commitErr)
	}

	return stats, rows.Err()
}

// CleanupOldLogs deletes logs older than the specified number of days
func (m *LoggingModel) CleanupOldLogs(ctx context.Context, daysToKeep int) error {
	// Begin the transaction
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for CleanupOldLogs: %w", err)
	}

	// Ensure rollback on failure
	defer func() {
		if p := recover(); p != nil {
			models.LogWarnWithContext(ctx, "Panic occurred, rolling back transaction: %v", p)
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	cutoff := `datetime('now', '-` + string(rune(daysToKeep+'0')) + ` days')`

	// Clean request logs
	if _, err := tx.ExecContext(ctx, `DELETE FROM RequestLogs WHERE Timestamp < `+cutoff); err != nil {
		return err
	}

	// Clean error logs
	if _, err := tx.ExecContext(ctx, `DELETE FROM ErrorLogs WHERE Timestamp < `+cutoff); err != nil {
		return err
	}

	// Clean system metrics
	if _, err := tx.ExecContext(ctx, `DELETE FROM SystemMetrics WHERE Timestamp < `+cutoff); err != nil {
		return err
	}

	// Commit the transaction
	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction for CleanupOldLogs: %w", commitErr)
	}

	return nil
}
