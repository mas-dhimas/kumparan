package mocks

import (
	"context"

	"kumparan-test/internal/article"
	"kumparan-test/internal/author"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) CreateArticle(ctx context.Context, art *article.Article) (*article.Article, error) {
	args := m.Called(ctx, art)
	return args.Get(0).(*article.Article), args.Error(1)
}

func (m *MockRepo) GetArticles(ctx context.Context, filter *article.ArticleFilter) ([]*article.Article, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*article.Article), args.Error(1)
}

func (m *MockRepo) GetArticlesByID(ctx context.Context, filter *article.ArticleFilter, ids []string) ([]*article.Article, error) {
	args := m.Called(ctx, filter, ids)
	return args.Get(0).([]*article.Article), args.Error(1)
}

type MockAuthorService struct {
	mock.Mock
}

func (m *MockAuthorService) GetOrCreateAuthor(ctx context.Context, name string) (*author.Author, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*author.Author), args.Error(1)
}

type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) IndexDocument(ctx context.Context, indexName, id string, doc interface{}) error {
	args := m.Called(ctx, indexName, id, doc)
	return args.Error(0)
}

func (m *MockSearchService) SearchDocuments(ctx context.Context, indexName string, query elastic.Query, from, size int, sortAsc bool, by string) (*elastic.SearchResult, error) {
	args := m.Called(ctx, indexName, query, from, size, sortAsc, by)
	return args.Get(0).(*elastic.SearchResult), args.Error(1)
}

func (m *MockSearchService) Close() {
	m.Called()
}
