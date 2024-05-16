package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Message struct {
	ID                   int64     `json:"id,omitempty" db:"id"`
	MessageDate          time.Time `json:"message_date,omitempty" db:"message_date"`
	Message              string    `json:"message,omitempty" db:"message"`
	PhoneNumber          string    `json:"phone_number,omitempty" db:"phone_number"`
	Attachment           string    `json:"attachment,omitempty" db:"attachment"`
	EnableSemanticSearch bool      `json:"enable_semantic_search,omitempty" db:"enable_semantic_search"`
	ChatID               int64     `json:"chat_id,omitempty" db:"chat_id"`
	Version              int32     `json:"version,omitempty" db:"version"`
	CreatedAt            time.Time `json:"-" db:"created_at"`
	ModifiedAt           time.Time `json:"-" db:"modified_at"`
}

type MessageModel struct {
	DB *sql.DB
}

func (m MessageModel) Insert(message *Message) error {
	query := `
		INSERT INTO messages (
			message_date,
			message,
			phone_number,
			attachment,
			enable_semantic_search,
			chat_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		message.MessageDate,
		message.Message,
		message.PhoneNumber,
		message.Attachment,
		message.EnableSemanticSearch,
		message.ChatID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&message.ID, &message.Version, &message.CreatedAt, &message.ModifiedAt)
}

func (m MessageModel) Get(id int64) (*Message, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		message_date,
		message,
		phone_number,
		attachment,
		enable_semantic_search,
		chat_id,
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
		&message.PhoneNumber,
		&message.Attachment,
		&message.EnableSemanticSearch,
		&message.ChatID,
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
		phone_number = $3,
		attachment = $4,
		enable_semantic_search = $5,
		chat_id = $6,
		version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

	args := []any{
		message.MessageDate,
		message.Message,
		message.PhoneNumber,
		message.Attachment,
		message.EnableSemanticSearch,
		message.ChatID,
		message.ID,
		message.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&message.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
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

func (m MessageModel) GetAll(enable_semantic_search bool, filters Filters) ([]*Message, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		message_date,
		message,
		phone_number,
		attachment,
		enable_semantic_search,
		chat_id,
		version, created_at, modified_at
		FROM messages
		WHERE
		(enable_semantic_search = $1 OR $1 = false)
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		enable_semantic_search,
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
			&message.PhoneNumber,
			&message.Attachment,
			&message.EnableSemanticSearch,
			&message.ChatID,
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
