package model

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"time"
)

func SetupDB(dsn string) error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS patients(
			id TEXT PRIMARY KEY, 
			hashed_id TEXT
		)`,
	}
	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// FileStruct represents an in-memory file
type FileStruct struct {
	Name    string
	Content []byte
}

// ExportPatientsToCSV exports the patients table to an in-memory CSV file
func ExportPatientsToCSV(db *sql.DB) (*FileStruct, error) {
	rows, err := db.Query("SELECT id, hashed_id FROM patients")
	if err != nil {
		return nil, fmt.Errorf("database query failed: %v", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write([]string{"id", "hashed_id"}); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write rows
	for rows.Next() {
		var id, hashedID string
		if err := rows.Scan(&id, &hashedID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		if err := writer.Write([]string{id, hashedID}); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %v", err)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing CSV writer: %v", err)
	}

	// Generate a unique filename
	filename := fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05"))

	return &FileStruct{
		Name:    filename,
		Content: buf.Bytes(),
	}, nil
}

// Helper function to write the FileStruct to an io.Writer
func (f *FileStruct) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.Content)
	return int64(n), err
}
