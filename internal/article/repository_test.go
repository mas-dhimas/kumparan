package article_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"kumparan-test/internal/article"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func setupRepoWithMock(t *testing.T) (article.Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	return article.NewPostgresRepository(db), mock, func() { db.Close() }
}

func TestCreateArticle_Success(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	art := &article.Article{
		Title:     "Test Title",
		Body:      "Test Body",
		AuthorID:  "author-123",
		CreatedAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO articles`).
		WithArgs(art.Title, art.Body, art.AuthorID, art.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("article-456", art.CreatedAt))

	result, err := repo.CreateArticle(context.Background(), art)
	assert.NoError(t, err)
	assert.Equal(t, "article-456", result.ID)
	assert.Equal(t, art.CreatedAt, result.CreatedAt)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateArticle_Fails(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	art := &article.Article{
		Title:     "Bad Title",
		Body:      "Bad Body",
		AuthorID:  "bad-author",
		CreatedAt: time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO articles`).
		WithArgs(art.Title, art.Body, art.AuthorID, art.CreatedAt).
		WillReturnError(assert.AnError)

	_, err := repo.CreateArticle(context.Background(), art)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticles_Success(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	filter := &article.ArticleFilter{Page: 1, Limit: 2, Author: "Bara"}

	rows := sqlmock.NewRows([]string{
		"id", "title", "body", "id", "name", "created_at",
	}).AddRow("a1", "T1", "B1", "auth1", "Bara", time.Now()).
		AddRow("a2", "T2", "B2", "auth2", "Bara", time.Now())

	mock.ExpectQuery(`SELECT a\.id, a\.title, a\.body, authors\.id, authors\.name, a\.created_at`).
		WithArgs("Bara", 2, 0).
		WillReturnRows(rows)

	results, err := repo.GetArticles(context.Background(), filter)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticles_ScanError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	query := `SELECT a.id, a.title, a.body, authors.id, authors.name, a.created_at FROM articles a JOIN authors ON a.author_id = authors.id ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`

	mock.ExpectQuery(query).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
			AddRow("id-1", "Test Title"))

	filter := &article.ArticleFilter{Page: 1, Limit: 10}
	articles, err := repo.GetArticles(context.Background(), filter)

	assert.Error(t, err)
	assert.Nil(t, articles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticles_RowsError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	query := `SELECT a.id, a.title, a.body, authors.id, authors.name, a.created_at FROM articles a JOIN authors ON a.author_id = authors.id ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`

	rows := sqlmock.NewRows([]string{"id", "title", "body", "author_id", "author_name", "created_at"}).
		AddRow("id-1", "Title", "Body", "auth-1", "Bagunda", time.Now())

	mock.ExpectQuery(query).
		WithArgs(10, 0).
		WillReturnRows(rows.RowError(0, nil).CloseError(errors.New("rows iteration error")))

	filter := &article.ArticleFilter{Page: 1, Limit: 10}
	articles, err := repo.GetArticles(context.Background(), filter)

	assert.ErrorContains(t, err, "rows iteration error")
	assert.Nil(t, articles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticles_DBError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	filter := &article.ArticleFilter{Page: 1, Limit: 10, Author: "Biri"}

	mock.ExpectQuery(`SELECT a\.id, a\.title, a\.body, authors\.id, authors\.name, a\.created_at`).
		WithArgs("Biri", 10, 0).
		WillReturnError(assert.AnError)

	_, err := repo.GetArticles(context.Background(), filter)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticlesByID_Success(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	filter := &article.ArticleFilter{Author: "Bara"}
	ids := []string{"id-1", "id-2"}

	rows := sqlmock.NewRows([]string{
		"id", "title", "body", "created_at", "id", "name",
	}).AddRow("id-1", "T1", "B1", time.Now(), "auth1", "Bara").
		AddRow("id-2", "T2", "B2", time.Now(), "auth2", "Bara")

	mock.ExpectQuery(`SELECT a\.id, a\.title, a\.body, a\.created_at, authors\.id, authors\.name`).
		WithArgs(sqlmock.AnyArg(), "Bara").
		WillReturnRows(rows)

	result, err := repo.GetArticlesByID(context.Background(), filter, ids)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticlesByID_EmptyInput(t *testing.T) {
	repo, _, cleanup := setupRepoWithMock(t)
	defer cleanup()

	filter := &article.ArticleFilter{}
	ids := []string{}

	result, err := repo.GetArticlesByID(context.Background(), filter, ids)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestGetArticlesByID_ScanError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	ids := []string{"id1", "id2"}
	filter := &article.ArticleFilter{Page: 1, Limit: 10}

	mock.ExpectQuery(`SELECT a.id, a.title, a.body, a.created_at, authors.id, authors.name FROM articles a .*WHERE a.id = ANY\(\$1\).*`).
		WithArgs(pq.Array(ids)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
			AddRow("id1", "Some Title"))

	articles, err := repo.GetArticlesByID(context.Background(), filter, ids)

	assert.Error(t, err)
	assert.Nil(t, articles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticlesByID_RowsError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	now := time.Now()
	ids := []string{"id1"}
	filter := &article.ArticleFilter{}

	rows := sqlmock.NewRows([]string{"id", "title", "body", "created_at", "author_id", "author_name"}).
		AddRow("id1", "Title", "Body", now, "auth-1", "Author").
		RowError(0, nil)
	rows.CloseError(errors.New("rows iteration error"))

	mock.ExpectQuery(`SELECT a.id, a.title, a.body, a.created_at, authors.id, authors.name FROM articles a .*WHERE a.id = ANY\(\$1\).*`).
		WithArgs(pq.Array(ids)).
		WillReturnRows(rows)

	articles, err := repo.GetArticlesByID(context.Background(), filter, ids)

	assert.ErrorContains(t, err, "rows iteration error")
	assert.Nil(t, articles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetArticlesByID_DBError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	filter := &article.ArticleFilter{}
	ids := []string{"id-1"}

	mock.ExpectQuery(`SELECT a\.id, a\.title, a\.body, a\.created_at, authors\.id, authors\.name`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(assert.AnError)

	_, err := repo.GetArticlesByID(context.Background(), filter, ids)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
