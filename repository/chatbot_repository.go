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

// CreateDocument inserts a new record into chatbot.documents
func (r *ChatbotRepository) CreateDocument(ctx context.Context, id, description, content string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chatbot.documents (id, description, content, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())`,
		id, description, content,
	)
	return err
}

// UpdateDocument updates description and content of an existing document
func (r *ChatbotRepository) UpdateDocument(ctx context.Context, id, description, content string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE chatbot.documents
		   SET description = $2,
		       content = $3,
		       updated_at = NOW()
		 WHERE id = $1`,
		id, description, content,
	)
	return err
}

// DeleteDocument removes a document by id
func (r *ChatbotRepository) DeleteDocument(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chatbot.documents WHERE id = $1`, id)
	return err
}

// CreatePrompting inserts a new record into chatbot.prompting
func (r *ChatbotRepository) CreatePrompting(ctx context.Context, id, ptype, content string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO chatbot.prompting (id, type, content, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())`,
		id, ptype, content,
	)
	return err
}

// UpdatePrompting updates type and content of an existing prompting
func (r *ChatbotRepository) UpdatePrompting(ctx context.Context, id, ptype, content string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE chatbot.prompting
		   SET type = $2,
		       content = $3,
		       updated_at = NOW()
		 WHERE id = $1`,
		id, ptype, content,
	)
	return err
}

// DeletePrompting removes a prompting by id
func (r *ChatbotRepository) DeletePrompting(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM chatbot.prompting WHERE id = $1`, id)
	return err
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
