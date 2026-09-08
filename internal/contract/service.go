package contract

import (
	"context"

	"github.com/StellarYard/stellaryard-core/internal/models"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// Service handles contract deployment and invocation.
type Service struct {
	db *storage.DB
}

// NewService creates a new contract service.
func NewService(db *storage.DB) *Service {
	return &Service{db: db}
}

// Deploy uploads WASM and deploys a contract to the local network.
func (svc *Service) Deploy(ctx context.Context, wasmPath, deployerPublicKey, network string) (*models.ContractDeployment, error) {
	// TODO: implement
	// 1. Read WASM file
	// 2. Upload to Soroban RPC
	// 3. Deploy contract
	// 4. Store deployment record in SQLite
	// NOTE: Validate WASM file before sending to core.
	// A corrupted file produces cryptic Soroban errors.
	panic("not implemented")
}

// Invoke calls a method on a deployed contract.
func (svc *Service) Invoke(ctx context.Context, contractID, method string, args []string) (string, error) {
	// TODO: implement
	// 1. Build XDR invocation
	// 2. Sign via Signer interface (injected)
	// 3. Submit to Soroban RPC
	// 4. Return result
	panic("not implemented")
}

// ListDeployments returns all recorded contract deployments.
func (svc *Service) ListDeployments(ctx context.Context) ([]models.ContractDeployment, error) {
	// TODO: implement
	panic("not implemented")
}
