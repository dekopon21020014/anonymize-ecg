package model

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"
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

// ExportPatientsToCSV exports the patients table to a CSV file
func ExportPatientsToCSV(db *sql.DB, filename string) error {
	rows, err := db.Query("SELECT id, hashed_id FROM patients")
	if err != nil {
		return err
	}
	defer rows.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"id", "hashed_id"}); err != nil {
		return err
	}

	// Write rows
	for rows.Next() {
		var id, hashedID string
		if err := rows.Scan(&id, &hashedID); err != nil {
			return err
		}
		if err := writer.Write([]string{id, hashedID}); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Println("CSV export successful")
	return nil
}
