package author

import (
	"context"
	"database/sql"
)

type Repository interface {
	CreateAuthor(ctx context.Context, author *Author) (*Author, error)
	GetAuthorByName(ctx context.Context, name string) (*Author, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

// CreateAuthor inserts a new author into the database.
func (r *postgresRepository) CreateAuthor(ctx context.Context, author *Author) (*Author, error) {
	query := `INSERT INTO authors (name) VALUES ($1) RETURNING id`
	err := r.db.QueryRow(query, author.Name).Scan(&author.ID)
	if err != nil {
		return nil, err
	}
	return author, nil
}

// GetAuthorByName retrieves an author by their name.
func (r *postgresRepository) GetAuthorByName(ctx context.Context, name string) (*Author, error) {
	query := `SELECT id, name FROM authors WHERE name = $1`
	var author Author
	err := r.db.QueryRow(query, name).Scan(&author.ID, &author.Name)
	if err != nil {
		return nil, err
	}
	return &author, nil
}
