package models

import "time"

// ContractDeployment represents a deployed Soroban contract.
type ContractDeployment struct {
	ID         string    `json:"id"`
	WASMHash   string    `json:"wasmHash"`
	ContractID string    `json:"contractId"`
	DeployedBy string    `json:"deployedBy"` // Account.PublicKey
	Network    string    `json:"network"`
	CreatedAt  time.Time `json:"createdAt"`
}

// DeployContractRequest is the request body for POST /contracts/deploy.
type DeployContractRequest struct {
	WASMPath string `json:"wasmPath"`
}

// InvokeContractRequest is the request body for POST /contracts/{contractId}/invoke.
type InvokeContractRequest struct {
	Method string   `json:"method"`
	Args   []string `json:"args"`
}
