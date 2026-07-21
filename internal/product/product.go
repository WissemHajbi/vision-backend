// Package product contains the product business rules.
package product

import "errors"

var (
	ErrNotFound      = errors.New("product not found")
	ErrAlreadyExists = errors.New("product already exists")
	ErrInvalidInput  = errors.New("invalid product input")
)

// Product uses integer cents internally. Integers avoid floating-point money
// errors such as 0.1 + 0.2 not being represented exactly.
type Product struct {
	ID         int64
	QRCode     string
	Name       string
	PriceCents int64
}
