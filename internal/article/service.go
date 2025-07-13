package article

import (
	"context"
	"fmt"
	"time"

	"kumparan-test/internal/author"
	"kumparan-test/pkg/search"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

type Service interface {
	PostArticle(ctx context.Context, req *CreateArticleRequest) (*Article, error)
	GetArticles(ctx context.Context, filter *ArticleFilter) ([]*Article, error)
}

type articleService struct {
	repo          Repository
	authorService author.Service
	esClient      search.SearchService
}

func NewArticleService(repo Repository, authorSvc author.Service, esClient search.SearchService) Service {
	return &articleService{
		repo:          repo,
		authorService: authorSvc,
		esClient:      esClient,
	}
}

func (s *articleService) PostArticle(ctx context.Context, req *CreateArticleRequest) (*Article, error) {

	authorObj, err := s.authorService.GetOrCreateAuthor(ctx, req.Author)
	if err != nil {
		logrus.WithError(err).Error("Failed to get or create author for article")
		return nil, fmt.Errorf("%w: failed to resolve author", err) // Wrap and return original error
	}

	article := &Article{
		Title: req.Title,
		Body:  req.Body,
		Author: author.Author{
			ID:   authorObj.ID,
			Name: req.Author,
		},
		AuthorID:  authorObj.ID,
		CreatedAt: time.Now(),
	}

	createdArticle, err := s.repo.CreateArticle(ctx, article)
	if err != nil {
		logrus.Errorf("Service failed to create article in DB, err : %s", err)
		return nil, fmt.Errorf("failed to post article: %w", err)
	}

	// Index in Elasticsearch (synchronously for simplicity)
	// In a high-throughput system, this would be asynchronous via a message queue
	// to avoid blocking the API response and ensure reliability.
	esDoc := map[string]interface{}{
		"id":         createdArticle.ID,
		"title":      createdArticle.Title,
		"body":       createdArticle.Body,
		"author":     createdArticle.Author.Name,
		"created_at": createdArticle.CreatedAt,
	}
	err = s.esClient.IndexDocument(ctx, search.ArticleIndexName, createdArticle.ID, esDoc)
	if err != nil {
		logrus.WithError(err).WithField("article_id", createdArticle.ID).
			Error("Failed to index article in Elasticsearch")
	}

	logrus.WithField("article_id", createdArticle.ID).Info("Article indexed in Elasticsearch")

	return createdArticle, nil
}

func (s *articleService) GetArticles(ctx context.Context, filter *ArticleFilter) ([]*Article, error) {
	// Set default pagination values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	articles := []*Article{}
	var err error

	if filter.Query != "" {
		logrus.WithField("query", filter.Query).Info("Performing Elasticsearch search")
		searchResult, err := s.esClient.SearchDocuments(
			ctx,
			search.ArticleIndexName,
			(elastic.NewMultiMatchQuery(filter.Query, "title", "body")),
			(filter.Page-1)*filter.Limit,
			filter.Limit,
			false,
			"published_at",
		)
		if err != nil {
			logrus.WithError(err).Error("Elasticsearch search failed")
			return nil, fmt.Errorf("failed to search articles: %w", err)
		}

		if searchResult.Hits.TotalHits.Value > 0 {
			var articleIDs []string
			for _, hit := range searchResult.Hits.Hits {
				articleIDs = append(articleIDs, hit.Id)
			}
			// Fetch full articles from PostgreSQL using IDs from Elasticsearch
			articles, err = s.repo.GetArticlesByID(ctx, filter, articleIDs)
			if err != nil {
				logrus.WithError(err).Error("Failed to retrieve full articles from DB after ES search")
				return nil, fmt.Errorf("failed to retrieve articles details: %w", err)
			}

		}
	} else {
		logrus.WithField("filter", fmt.Sprintf("%#v", *filter)).Info("Performing PostgreSQL query for articles")
		articles, err = s.repo.GetArticles(ctx, filter)
		if err != nil {
			logrus.Errorf("Service failed to get articles from DB, err : %s", err)
			return nil, fmt.Errorf("failed to get articles: %w", err)
		}
	}

	return articles, nil
}
