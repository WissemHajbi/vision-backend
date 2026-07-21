package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository describes storage operations without exposing SQLite to the
// business layer. Another database could implement this interface later.
type Repository interface {
	GetByQRCode(context.Context, string) (Product, error)
	Create(context.Context, Product) (Product, error)
}

var _ Repository = (*SQLiteRepository)(nil)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) GetByQRCode(ctx context.Context, qrCode string) (Product, error) {
	var product Product
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, qr_code, name, price_cents FROM products WHERE qr_code = ?`,
		qrCode,
	).Scan(&product.ID, &product.QRCode, &product.Name, &product.PriceCents)
	if errors.Is(err, sql.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, product Product) (Product, error) {
	result, err := r.db.ExecContext(
		ctx,
		`INSERT INTO products (qr_code, name, price_cents) VALUES (?, ?, ?)`,
		product.QRCode, product.Name, product.PriceCents,
	)
	if err != nil {
		// Checking first gives callers a stable domain error while the UNIQUE
		// constraint still protects against concurrent duplicate inserts.
		if _, lookupErr := r.GetByQRCode(ctx, product.QRCode); lookupErr == nil {
			return Product{}, ErrAlreadyExists
		}
		return Product{}, fmt.Errorf("create product: %w", err)
	}

	product.ID, err = result.LastInsertId()
	if err != nil {
		return Product{}, fmt.Errorf("read inserted product id: %w", err)
	}
	return product, nil
}
