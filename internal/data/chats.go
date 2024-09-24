package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Chat struct {
	ID         int64     `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Version    int32     `json:"version" db:"version"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ModifiedAt time.Time `json:"modified_at" db:"modified_at"`
}

type ChatModel struct {
	DB *sql.DB
}

func (m ChatModel) Insert(chat *Chat) error {
	query := `
		INSERT INTO chats (
			name
		)
		VALUES (
			$1
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		chat.Name,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&chat.ID, &chat.Version, &chat.CreatedAt, &chat.ModifiedAt)

	if err != nil {
		return chatCustomError(err)
	}

	return nil

}

func (m ChatModel) Get(id int64) (*Chat, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		name,
		version, created_at, modified_at
		FROM chats
		WHERE id = $1`

	var chat Chat

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&chat.ID,
		&chat.Name,
		&chat.Version,
		&chat.CreatedAt,
		&chat.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &chat, nil
}

func (m ChatModel) Update(chat *Chat) error {
	query := `
		UPDATE chats
		SET
		name = $1,
		version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`

	args := []any{
		chat.Name,
		chat.ID,
		chat.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&chat.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return chatCustomError(err)
		}
	}

	return nil
}

func (m ChatModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM chats
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

func (m ChatModel) GetAll(name string, filters Filters) ([]*Chat, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		version, created_at, modified_at
		FROM chats
		WHERE
		(to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	chats := []*Chat{}

	for rows.Next() {
		var chat Chat

		err := rows.Scan(
			&totalRecords,
			&chat.ID,
			&chat.Name,
			&chat.Version,
			&chat.CreatedAt,
			&chat.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		chats = append(chats, &chat)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return chats, metadata, nil
}
