package models

import "time"

// LedgerSnapshot represents a point-in-time view of the ledger.
type LedgerSnapshot struct {
	Sequence  uint32    `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
	TxCount   int       `json:"txCount"`
}

// Transaction represents a ledger transaction.
type Transaction struct {
	Hash      string    `json:"hash"`
	Sequence  uint32    `json:"sequence"`
	Source    string    `json:"source"`
	Fee       int64     `json:"fee"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"createdAt"`
}
