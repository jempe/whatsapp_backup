package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Phrase struct {
	ID           int64     `json:"id,omitempty" db:"id"`
	Content      string    `json:"content,omitempty" db:"content"`
	Tokens       int       `json:"tokens,omitempty" db:"tokens"`
	Sequence     int       `json:"sequence,omitempty" db:"sequence"`
	ContentField string    `json:"content_field,omitempty" db:"content_field"`
	MessageID    int64     `json:"message_id,omitempty" db:"message_id"`
	Similarity   float64   `json:"similarity,omitempty" db:"similarity"`
	Version      int32     `json:"version,omitempty" db:"version"`
	CreatedAt    time.Time `json:"-" db:"created_at"`
	ModifiedAt   time.Time `json:"-" db:"modified_at"`
}

type PhraseModel struct {
	DB *sql.DB
}

func (m PhraseModel) Insert(phrase *Phrase) error {
	query := `
		INSERT INTO phrases (
			content,
			tokens,
			sequence,
			content_field,
			message_id
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		phrase.Content,
		phrase.Tokens,
		phrase.Sequence,
		phrase.ContentField,
		phrase.MessageID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&phrase.ID, &phrase.Version, &phrase.CreatedAt, &phrase.ModifiedAt)
}

func (m PhraseModel) Get(id int64) (*Phrase, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		content,
		tokens,
		sequence,
		content_field,
		message_id,
		version, created_at, modified_at
		FROM phrases
		WHERE id = $1`

	var phrase Phrase

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&phrase.ID,
		&phrase.Content,
		&phrase.Tokens,
		&phrase.Sequence,
		&phrase.ContentField,
		&phrase.MessageID,
		&phrase.Version,
		&phrase.CreatedAt,
		&phrase.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &phrase, nil
}

func (m PhraseModel) Update(phrase *Phrase) error {
	query := `
		UPDATE phrases
		SET
		content = $1,
		tokens = $2,
		sequence = $3,
		content_field = $4,
		message_id = $5,
		version = version + 1
		WHERE id = $6 AND version = $7
		RETURNING version`

	args := []any{
		phrase.Content,
		phrase.Tokens,
		phrase.Sequence,
		phrase.ContentField,
		phrase.MessageID,
		phrase.ID,
		phrase.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&phrase.Version)
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

func (m PhraseModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM phrases
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

func (m PhraseModel) GetAll(content_field string, message_id int64, filters Filters) ([]*Phrase, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		content,
		tokens,
		sequence,
		content_field,
		message_id,
		version, created_at, modified_at
		FROM phrases
		WHERE
		(to_tsvector('simple', content_field) @@ plainto_tsquery('simple', $1) OR $1 = '') AND 
		(message_id = $2 OR $2 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		content_field,
		message_id,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	phrases := []*Phrase{}

	for rows.Next() {
		var phrase Phrase

		err := rows.Scan(
			&totalRecords,
			&phrase.ID,
			&phrase.Content,
			&phrase.Tokens,
			&phrase.Sequence,
			&phrase.ContentField,
			&phrase.MessageID,
			&phrase.Version,
			&phrase.CreatedAt,
			&phrase.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		phrases = append(phrases, &phrase)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return phrases, metadata, nil
}
