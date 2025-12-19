package repository

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
)

// UserRepositoryImpl implements UserRepository
type BackUpRepository struct {
	db *sql.DB
}

func NewBackUpRepository(db *sql.DB) *BackUpRepository {
	return &BackUpRepository{db: db}
}

func (r *BackUpRepository) BackupAllTablesToCSVZip() ([]byte, error) {
	db := r.db
	rows, err := db.Query(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tables = append(tables, t)
		}
	}
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for _, table := range tables {
		dataRows, err := db.Query(fmt.Sprintf(`SELECT * FROM "%s"`, table))
		if err != nil {
			continue
		}
		cols, _ := dataRows.Columns()
		records := [][]string{cols}
		for dataRows.Next() {
			vals := make([]interface{}, len(cols))
			valPtrs := make([]interface{}, len(cols))
			for i := range vals {
				valPtrs[i] = &vals[i]
			}
			if err := dataRows.Scan(valPtrs...); err != nil {
				continue
			}
			rec := make([]string, len(cols))
			for i, v := range vals {
				if v == nil {
					rec[i] = ""
				} else {
					rec[i] = fmt.Sprintf("%v", v)
				}
			}
			records = append(records, rec)
		}
		dataRows.Close()
		if len(records) > 1 {
			f, _ := zipWriter.Create(table + ".csv")
			w := csv.NewWriter(f)
			w.WriteAll(records)
			w.Flush()
		}
	}
	zipWriter.Close()
	return buf.Bytes(), nil
}
