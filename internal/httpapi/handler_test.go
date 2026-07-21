package httpapi_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/vision-products-api/internal/database"
	"github.com/example/vision-products-api/internal/httpapi"
	"github.com/example/vision-products-api/internal/product"
)

func TestCreateThenGetProduct(t *testing.T) {
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	service := product.NewService(product.NewSQLiteRepository(db))
	handler := httpapi.NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))

	create := httptest.NewRequest(http.MethodPost, "/api/v1/products", strings.NewReader(
		`{"qr_code":"6191234567890","name":"Cola","price":3.50}`,
	))
	create.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, create)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createResponse.Code, createResponse.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/products?qrcode=6191234567890", nil)
	getResponse := httptest.NewRecorder()
	handler.ServeHTTP(getResponse, get)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", getResponse.Code, getResponse.Body.String())
	}
	if body := getResponse.Body.String(); !strings.Contains(body, `"name":"Cola"`) || !strings.Contains(body, `"price":3.5`) {
		t.Fatalf("unexpected response: %s", body)
	}
}

func TestGetMissingProductReturns404(t *testing.T) {
	db, err := database.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	service := product.NewService(product.NewSQLiteRepository(db))
	handler := httpapi.NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/api/v1/products?qrcode=missing", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
