package db

import (
	"database/sql"
	"time"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
	_ "modernc.org/sqlite"
)

func InitDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp TEXT NOT NULL,
			symbol TEXT NOT NULL,
			log_type TEXT NOT NULL,
			message TEXT NOT NULL
		)
	`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC)`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_logs_symbol ON logs(symbol)`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func SaveLog(db *sql.DB, entry *models.LogEntry) error {
	_, err := db.Exec(
		`INSERT INTO logs (timestamp, symbol, log_type, message) VALUES (?, ?, ?, ?)`,
		entry.Timestamp.Format(time.RFC3339),
		entry.Symbol,
		entry.LogType.String(),
		entry.Message,
	)
	return err
}

func LoadLogs(db *sql.DB, limit int) ([]models.LogEntry, error) {
	rows, err := db.Query(`SELECT timestamp, symbol, log_type, message FROM logs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.LogEntry, 0)
	for rows.Next() {
		var ts string
		var symbol string
		var logType string
		var message string

		if err := rows.Scan(&ts, &symbol, &logType, &message); err != nil {
			return nil, err
		}

		parsedTS, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, err
		}

		logs = append(logs, models.LogEntry{
			Timestamp: parsedTS,
			Symbol:    symbol,
			Message:   message,
			LogType:   models.LogTypeFromString(logType),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, nil
}

func AddSessionStartLog(db *sql.DB) error {
	return SaveLog(db, &models.LogEntry{
		Timestamp: time.Now(),
		Symbol:    "SYSTEM",
		Message:   "Session started",
		LogType:   models.SessionStart,
	})
}
