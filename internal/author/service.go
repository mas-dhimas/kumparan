package author

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	ErrInternalDBError = errors.New("internal database error")
)

type Service interface {
	GetOrCreateAuthor(ctx context.Context, name string) (*Author, error)
}

type authorService struct {
	repo Repository
}

func NewAuthorService(repo Repository) Service {
	return &authorService{repo: repo}
}

// GetOrCreateAuthor attempts to get an author by name; if not found, it creates them.
func (s *authorService) GetOrCreateAuthor(ctx context.Context, name string) (*Author, error) {
	author, err := s.repo.GetAuthorByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Author not found, create new
			newAuthor := &Author{
				Name: name,
			}
			createdAuthor, createErr := s.repo.CreateAuthor(ctx, newAuthor)
			if createErr != nil {
				logrus.WithError(createErr).Error("Failed to create new author in DB")
				return nil, ErrInternalDBError
			}
			logrus.WithField("author_id", createdAuthor.ID).Info("New author created successfully")
			return createdAuthor, nil
		}

		logrus.WithError(err).Error("Failed to lookup author by name in DB")
		return nil, ErrInternalDBError
	}

	return author, nil
}
