package signer

import "context"

// Signer is the interface through which all transaction signing flows.
// Core never holds or transmits a raw secret key outside of this boundary.
//
// V1 implementation: LocalTestSigner (testnet-only keys in SQLite)
// Future implementation: ExternalSigner (hardware wallet, browser extension)
type Signer interface {
	// Sign returns a signed transaction envelope for the given unsigned
	// transaction, for the given public key. Implementations decide how.
	Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (signedTxXDR string, err error)

	// PublicKeys returns the public keys this signer can sign for.
	PublicKeys(ctx context.Context) ([]string, error)
}
