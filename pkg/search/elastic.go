package search

import (
	"context"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

// NewElasticsearchClient initializes and returns a new Elasticsearch client.
func NewElasticsearchClient(url string) (*elastic.Client, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(10*time.Second),
		elastic.SetErrorLog(logrus.StandardLogger()),
		elastic.SetInfoLog(nil),
		elastic.SetTraceLog(nil),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, code, err := client.Ping(url).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping Elasticsearch: %w", err)
	}
	logrus.Infof("Elasticsearch connected to %s (version %s, code %d)", info.Name, info.Version.Number, code)

	// Create the index if it doesn't exist
	exists, err := client.IndexExists(ArticleIndexName).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check Elasticsearch index existence: %w", err)
	}
	if !exists {
		createIndex, err := client.CreateIndex(ArticleIndexName).BodyString(articleMapping).Do(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create Elasticsearch index: %w", err)
		}
		if !createIndex.Acknowledged {
			return nil, fmt.Errorf("failed to acknowledge Elasticsearch index creation")
		}
		logrus.Infof("Elasticsearch index '%s' created successfully", ArticleIndexName)
	} else {
		logrus.Infof("Elasticsearch index '%s' already exists", ArticleIndexName)
	}

	return client, nil
}

// articleMapping defines the mapping for the articles index.
// This helps Elasticsearch understand the data types and how to index them.
const articleMapping = `
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "id": { "type": "keyword" },
      "title": { "type": "text", "analyzer": "standard" },
      "body": { "type": "text", "analyzer": "standard" },
      "author": { "type": "keyword" },
      "published_at": { "type": "date" }
    }
  }
}
`
