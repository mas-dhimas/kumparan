package api

import (
	"net/http"
	"strconv"

	"kumparan-test/internal/article"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	CustomIDHeaderKeys = "Custom-ID"
)

type Handler struct {
	articleService article.Service
}

func NewHandler(articleSvc article.Service) *Handler {
	return &Handler{
		articleService: articleSvc,
	}
}

// RegisterRoutes registers the API routes with the provided router.
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.Add("GET", "/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "I'm alive")
	})

	v1 := e.Group("/api/v1")

	articles := v1.Group("/articles")
	articles.POST("", h.PostArticle)
	articles.GET("", h.GetArticles)
}

// PostArticle handles the creation of a new article.
// @Summary Post a new article
// @Description Creates a new news article with a title, body, and author.
// @Tags articles
// @Accept json
// @Produce json
// @Param article body article.CreateArticleRequest true "Article object to be created"
// @Success 201 {object} article.Article "Successfully created article"
// @Failure 400 {object} ErrorResponse "Invalid request payload or missing fields"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /articles [post]
func (h *Handler) PostArticle(e echo.Context) error {
	var req article.CreateArticleRequest

	if err := e.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload or malformed JSON")
	}

	if req.Title == "" || req.Body == "" || req.Author == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing required fields: title, body, and author are mandatory")
	}

	createdArticle, err := h.articleService.PostArticle(e.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create article due to internal error")
	}

	return e.JSON(http.StatusCreated, createdArticle)
}

// GetArticles handles retrieving a list of articles.
// @Summary Get a list of articles
// @Description Retrieves a list of news articles, sorted by latest first, with optional filters.
// @Tags articles
// @Accept json
// @Produce json
// @Param query query string false "Keywords to search in article title and body"
// @Param author query string false "Filter by author's name"
// @Param page query int false "Page number for pagination (default 1)"
// @Param limit query int false "Number of articles per page (default 10, max 100)"
// @Success 200 {array} article.Article "Successfully retrieved list of articles"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /articles [get]
func (h *Handler) GetArticles(e echo.Context) error {
	filter := &article.ArticleFilter{
		Query:  e.QueryParam("query"),
		Author: e.QueryParam("author"),
		Page:   parseIntOrDefault(e.Request().URL.Query().Get("page"), 1),
		Limit:  parseIntOrDefault(e.Request().URL.Query().Get("limit"), 10),
	}

	articles, err := h.articleService.GetArticles(e.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve articles due to internal error")
	}

	return e.JSON(http.StatusOK, articles)
}

// ErrorResponse represents a standardized error response.
type ErrorResponse struct {
	Message string `json:"message"`
}

// parseIntOrDefault parses a string to an int, returning a default value on error.
func parseIntOrDefault(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		logrus.WithError(err).Warnf("Invalid integer parameter: %s, using default %d", s, defaultValue)
		return defaultValue
	}
	return val
}
