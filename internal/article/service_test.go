package article_test

import (
	"context"
	"fmt"
	"kumparan-test/internal/article"
	"kumparan-test/internal/article/mocks"
	"kumparan-test/internal/author"
	"kumparan-test/pkg/search"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostArticle_Success(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	req := &article.CreateArticleRequest{
		Title:  "Hello",
		Body:   "World",
		Author: "Matahari",
	}

	authorObj := &author.Author{ID: "author-1", Name: "Matahari"}
	mockAuthor.On("GetOrCreateAuthor", mock.Anything, "Matahari").Return(authorObj, nil)

	createdArticle := &article.Article{
		ID:       "article-1",
		Title:    "Hello",
		Body:     "World",
		AuthorID: authorObj.ID,
		Author:   *authorObj,
	}
	mockRepo.On("CreateArticle", mock.Anything, mock.MatchedBy(func(a *article.Article) bool {
		return a.Title == "Hello" && a.AuthorID == "author-1"
	})).Return(createdArticle, nil)

	mockSearch.On("IndexDocument", mock.Anything, search.ArticleIndexName, "article-1", mock.Anything).Return(nil)

	_, err := service.PostArticle(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockAuthor.AssertExpectations(t)
	mockSearch.AssertExpectations(t)
}

func TestGetArticles_WithQuery_UsesElasticsearch(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	filter := &article.ArticleFilter{
		Query: "Go testing",
		Page:  1,
		Limit: 10,
	}

	esHits := []*elastic.SearchHit{
		{Id: "article-1"},
	}
	esResult := &elastic.SearchResult{
		Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{Value: 1}, Hits: esHits},
	}

	mockSearch.On("SearchDocuments", mock.Anything, search.ArticleIndexName, mock.Anything, 0, 10, false, "published_at").
		Return(esResult, nil)

	mockRepo.On("GetArticlesByID", mock.Anything, filter, []string{"article-1"}).
		Return([]*article.Article{
			{ID: "article-1", Title: "Test", Body: "Body"},
		}, nil)

	articles, err := service.GetArticles(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Equal(t, "article-1", articles[0].ID)

	mockRepo.AssertExpectations(t)
	mockSearch.AssertExpectations(t)
}

func TestGetArticles_NoQuery_UsesPostgreSQL(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	filter := &article.ArticleFilter{
		Query:  "",
		Page:   -1,
		Limit:  -10,
		Author: "Bob",
	}

	mockRepo.On("GetArticles", mock.Anything, filter).
		Return([]*article.Article{
			{ID: "article-2", Title: "PostgreSQL", Body: "Article"},
		}, nil)

	articles, err := service.GetArticles(context.Background(), filter)

	assert.NoError(t, err)
	assert.Len(t, articles, 1)
	assert.Equal(t, "article-2", articles[0].ID)

	mockRepo.AssertExpectations(t)
}

func TestPostArticle_AuthorServiceFails(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	req := &article.CreateArticleRequest{Title: "X", Body: "Y", Author: "Fail"}

	mockAuthor.On("GetOrCreateAuthor", mock.Anything, "Fail").
		Return((*author.Author)(nil), fmt.Errorf("db down"))

	_, err := service.PostArticle(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve author")

	mockAuthor.AssertExpectations(t)
}

func TestPostArticle_DBCreateFails(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	req := &article.CreateArticleRequest{Title: "Title", Body: "Body", Author: "Author"}
	authorObj := &author.Author{ID: "auth1", Name: "Author"}

	mockAuthor.On("GetOrCreateAuthor", mock.Anything, "Author").Return(authorObj, nil)
	mockRepo.On("CreateArticle", mock.Anything, mock.Anything).Return((*article.Article)(nil), fmt.Errorf("insert failed"))

	_, err := service.PostArticle(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to post article")

	mockRepo.AssertExpectations(t)
	mockAuthor.AssertExpectations(t)
}

func TestPostArticle_ESIndexFailsButStillReturnsArticle(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	req := &article.CreateArticleRequest{Title: "Title", Body: "Body", Author: "Matahari"}
	authorObj := &author.Author{ID: "auth-1", Name: "Matahari"}
	articleObj := &article.Article{
		ID:       "art-1",
		Title:    req.Title,
		Body:     req.Body,
		AuthorID: authorObj.ID,
		Author:   *authorObj,
	}

	mockAuthor.On("GetOrCreateAuthor", mock.Anything, "Matahari").Return(authorObj, nil)
	mockRepo.On("CreateArticle", mock.Anything, mock.Anything).Return(articleObj, nil)
	mockSearch.On("IndexDocument", mock.Anything, search.ArticleIndexName, "art-1", mock.Anything).
		Return(fmt.Errorf("ES unavailable"))

	result, err := service.PostArticle(context.Background(), req)

	assert.NoError(t, err) // ES failure should not block article creation
	assert.Equal(t, "art-1", result.ID)

	mockRepo.AssertExpectations(t)
	mockSearch.AssertExpectations(t)
}

func TestGetArticles_ESFails(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	filter := &article.ArticleFilter{
		Query: "fail search",
		Page:  1,
		Limit: 10,
	}

	mockSearch.On("SearchDocuments", mock.Anything, search.ArticleIndexName, mock.Anything, 0, 10, false, "published_at").
		Return((*elastic.SearchResult)(nil), fmt.Errorf("es timeout"))

	_, err := service.GetArticles(context.Background(), filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to search articles")

	mockSearch.AssertExpectations(t)
}

func TestGetArticles_ESWorks_ButDBFails(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	filter := &article.ArticleFilter{
		Query: "elastic",
		Page:  1,
		Limit: 10,
	}

	esHits := []*elastic.SearchHit{{Id: "id-1"}}
	esResult := &elastic.SearchResult{
		Hits: &elastic.SearchHits{TotalHits: &elastic.TotalHits{Value: 1}, Hits: esHits},
	}

	mockSearch.On("SearchDocuments", mock.Anything, search.ArticleIndexName, mock.Anything, 0, 10, false, "published_at").
		Return(esResult, nil)

	mockRepo.On("GetArticlesByID", mock.Anything, filter, []string{"id-1"}).
		Return(([]*article.Article)(nil), fmt.Errorf("db failure"))

	_, err := service.GetArticles(context.Background(), filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve articles details")

	mockRepo.AssertExpectations(t)
}

func TestGetArticles_PGOnly_Fails(t *testing.T) {
	mockRepo := new(mocks.MockRepo)
	mockAuthor := new(mocks.MockAuthorService)
	mockSearch := new(mocks.MockSearchService)

	service := article.NewArticleService(mockRepo, mockAuthor, mockSearch)

	filter := &article.ArticleFilter{
		Query:  "",
		Page:   1,
		Limit:  500,
		Author: "Bob",
	}

	mockRepo.On("GetArticles", mock.Anything, filter).
		Return(([]*article.Article)(nil), fmt.Errorf("pg error"))

	_, err := service.GetArticles(context.Background(), filter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get articles")

	mockRepo.AssertExpectations(t)
}
