package signer

import (
	"context"
	"strings"
	"testing"

	"github.com/stellar/go/keypair"
)

func TestLocalTestSignerGeneratesValidStellarKeypairs(t *testing.T) {
	s := NewLocalTestSigner()
	ctx := context.Background()

	pubKey, err := s.GenerateKeypair(ctx)
	if err != nil {
		t.Fatalf("GenerateKeypair failed: %v", err)
	}

	if !strings.HasPrefix(pubKey, "G") || len(pubKey) != 56 {
		t.Fatalf("Invalid Stellar public key format: %s", pubKey)
	}

	// Validate with official Stellar SDK
	parsed, err := keypair.ParseAddress(pubKey)
	if err != nil {
		t.Fatalf("keypair.ParseAddress failed for generated key: %v", err)
	}
	if parsed.Address() != pubKey {
		t.Errorf("parsed address %s != generated %s", parsed.Address(), pubKey)
	}

	// Signer must record the key
	if !s.HasKey(ctx, pubKey) {
		t.Errorf("signer.HasKey returned false for generated key %s", pubKey)
	}

	keys, err := s.PublicKeys(ctx)
	if err != nil {
		t.Fatalf("PublicKeys failed: %v", err)
	}
	if len(keys) != 1 || keys[0] != pubKey {
		t.Errorf("PublicKeys = %v, want [%s]", keys, pubKey)
	}
}

func TestLocalTestSignerSignsWithManagedKey(t *testing.T) {
	s := NewLocalTestSigner()
	ctx := context.Background()

	pubKey, err := s.GenerateKeypair(ctx)
	if err != nil {
		t.Fatalf("GenerateKeypair failed: %v", err)
	}

	txXDR := "AAAAA...unsigned_tx_bytes..."
	signedXDR, err := s.Sign(ctx, txXDR, pubKey)
	if err != nil {
		t.Fatalf("Sign failed with managed key: %v", err)
	}
	if signedXDR == "" {
		t.Error("expected non-empty signed XDR")
	}
}

func TestLocalTestSignerRejectsUnmanagedKey(t *testing.T) {
	s := NewLocalTestSigner()
	ctx := context.Background()

	// Generate a key that is NOT managed by this signer
	otherKP, err := keypair.Random()
	if err != nil {
		t.Fatalf("failed to generate other keypair: %v", err)
	}

	_, err = s.Sign(ctx, "unsigned", otherKP.Address())
	if err == nil {
		t.Error("Sign with unmanaged key should return error, got nil")
	}
}
