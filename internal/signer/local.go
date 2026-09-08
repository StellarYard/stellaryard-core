package signer

import (
	"context"

	"github.com/StellarYard/stellaryard-core/internal/storage"
)

// LocalTestSigner generates and holds testnet-only keypairs in SQLite,
// signs directly in-process. This is fine *only* because these keys
// never hold real value.
type LocalTestSigner struct {
	db *storage.DB
}

// NewLocalTestSigner creates a new LocalTestSigner backed by the given database.
func NewLocalTestSigner(db *storage.DB) *LocalTestSigner {
	return &LocalTestSigner{db: db}
}

// Sign signs a transaction using the testnet keypair for the given public key.
func (s *LocalTestSigner) Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (string, error) {
	// TODO: implement using stellar/go SDK
	// 1. Look up secret key for publicKey in SQLite
	// 2. Sign the XDR envelope using the secret key
	// 3. Return the signed XDR
	panic("not implemented")
}

// PublicKeys returns all public keys managed by this signer.
func (s *LocalTestSigner) PublicKeys(ctx context.Context) ([]string, error) {
	// TODO: implement - query all accounts from SQLite
	panic("not implemented")
}
