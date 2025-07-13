package article

import (
	"kumparan-test/internal/author"
	"time"
)

// Article represents the structure of a news article.
type Article struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Body      string        `json:"body"`
	AuthorID  string        `json:"author_id,omitempty"`
	Author    author.Author `json:"author"`
	CreatedAt time.Time     `json:"created_at"`
}

// CreateArticleRequest represents the request body for creating a new article.
type CreateArticleRequest struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Author string `json:"author"`
}

// ArticleFilter represents the optional query parameters for listing articles.
type ArticleFilter struct {
	Query  string // Keywords to search in title and body
	Author string // Filter by author's name
	Page   int    // For pagination (default 1)
	Limit  int    // For pagination (default 10)
}
