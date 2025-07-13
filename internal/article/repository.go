package article

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type Repository interface {
	CreateArticle(ctx context.Context, article *Article) (*Article, error)
	GetArticles(ctx context.Context, filter *ArticleFilter) ([]*Article, error)
	GetArticlesByID(ctx context.Context, filter *ArticleFilter, ids []string) ([]*Article, error) // For fetching full articles from ES IDs
}

type postgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository.
func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

// CreateArticle inserts a new article into the database.
func (r *postgresRepository) CreateArticle(ctx context.Context, article *Article) (*Article, error) {
	query := `INSERT INTO articles (title, body, author_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	err := r.db.QueryRow(query, article.Title, article.Body, article.AuthorID, article.CreatedAt).Scan(&article.ID, &article.CreatedAt)
	if err != nil {
		return nil, err
	}

	return article, nil
}

// GetArticles retrieves a list of articles from the database based on filters.
// This method is used when no full-text search query is provided.
func (r *postgresRepository) GetArticles(ctx context.Context, filter *ArticleFilter) ([]*Article, error) {
	articles := []*Article{}
	var err error

	// Base query
	query := "SELECT a.id, a.title, a.body, authors.id, authors.name, a.created_at FROM articles a "
	query += "JOIN authors ON a.author_id = authors.id"
	args := []interface{}{}
	argCount := 1

	// Add author filter if present
	if filter.Author != "" {
		query += fmt.Sprintf(" WHERE authors.name = $%d", argCount)
		args = append(args, filter.Author)
		argCount++
	}

	// Order by latest first
	query += " ORDER BY created_at DESC"

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Title, &article.Body, &article.Author.ID, &article.Author.Name, &article.CreatedAt); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return articles, nil
}

// GetArticlesByID retrieves articles by their IDs. Used after an Elasticsearch search.
func (r *postgresRepository) GetArticlesByID(ctx context.Context, filter *ArticleFilter, ids []string) ([]*Article, error) {
	if len(ids) == 0 {
		return []*Article{}, nil
	}

	articles := []*Article{}
	var args []interface{}
	args = append(args, pq.Array(ids))

	query := `SELECT a.id, a.title, a.body, a.created_at, authors.id, authors.name FROM articles a `
	query += `JOIN authors ON a.author_id = authors.id `
	query += `WHERE a.id = ANY($1) `
	if filter != nil && filter.Author != "" {
		query += `AND authors.name = $2 `
		args = append(args, filter.Author)
	}
	query += `ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var article Article
		if err := rows.Scan(&article.ID, &article.Title, &article.Body, &article.CreatedAt, &article.Author.ID, &article.Author.Name); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return articles, nil
}
