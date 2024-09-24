package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Phrase struct {
	ID           int64     `json:"id,omitempty" db:"id"`
	MessageID    int64     `json:"message_id,omitempty" db:"message_id"`
	Title        string    `json:"title,omitempty" db:"title"`
	Similarity   float64   `json:"similarity,omitempty" db:"similarity"`
	Content      string    `json:"content,omitempty" db:"content"`
	Tokens       int       `json:"tokens,omitempty" db:"tokens"`
	Sequence     int       `json:"sequence,omitempty" db:"sequence"`
	ContentField string    `json:"content_field,omitempty" db:"content_field"`
	Version      int32     `json:"version,omitempty" db:"version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ModifiedAt   time.Time `json:"modified_at" db:"modified_at"`
}

type PhraseModel struct {
	DB *sql.DB
}

func (m PhraseModel) Insert(phrase *Phrase) error {
	query := `
		INSERT INTO phrases (
			message_id
			, title, content, tokens, sequence, content_field
		)
		VALUES (
			$1
			, $2, $3, $4, $5, $6
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		phrase.MessageID,
		phrase.Title,
		phrase.Content,
		phrase.Tokens,
		phrase.Sequence,
		phrase.ContentField,
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
		message_id,
		title, content, tokens, sequence, content_field,
		version, created_at, modified_at
		FROM phrases
		WHERE id = $1`

	var phrase Phrase

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&phrase.ID,
		&phrase.MessageID,
		&phrase.Title,
		&phrase.Content,
		&phrase.Tokens,
		&phrase.Sequence,
		&phrase.ContentField,
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
		message_id = $1,
		title = $2,
		content = $3,
		tokens = $4,
		sequence = $5,
		content_field = $6,
		version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

	args := []any{
		phrase.MessageID,
		phrase.Title,
		phrase.Content,
		phrase.Tokens,
		phrase.Sequence,
		phrase.ContentField,
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

func (m PhraseModel) GetAll(message_id int64, filters Filters) ([]*Phrase, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		message_id,
		title, content, tokens, sequence, content_field,
		version, created_at, modified_at
		FROM phrases
		WHERE
		(message_id = $1 OR $1 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
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
			&phrase.MessageID,
			&phrase.Title,
			&phrase.Content,
			&phrase.Tokens,
			&phrase.Sequence,
			&phrase.ContentField,
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

// update_embedding_start
func (m PhraseModel) UpdateEmbedding(phrase *Phrase, embeddings pgvector.Vector, provider string) error {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE phrases
		SET
		%s = $1,
		version = version + 1
		WHERE id = $2 AND version = $3
		RETURNING version`, embeddings_field)

	args := []any{
		embeddings,
		phrase.ID,
		phrase.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&phrase.Version)
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

//update_embedding_end

// get_all_semantic_start
func (m PhraseModel) GetAllSemantic(embedding pgvector.Vector, similarity float64, provider string, content_fields []string, message_id int64, filters Filters) ([]*Phrase, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), phrases.id,
			phrases.message_id,
			phrases.title,
			phrases.content, 
			phrases.tokens, 
			phrases.sequence, 
			phrases.content_field,
			phrases.version, phrases.created_at, phrases.modified_at, 1 - (%s <=> $2) AS cosine_similarity
		FROM phrases
		WHERE
		(phrases.message_id = $1 OR $1 = 0) AND
		1 - (%s <=> $2) > $3 AND
		(phrases.content_field = ANY ($4) OR $4 = '{}')
		ORDER BY cosine_similarity DESC
		LIMIT $5 OFFSET $6`, embeddings_field, embeddings_field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		message_id,
		embedding,
		similarity,
		pq.Array(content_fields),
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
			&phrase.MessageID,
			&phrase.Title,
			&phrase.Content,
			&phrase.Tokens,
			&phrase.Sequence,
			&phrase.ContentField,
			&phrase.Version,
			&phrase.CreatedAt,
			&phrase.ModifiedAt,
			&phrase.Similarity,
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

//get_all_semantic_end

// get_all_without_embeddings_start
func (m PhraseModel) GetAllWithoutEmbeddings(limit int, provider string) ([]*Phrase, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		title,
		content,
		version
		FROM phrases
		WHERE %s IS NULL
		AND content != ''
		LIMIT $1
		`, embeddings_field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		limit,
	}

	filters := Filters{
		Page:     1,
		PageSize: limit,
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
			&phrase.Title,
			&phrase.Content,
			&phrase.Version,
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

//get_all_without_embeddings_end
