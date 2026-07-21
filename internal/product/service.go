package product

import (
	"context"
	"strings"
)

// Service contains business rules between HTTP and storage.
type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetByQRCode(ctx context.Context, qrCode string) (Product, error) {
	qrCode = strings.TrimSpace(qrCode)
	if qrCode == "" {
		return Product{}, ErrInvalidInput
	}
	return s.repository.GetByQRCode(ctx, qrCode)
}

func (s *Service) Create(ctx context.Context, qrCode, name string, priceCents int64) (Product, error) {
	candidate := Product{
		QRCode:     strings.TrimSpace(qrCode),
		Name:       strings.TrimSpace(name),
		PriceCents: priceCents,
	}
	if candidate.QRCode == "" || candidate.Name == "" || candidate.PriceCents <= 0 {
		return Product{}, ErrInvalidInput
	}
	return s.repository.Create(ctx, candidate)
}
