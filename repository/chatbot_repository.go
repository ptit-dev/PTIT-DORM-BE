package repository

import (
	"context"
	"database/sql"
)

type ChatbotRepository struct {
	db *sql.DB
}

func NewChatbotRepository(db *sql.DB) *ChatbotRepository {
	return &ChatbotRepository{db: db}
}

// Get all documents from chatbot.documents
func (r *ChatbotRepository) GetDocuments(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM chatbot.documents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRowsToMap(rows)
}

// Get all promptings from chatbot.prompting
func (r *ChatbotRepository) GetPromptings(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM chatbot.prompting`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRowsToMap(rows)
}

// scanRowsToMap converts sql.Rows to []map[string]interface{} generically
func scanRowsToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			v := values[i]
			if b, ok := v.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = v
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
