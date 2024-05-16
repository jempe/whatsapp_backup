package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/pgvector/pgvector-go"
)

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

func (m PhraseModel) GetAllSemantic(embedding pgvector.Vector, similarity float64, provider string, content_field string, message_id int64, filters Filters) ([]*Phrase, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), phrases.id,
			phrases.content,
			phrases.tokens,
			phrases.sequence,
			phrases.content_field,
			phrases.message_id,
			phrases.version, phrases.created_at, phrases.modified_at, 1 - (%s <=> $3) AS cosine_similarity
		FROM phrases
		--ADD CUSTOM JOINS HERE
		WHERE
		(phrases.content_field = $1 OR $1 = '') AND
		(phrases.message_id = $2 OR $2 = 0) AND
		1 - (%s <=> $3) > $4
		ORDER BY cosine_similarity DESC
		LIMIT $5 OFFSET $6`, embeddings_field, embeddings_field)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		content_field,
		message_id,
		embedding,
		similarity,
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

func (m PhraseModel) GetAllWithoutEmbeddings(limit int, provider string) ([]*Phrase, Metadata, error) {
	embeddings_field, err := getEmbeddingsField(provider)
	if err != nil {
		return nil, Metadata{}, err
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
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

//custom_code
