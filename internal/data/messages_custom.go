package data

import (
	"context"
	"fmt"
	"time"
)

func (m MessageModel) GetAllNotInSemantic(filters Filters) ([]*Message, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		message,
		version, created_at, modified_at
		FROM messages
		WHERE
		enable_semantic_search = true AND
		id NOT IN (SELECT message_id FROM phrases)
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	messages := []*Message{}

	for rows.Next() {
		var message Message

		err := rows.Scan(
			&totalRecords,
			&message.ID,
			&message.Message,
			&message.Version,
			&message.CreatedAt,
			&message.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return messages, metadata, nil
}

//custom_code
