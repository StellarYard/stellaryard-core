package models

import "time"

// Account represents a local/testnet keypair core manages.
// NOTE: SecretKey field only ever populated for LocalTestSigner-managed
// accounts. This model does not extend to mainnet accounts — a future
// mainnet account model should NOT have a SecretKey field at all.
type Account struct {
	ID        string    `json:"id"`
	PublicKey string    `json:"publicKey"`
	SecretKey string    `json:"secretKey,omitempty"` // testnet only; empty/absent for external-signer accounts
	Label     string    `json:"label"`
	Network   string    `json:"network"` // "local" | "testnet" — never "mainnet" in v1
	Balance   string    `json:"balance,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreateAccountRequest is the request body for POST /accounts.
type CreateAccountRequest struct {
	Label   string `json:"label"`
	Network string `json:"network"`
}
