package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Message struct {
	ID                   int64     `json:"id" db:"id"`
	MessageDate          time.Time `json:"message_date" db:"message_date"`
	Message              string    `json:"message" db:"message"`
	ContactID            int64     `json:"contact_id" db:"contact_id"`
	Attachment           string    `json:"attachment" db:"attachment"`
	ChatID               int64     `json:"chat_id" db:"chat_id"`
	EnableSemanticSearch bool      `json:"enable_semantic_search" db:"enable_semantic_search"`
	Version              int32     `json:"version" db:"version"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	ModifiedAt           time.Time `json:"modified_at" db:"modified_at"`
}

type MessageModel struct {
	DB *sql.DB
}

func (m MessageModel) Insert(message *Message) error {
	query := `
		INSERT INTO messages (
			message_date,
			message,
			contact_id,
			attachment,
			chat_id
			, enable_semantic_search
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
			, $6
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		message.MessageDate,
		message.Message,
		message.ContactID,
		message.Attachment,
		message.ChatID,
		message.EnableSemanticSearch,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&message.ID, &message.Version, &message.CreatedAt, &message.ModifiedAt)

	if err != nil {
		return messageCustomError(err)
	}

	return nil

}

func (m MessageModel) Get(id int64) (*Message, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		message_date,
		message,
		contact_id,
		attachment,
		chat_id,
		enable_semantic_search,
		version, created_at, modified_at
		FROM messages
		WHERE id = $1`

	var message Message

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.MessageDate,
		&message.Message,
		&message.ContactID,
		&message.Attachment,
		&message.ChatID,
		&message.EnableSemanticSearch,
		&message.Version,
		&message.CreatedAt,
		&message.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &message, nil
}

func (m MessageModel) Update(message *Message) error {
	query := `
		UPDATE messages
		SET
		message_date = $1,
		message = $2,
		contact_id = $3,
		attachment = $4,
		chat_id = $5,
		enable_semantic_search = $6,
		version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

	args := []any{
		message.MessageDate,
		message.Message,
		message.ContactID,
		message.Attachment,
		message.ChatID,
		message.EnableSemanticSearch,
		message.ID,
		message.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&message.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return messageCustomError(err)
		}
	}

	return nil
}

func (m MessageModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM messages
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MessageModel) GetAll(filters Filters) ([]*Message, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		message_date,
		message,
		contact_id,
		attachment,
		chat_id,
		enable_semantic_search,
		version, created_at, modified_at
		FROM messages
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
			&message.MessageDate,
			&message.Message,
			&message.ContactID,
			&message.Attachment,
			&message.ChatID,
			&message.EnableSemanticSearch,
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

/*not_in_semantic_start*/
func (m MessageModel) GetAllNotInSemantic(filters Filters, contentField string) ([]*Message, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		message,
		version, created_at, modified_at
		FROM messages
		WHERE
		enable_semantic_search = true AND
		%s != '' AND
		id NOT IN (SELECT message_id FROM phrases WHERE content_field = '%s')
		ORDER BY %s %s, id ASC
		LIMIT $1 OFFSET $2`, contentField, contentField, filters.sortColumn(), filters.sortDirection())

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

/*not_in_semantic_end*/
