package signer

import "context"

// Signer is the interface for signing Stellar transactions and managing keys.
// Core never holds or transmits a raw secret key outside of this boundary.
type Signer interface {
	// Sign signs an unsigned transaction XDR and returns the signed XDR.
	Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (signedTxXDR string, err error)

	// PublicKeys returns all public keys managed by this signer.
	PublicKeys(ctx context.Context) ([]string, error)

	// GenerateKeypair generates and stores a new keypair, returning the public address.
	GenerateKeypair(ctx context.Context) (publicKey string, err error)

	// HasKey returns true if the signer manages the private key for the given public address.
	HasKey(ctx context.Context, publicKey string) bool
}
