package model

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"time"
)

type Patient struct {
	Id        string
	HashedId  string
	Name      string
	Birthtime string
}

func SetupDB(dsn string) error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS patients(
			id TEXT PRIMARY KEY, 
			hashed_id TEXT NOT NULL,
			name TEXT,
			birthtime TEXT
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

// File represents an in-memory file
type File struct {
	Name    string
	Content []byte
}

// ExportPatientsToCSV exports the patients table to an in-memory CSV file
func ExportPatientsToCSV(db *sql.DB) (*File, error) {
	rows, err := db.Query("SELECT * FROM patients")
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	if err := writer.Write([]string{"id", "hashed_id", "name", "birthtime"}); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for rows.Next() {
		var id, hashedID, name, birthtime string
		if err := rows.Scan(&id, &hashedID, &name, &birthtime); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		if err := writer.Write([]string{id, hashedID, name, birthtime}); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("error flushing CSV writer: %w", err)
	}

	// Generate a unique filename
	filename := fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05"))

	return &File{
		Name:    filename,
		Content: buf.Bytes(),
	}, nil
}

// Helper function to write the File to an io.Writer
func (f *File) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(f.Content)
	return int64(n), err
}

// 指定されたテーブル全てのエントリを削除
func DeleteAllEntry(db *sql.DB, table string) error {
	// SQL文を作成
	query := fmt.Sprintf("DELETE FROM %s", table)

	// トランザクションを開始
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// SQL文を実行
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	// トランザクションをコミット
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func Put(db *sql.DB, patient Patient) error {
	fmt.Printf("patient = %+v\n", patient)
	// まず、指定されたIDが存在するかを確認
	var existing Patient
	query := `SELECT id, hashed_id, name, birthtime FROM patients WHERE id = ?`
	err := db.QueryRow(query, patient.Id).Scan(&existing.Id, &existing.HashedId, &existing.Name, &existing.Birthtime)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// IDが存在しない場合は、新しいレコードを挿入
			insertQuery := `INSERT INTO patients (id, hashed_id, name, birthtime) VALUES (?, ?, ?, ?)`
			_, err := db.Exec(insertQuery, patient.Id, patient.HashedId, patient.Name, patient.Birthtime)
			if err != nil {
				fmt.Println("maji")
				return fmt.Errorf("failed to insert new patient: %w", err)
			}
		} else {
			// その他のエラーが発生した場合
			return fmt.Errorf("failed to check existing patient: %w", err)
		}
	} else {
		// IDが存在する場合は、空のカラムのみを更新
		updateFields := []string{}
		args := []interface{}{}

		if existing.Name == "" && patient.Name != "" {
			updateFields = append(updateFields, "name = ?")
			args = append(args, patient.Name)
		}

		if existing.Birthtime == "" && patient.Birthtime != "" {
			updateFields = append(updateFields, "birthtime = ?")
			args = append(args, patient.Birthtime)
		}

		if len(updateFields) > 0 {
			updateQuery := `UPDATE patients SET ` + updateFields[0]

			// 複数のカラムが更新対象の場合、クエリを追加
			for i := 1; i < len(updateFields); i++ {
				updateQuery += fmt.Sprintf(", %s", updateFields[i])
			}

			updateQuery += ` WHERE id = ?`
			args = append(args, patient.Id)

			_, err := db.Exec(updateQuery, args...)
			if err != nil {
				return fmt.Errorf("failed to update patient: %v", err)
			}
		}
	}
	return nil
}
