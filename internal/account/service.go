package account

import (
	"context"

	"github.com/StellarYard/stellaryard-core/internal/models"
	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// Service handles account operations.
type Service struct {
	db     *storage.DB
	signer signer.Signer
}

// NewService creates a new account service.
func NewService(db *storage.DB, s signer.Signer) *Service {
	return &Service{db: db, signer: s}
}

// Create generates a new testnet keypair, funds it, and stores it.
func (svc *Service) Create(ctx context.Context, label, network string) (*models.Account, error) {
	// TODO: implement
	// 1. Generate keypair via signer
	// 2. Fund via Friendbot (testnet) or local genesis
	// 3. Store in SQLite
	// NOTE: Friendbot has rate limits and can be slow.
	// Implement retry with backoff. See ROADMAP.md for known pitfalls.
	panic("not implemented")
}

// List returns all managed accounts.
func (svc *Service) List(ctx context.Context) ([]models.Account, error) {
	// TODO: implement
	panic("not implemented")
}

// GetByPublicKey returns an account and its current balance from Horizon.
func (svc *Service) GetByPublicKey(ctx context.Context, publicKey string) (*models.Account, error) {
	// TODO: implement - query SQLite + fetch balance from Horizon
	// NOTE: Horizon may lag after container start. Balance may be stale.
	panic("not implemented")
}
