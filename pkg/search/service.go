package search

import (
	"context"
	"fmt"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

const (
	ArticleIndexName = "articles"
)

// SearchService defines the interface for generic Elasticsearch operations.
// It exposes methods for indexing and performing general searches.
type SearchService interface {
	// IndexDocument adds or updates a document in a specified index.
	IndexDocument(ctx context.Context, indexName string, id string, doc interface{}) error
	// SearchDocuments performs a search using a provided Elasticsearch query.
	// It returns the raw search result which can then be processed by the caller.
	SearchDocuments(ctx context.Context, indexName string, query elastic.Query, from, size int, sort_asc bool, by string) (*elastic.SearchResult, error)
	// Close closes the underlying Elasticsearch client connection (if needed).
	Close()
}

// elasticSearchService implements the SearchService interface.
type elasticSearchService struct {
	client *elastic.Client
}

// NewSearchService creates a new instance of SearchService.
func NewSearchService(esClient *elastic.Client) SearchService {
	return &elasticSearchService{client: esClient}
}

// IndexDocument implementation
func (s *elasticSearchService) IndexDocument(ctx context.Context, indexName string, id string, doc interface{}) error {
	_, err := s.client.Index().
		Index(indexName).
		Id(id).
		BodyJson(doc).
		Do(ctx)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"index": indexName,
			"id":    id,
		}).Error("Failed to index document in Elasticsearch")
		return fmt.Errorf("failed to index document: %w", err)
	}
	logrus.WithFields(logrus.Fields{"index": indexName, "id": id}).Info("Document indexed in Elasticsearch")
	return nil
}

// SearchDocuments implementation
func (s *elasticSearchService) SearchDocuments(ctx context.Context, indexName string, query elastic.Query, from, size int, sort_asc bool, by string) (*elastic.SearchResult, error) {
	searchService := s.client.Search().
		Index(indexName).
		Query(query).
		From(from).
		Size(size)

	if by != "" {
		searchService.Sort(by, sort_asc)
	}

	searchResult, err := searchService.Do(ctx)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"index": indexName,
			"from":  from,
			"size":  size,
		}).Error("Elasticsearch search failed")
		return nil, fmt.Errorf("failed to perform search: %w", err)
	}
	return searchResult, nil
}

// Close implementation
func (s *elasticSearchService) Close() {
	s.client.Stop()
}
