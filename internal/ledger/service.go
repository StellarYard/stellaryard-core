package ledger

import (
	"context"

	"github.com/StellarYard/stellaryard-core/internal/models"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// Service handles ledger inspection queries.
type Service struct {
	db *storage.DB
}

// NewService creates a new ledger service.
func NewService(db *storage.DB) *Service {
	return &Service{db: db}
}

// Snapshot returns the current ledger state summary.
func (svc *Service) Snapshot(ctx context.Context) (*models.LedgerSnapshot, error) {
	// TODO: implement - fetch from Horizon
	// NOTE: Horizon may lag after container start.
	// Consider whether to cache in SQLite or live-query Horizon.
	// See ROADMAP.md open question on LedgerSnapshot persistence.
	panic("not implemented")
}

// Transactions returns a paginated list of recent transactions.
func (svc *Service) Transactions(ctx context.Context, limit, offset int) ([]models.Transaction, error) {
	// TODO: implement - fetch from Horizon
	// NOTE: XDR decoding depth is undecided (see ROADMAP.md Phase 3).
	// Decide how much decoding belongs in core vs. left raw for consumers.
	panic("not implemented")
}
