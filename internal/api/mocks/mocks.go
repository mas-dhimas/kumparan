package mocks

import (
	"context"
	"kumparan-test/internal/article"

	"github.com/stretchr/testify/mock"
)

// MockArticleService for testing
type MockArticleService struct {
	mock.Mock
}

func (m *MockArticleService) PostArticle(ctx context.Context, req *article.CreateArticleRequest) (*article.Article, error) {
	args := m.Called(ctx, req)
	if result := args.Get(0); result != nil {
		return result.(*article.Article), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockArticleService) GetArticles(ctx context.Context, filter *article.ArticleFilter) ([]*article.Article, error) {
	args := m.Called(ctx, filter)
	if result := args.Get(0); result != nil {
		return result.([]*article.Article), args.Error(1)
	}
	return nil, args.Error(1)
}
