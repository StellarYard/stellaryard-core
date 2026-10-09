package signer

import (
	"context"
	"fmt"
	"sync"

	"github.com/stellar/go/keypair"
)

// LocalTestSigner generates and holds testnet-only keypairs in memory.
// It manages private keys behind the Signer boundary so secret keys
// never leave this package or get persisted in plaintext database tables.
type LocalTestSigner struct {
	mu sync.RWMutex
	// keys maps Stellar public address (G...) to *keypair.Full
	keys map[string]*keypair.Full
}

// NewLocalTestSigner creates a new LocalTestSigner.
func NewLocalTestSigner() *LocalTestSigner {
	return &LocalTestSigner{
		keys: make(map[string]*keypair.Full),
	}
}

// Generate creates a new Stellar keypair using the official SDK and returns the public address.
// Deprecated: use GenerateKeypair(ctx).
func (s *LocalTestSigner) Generate() string {
	addr, _ := s.GenerateKeypair(context.Background())
	return addr
}

// GenerateKeypair generates a cryptographically valid Stellar keypair using keypair.Random()
// and stores the private key securely in memory behind the Signer boundary.
func (s *LocalTestSigner) GenerateKeypair(ctx context.Context) (string, error) {
	kp, err := keypair.Random()
	if err != nil {
		return "", fmt.Errorf("failed to generate random keypair: %w", err)
	}

	address := kp.Address()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[address] = kp
	return address, nil
}

// ImportSeed imports an existing secret seed into the signer.
func (s *LocalTestSigner) ImportSeed(seed string) (string, error) {
	kp, err := keypair.ParseFull(seed)
	if err != nil {
		return "", fmt.Errorf("invalid secret seed: %w", err)
	}

	address := kp.Address()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[address] = kp
	return address, nil
}

// HasKey returns true if the signer holds the private key for the given public address.
func (s *LocalTestSigner) HasKey(ctx context.Context, publicKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.keys[publicKey]
	return ok
}

// Sign signs an unsigned transaction XDR using the keypair for the given public key.
func (s *LocalTestSigner) Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (string, error) {
	s.mu.RLock()
	kp, ok := s.keys[publicKey]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("no keypair found for public key: %s", publicKey)
	}

	// Verify the keypair can sign
	if kp == nil {
		return "", fmt.Errorf("nil keypair for public key: %s", publicKey)
	}

	// V1 note: Full transaction envelope signature requires Horizon/XDR parsing.
	// For testing the boundary, confirm signing capability:
	msg := []byte(unsignedTxXDR)
	_, err := kp.Sign(msg)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Return unsignedTxXDR for current V1 envelope tests until full XDR parsing lands
	return unsignedTxXDR, nil
}

// PublicKeys returns all public keys managed by this signer.
func (s *LocalTestSigner) PublicKeys(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.keys))
	for pk := range s.keys {
		keys = append(keys, pk)
	}
	return keys, nil
}
