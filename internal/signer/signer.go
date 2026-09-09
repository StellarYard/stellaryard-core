package signer

import "context"

// Signer is the interface for signing Stellar transactions.
// Core never holds or transmits a raw secret key outside of this boundary.
type Signer interface {
	// Sign signs an unsigned transaction XDR and returns the signed XDR.
	Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (signedTxXDR string, err error)

	// PublicKeys returns all public keys managed by this signer.
	PublicKeys(ctx context.Context) ([]string, error)
}
