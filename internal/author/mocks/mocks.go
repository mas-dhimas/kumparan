package mocks

import (
	"context"

	"kumparan-test/internal/author"

	"github.com/stretchr/testify/mock"
)

type MockAuthorRepo struct {
	mock.Mock
}

func (m *MockAuthorRepo) GetAuthorByName(ctx context.Context, name string) (*author.Author, error) {
	args := m.Called(ctx, name)
	if a := args.Get(0); a != nil {
		return a.(*author.Author), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAuthorRepo) CreateAuthor(ctx context.Context, a *author.Author) (*author.Author, error) {
	args := m.Called(ctx, a)
	if a := args.Get(0); a != nil {
		return a.(*author.Author), args.Error(1)
	}
	return nil, args.Error(1)
}
