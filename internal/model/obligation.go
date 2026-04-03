package model

import "github.com/gagliardetto/solana-go"

type KaminoObligationAccount struct {
	Address solana.PublicKey
	Data    []byte
}

type DiscoveryProfileResult struct {
	Profile string
	Matches int
	Error   string
}

type DiscoveryDiagnostics struct {
	ProfilesTried  int
	ProfilesFailed int
	Profiles       []DiscoveryProfileResult
}
