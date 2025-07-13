package author_test

import (
	"context"
	"database/sql"
	"errors"
	"kumparan-test/internal/author"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/assert"
)

func setupRepoWithMock(t *testing.T) (author.Repository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	return author.NewPostgresRepository(db), mock, func() { db.Close() }
}

func TestCreateAuthor_Success(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO authors \(name\) VALUES \(\$1\) RETURNING id`).
		WithArgs("Bara Biri").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("auth-1"))

	a := &author.Author{Name: "Bara Biri"}
	result, err := repo.CreateAuthor(context.Background(), a)

	assert.NoError(t, err)
	assert.Equal(t, "auth-1", result.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateAuthor_Error(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO authors \(name\) VALUES \(\$1\) RETURNING id`).
		WithArgs("Bara Biri").
		WillReturnError(errors.New("insert failed"))

	a := &author.Author{Name: "Bara Biri"}
	result, err := repo.CreateAuthor(context.Background(), a)

	assert.ErrorContains(t, err, "insert failed")
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAuthorByName_Success(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name FROM authors WHERE name = \$1`).
		WithArgs("Bara Biri").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("auth-1", "Bara Biri"))

	result, err := repo.GetAuthorByName(context.Background(), "Bara Biri")

	assert.NoError(t, err)
	assert.Equal(t, "auth-1", result.ID)
	assert.Equal(t, "Bara Biri", result.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAuthorByName_NotFound(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name FROM authors WHERE name = \$1`).
		WithArgs("Unknown").
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetAuthorByName(context.Background(), "Unknown")

	assert.ErrorIs(t, err, sql.ErrNoRows)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAuthorByName_ScanError(t *testing.T) {
	repo, mock, cleanup := setupRepoWithMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, name FROM authors WHERE name = \$1`).
		WithArgs("Bara Biri").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("auth-1")) // missing name

	result, err := repo.GetAuthorByName(context.Background(), "Bara Biri")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}
