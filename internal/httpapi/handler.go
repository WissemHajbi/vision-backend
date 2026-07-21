// Package httpapi translates HTTP requests into product service calls.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"

	"github.com/example/vision-products-api/internal/product"
)

type Handler struct {
	products *product.Service
	logger   *slog.Logger
}

type productResponse struct {
	ID     int64   `json:"id"`
	QRCode string  `json:"qr_code"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}

type createProductRequest struct {
	QRCode string  `json:"qr_code"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(products *product.Service, logger *slog.Logger) http.Handler {
	h := &Handler{products: products, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/products", h.getProduct)
	mux.HandleFunc("POST /api/v1/products", h.createProduct)
	mux.HandleFunc("GET /healthz", h.health)
	return h.logRequests(mux)
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	found, err := h.products.GetByQRCode(r.Context(), r.URL.Query().Get("qrcode"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toResponse(found))
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	// Limit input size so a client cannot make the server allocate an
	// unbounded request body.
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()

	var input createProductRequest
	if err := decoder.Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "body must contain one JSON object"})
		return
	}

	priceCents, valid := dollarsToCents(input.Price)
	if !valid {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "price must be positive with at most two decimal places"})
		return
	}

	created, err := h.products.Create(r.Context(), input.QRCode, input.Name, priceCents)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toResponse(created))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, product.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "qr_code, name, and a positive price are required"})
	case errors.Is(err, product.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "product not found"})
	case errors.Is(err, product.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "a product with this qr_code already exists"})
	default:
		h.logger.Error("request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func (h *Handler) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func dollarsToCents(price float64) (int64, bool) {
	if math.IsNaN(price) || math.IsInf(price, 0) || price <= 0 || price > 10_000_000 {
		return 0, false
	}
	cents := math.Round(price * 100)
	if math.Abs(price*100-cents) > 0.000001 {
		return 0, false
	}
	return int64(cents), true
}

func toResponse(value product.Product) productResponse {
	return productResponse{
		ID: value.ID, QRCode: value.QRCode, Name: value.Name,
		Price: float64(value.PriceCents) / 100,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
