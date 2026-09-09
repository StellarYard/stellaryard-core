package signer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

// LocalTestSigner generates and holds testnet-only keypairs in memory.
// V1 implementation — signs transactions in-process using test keys.
type LocalTestSigner struct {
	mu   sync.RWMutex
	keys map[string]string // publicKey -> secretKey (hex-encoded placeholders)
}

// NewLocalTestSigner creates a new LocalTestSigner.
func NewLocalTestSigner() *LocalTestSigner {
	return &LocalTestSigner{
		keys: make(map[string]string),
	}
}

// Generate creates a new random keypair and returns the public key.
func (s *LocalTestSigner) Generate() string {
	pubBytes := make([]byte, 32)
	secBytes := make([]byte, 32)
	rand.Read(pubBytes)
	rand.Read(secBytes)

	publicKey := "G" + hex.EncodeToString(pubBytes)[:55]
	secretKey := "S" + hex.EncodeToString(secBytes)[:55]

	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[publicKey] = secretKey
	return publicKey
}

// Sign signs an unsigned transaction XDR using the keypair for the given public key.
// V1: Returns the unsigned XDR as-is — full signing requires Stellar SDK integration.
func (s *LocalTestSigner) Sign(ctx context.Context, unsignedTxXDR string, publicKey string) (string, error) {
	s.mu.RLock()
	_, ok := s.keys[publicKey]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("no keypair found for public key: %s", publicKey)
	}

	// V1 placeholder — actual signing will be implemented with Stellar SDK
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
