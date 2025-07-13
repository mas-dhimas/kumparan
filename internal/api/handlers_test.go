package api_test

import (
	"encoding/json"
	"errors"
	"kumparan-test/internal/api"
	"kumparan-test/internal/article"
	"kumparan-test/internal/author"

	"kumparan-test/internal/api/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostArticle_Success(t *testing.T) {
	e := echo.New()
	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)

	reqBody := `{"title":"Test","body":"Content","author":"Bara"}`
	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	expected := &article.Article{
		ID:        "1",
		Title:     "Test",
		Body:      "Content",
		AuthorID:  "auth-1",
		Author:    author.Author{Name: "Bara"},
		CreatedAt: time.Now(),
	}

	mockSvc.On("PostArticle", mock.Anything, mock.AnythingOfType("*article.CreateArticleRequest")).Return(expected, nil)

	err := handler.PostArticle(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp article.Article
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.Equal(t, "Test", resp.Title)
}

func TestPostArticle_BadJSON(t *testing.T) {
	e := echo.New()
	handler := api.NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	err := handler.PostArticle(e.NewContext(req, rec))
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
}

func TestPostArticle_MissingFields(t *testing.T) {
	e := echo.New()
	handler := api.NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(`{"title":"T"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	err := handler.PostArticle(e.NewContext(req, rec))
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
}

func TestPostArticle_InternalError(t *testing.T) {
	e := echo.New()
	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)

	reqBody := `{"title":"T","body":"B","author":"A"}`
	req := httptest.NewRequest(http.MethodPost, "/articles", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)

	mockSvc.On("PostArticle", mock.Anything, mock.AnythingOfType("*article.CreateArticleRequest")).
		Return(nil, errors.New("db error"))

	err := handler.PostArticle(ctx)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
}

func TestGetArticles_Success(t *testing.T) {
	e := echo.New()
	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/articles?query=test&page=1&limit=2", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	mockSvc.On("GetArticles", mock.Anything, &article.ArticleFilter{
		Query:  "test",
		Author: "",
		Page:   1,
		Limit:  2,
	}).Return([]*article.Article{
		{ID: "1", Title: "T"},
	}, nil)

	err := handler.GetArticles(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetArticles_InternalError(t *testing.T) {
	e := echo.New()
	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/articles?query=err", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	mockSvc.On("GetArticles", mock.Anything, mock.AnythingOfType("*article.ArticleFilter")).
		Return(nil, errors.New("something bad"))

	err := handler.GetArticles(ctx)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
}

func TestGetArticles_InvalidPageLimit(t *testing.T) {
	e := echo.New()
	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/articles?page=abc&limit=def", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	mockSvc.On("GetArticles", mock.Anything, &article.ArticleFilter{
		Query:  "",
		Author: "",
		Page:   1,
		Limit:  10,
	}).Return([]*article.Article{}, nil)

	err := handler.GetArticles(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRegisterRoutes_Healthcheck(t *testing.T) {
	e := echo.New()

	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)
	handler.RegisterRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "I'm alive", rec.Body.String())
}

func TestRegisterRoutes_PostArticle_WiredCorrectly(t *testing.T) {
	e := echo.New()
	mockSvc := new(mocks.MockArticleService)
	handler := api.NewHandler(mockSvc)
	handler.RegisterRoutes(e)

	payload := `{"title":"Test","body":"Content","author":"Bara"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/articles", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	expected := &article.Article{
		ID:        "123",
		Title:     "Test",
		Body:      "Content",
		Author:    author.Author{Name: "Bara"},
		AuthorID:  "auth-1",
		CreatedAt: time.Now(),
	}
	mockSvc.On("PostArticle", mock.Anything, mock.Anything).Return(expected, nil)

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	mockSvc.AssertExpectations(t)
}
