package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Contact struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Version     int32     `json:"version" db:"version"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	ModifiedAt  time.Time `json:"modified_at" db:"modified_at"`
}

type ContactModel struct {
	DB *sql.DB
}

func (m ContactModel) Insert(contact *Contact) error {
	query := `
		INSERT INTO contacts (
			name,
			phone_number
		)
		VALUES (
			$1,
			$2
		)
		RETURNING id, version, created_at, modified_at`

	args := []any{
		contact.Name,
		contact.PhoneNumber,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&contact.ID, &contact.Version, &contact.CreatedAt, &contact.ModifiedAt)

	if err != nil {
		return contactCustomError(err)
	}

	return nil

}

func (m ContactModel) Get(id int64) (*Contact, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id,
		name,
		phone_number,
		version, created_at, modified_at
		FROM contacts
		WHERE id = $1`

	var contact Contact

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // releases resources if slowOperation completes before timeout elapses, prevents memory leak

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&contact.ID,
		&contact.Name,
		&contact.PhoneNumber,
		&contact.Version,
		&contact.CreatedAt,
		&contact.ModifiedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &contact, nil
}

func (m ContactModel) Update(contact *Contact) error {
	query := `
		UPDATE contacts
		SET
		name = $1,
		phone_number = $2,
		version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version`

	args := []any{
		contact.Name,
		contact.PhoneNumber,
		contact.ID,
		contact.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&contact.Version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEditConflict
		} else {
			return contactCustomError(err)
		}
	}

	return nil
}

func (m ContactModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM contacts
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

func (m ContactModel) GetAll(name string, phone_number string, filters Filters) ([]*Contact, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id,
		name,
		phone_number,
		version, created_at, modified_at
		FROM contacts
		WHERE
		(to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '') AND 
		(to_tsvector('simple', phone_number) @@ plainto_tsquery('simple', $2) OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		name,
		phone_number,
		filters.limit(),
		filters.offset(),
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	contacts := []*Contact{}

	for rows.Next() {
		var contact Contact

		err := rows.Scan(
			&totalRecords,
			&contact.ID,
			&contact.Name,
			&contact.PhoneNumber,
			&contact.Version,
			&contact.CreatedAt,
			&contact.ModifiedAt,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		contacts = append(contacts, &contact)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return contacts, metadata, nil
}
