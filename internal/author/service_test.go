package author_test

import (
	"context"
	"database/sql"
	"errors"
	"kumparan-test/internal/author"
	"kumparan-test/internal/author/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrCreateAuthor_AuthorExists(t *testing.T) {
	mockRepo := new(mocks.MockAuthorRepo)
	svc := author.NewAuthorService(mockRepo)

	existing := &author.Author{ID: "auth-1", Name: "Bara"}

	mockRepo.On("GetAuthorByName", mock.Anything, "Bara").Return(existing, nil)

	result, err := svc.GetOrCreateAuthor(context.Background(), "Bara")

	assert.NoError(t, err)
	assert.Equal(t, existing, result)
	mockRepo.AssertExpectations(t)
}

func TestGetOrCreateAuthor_LookupError(t *testing.T) {
	mockRepo := new(mocks.MockAuthorRepo)
	svc := author.NewAuthorService(mockRepo)

	mockRepo.On("GetAuthorByName", mock.Anything, "Bara").Return(nil, errors.New("db unavailable"))

	result, err := svc.GetOrCreateAuthor(context.Background(), "Bara")

	assert.ErrorIs(t, err, author.ErrInternalDBError)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetOrCreateAuthor_CreateFails(t *testing.T) {
	mockRepo := new(mocks.MockAuthorRepo)
	svc := author.NewAuthorService(mockRepo)

	mockRepo.On("GetAuthorByName", mock.Anything, "Bara").Return(nil, sql.ErrNoRows)
	mockRepo.On("CreateAuthor", mock.Anything, mock.MatchedBy(func(a *author.Author) bool {
		return a.Name == "Bara"
	})).Return(nil, errors.New("insert failed"))

	result, err := svc.GetOrCreateAuthor(context.Background(), "Bara")

	assert.ErrorIs(t, err, author.ErrInternalDBError)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestGetOrCreateAuthor_CreatesNew(t *testing.T) {
	mockRepo := new(mocks.MockAuthorRepo)
	svc := author.NewAuthorService(mockRepo)

	newAuthor := &author.Author{ID: "auth-2", Name: "Bara"}

	mockRepo.On("GetAuthorByName", mock.Anything, "Bara").Return(nil, sql.ErrNoRows)
	mockRepo.On("CreateAuthor", mock.Anything, mock.MatchedBy(func(a *author.Author) bool {
		return a.Name == "Bara"
	})).Return(newAuthor, nil)

	result, err := svc.GetOrCreateAuthor(context.Background(), "Bara")

	assert.NoError(t, err)
	assert.Equal(t, newAuthor, result)
	mockRepo.AssertExpectations(t)
}
