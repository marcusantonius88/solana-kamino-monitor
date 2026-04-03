package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"kamino-simulador/internal/model"
	solanapkg "kamino-simulador/pkg/solana"
)

var ErrInvalidWallet = errors.New("invalid Solana wallet address")

type PositionResult struct {
	HealthFactor *float64               `json:"healthFactor,omitempty"`
	Debug        *PositionDebugResponse `json:"debug,omitempty"`
	Diagnostics  PositionDiagnostics    `json:"-"`
}

type PositionDiagnostics struct {
	ProfilesTried     int
	ProfilesFailed    int
	ObligationsFound  int
	ObligationsParsed int
}

type PositionDebugResponse struct {
	ProfilesTried     int      `json:"profilesTried"`
	ProfilesFailed    int      `json:"profilesFailed"`
	ObligationsFound  int      `json:"obligationsFound"`
	ObligationsParsed int      `json:"obligationsParsed"`
	Reason            string   `json:"reason,omitempty"`
	ProfileErrors     []string `json:"profileErrors,omitempty"`
}

type SolanaRepository interface {
	GetAccountDataByWallet(ctx context.Context, wallet solana.PublicKey) ([]byte, error)
	DiscoverKaminoObligationsByWallet(
		ctx context.Context,
		wallet solana.PublicKey,
		fullScan bool,
	) ([]model.KaminoObligationAccount, model.DiscoveryDiagnostics, error)
}

type PositionService struct {
	solanaRepository SolanaRepository
}

func NewPositionService(solanaRepository SolanaRepository) *PositionService {
	return &PositionService{solanaRepository: solanaRepository}
}

func (s *PositionService) GetWalletPosition(ctx context.Context, walletAddress string, debug bool) (PositionResult, error) {
	wallet, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		return PositionResult{}, fmt.Errorf("%w: %v", ErrInvalidWallet, err)
	}

	_, err = s.solanaRepository.GetAccountDataByWallet(ctx, wallet)
	if err != nil {
		return PositionResult{}, err
	}

	// Phase 2: Discover Kamino obligations (PDA discovery)
	// Phase 3: Decode obligation account data
	// Phase 4: Calculate and return health factor.
	obligationAccounts, discoveryDiag, err := s.solanaRepository.DiscoverKaminoObligationsByWallet(ctx, wallet, debug)
	if err != nil {
		return PositionResult{}, err
	}

	parsedObligations := make([]solanapkg.ParsedObligationAccount, 0, len(obligationAccounts))
	for _, obligationAccount := range obligationAccounts {
		parsed, parseErr := solanapkg.ParseKaminoObligationAccount(obligationAccount.Address, obligationAccount.Data)
		if parseErr != nil {
			return PositionResult{}, fmt.Errorf("failed to decode Kamino obligation account %s: %w", obligationAccount.Address.String(), parseErr)
		}
		parsedObligations = append(parsedObligations, parsed)
	}

	// Phase 4 will use parsedObligations to calculate health factor.
	healthFactor, ok := calculateWalletHealthFactor(parsedObligations)

	diag := PositionDiagnostics{
		ProfilesTried:     discoveryDiag.ProfilesTried,
		ProfilesFailed:    discoveryDiag.ProfilesFailed,
		ObligationsFound:  len(obligationAccounts),
		ObligationsParsed: len(parsedObligations),
	}

	if !ok {
		reason := "no active debt found in discovered obligations"
		if diag.ProfilesFailed > 0 {
			reason = "inconclusive: some discovery profiles failed (likely RPC rate limit)"
		}
		result := PositionResult{Diagnostics: diag}
		if debug {
			result.Debug = &PositionDebugResponse{
				ProfilesTried:     diag.ProfilesTried,
				ProfilesFailed:    diag.ProfilesFailed,
				ObligationsFound:  diag.ObligationsFound,
				ObligationsParsed: diag.ObligationsParsed,
				Reason:            reason,
				ProfileErrors:     collectProfileErrors(discoveryDiag),
			}
		}
		return result, nil
	}

	result := PositionResult{
		HealthFactor: &healthFactor,
		Diagnostics:  diag,
	}
	if debug {
		result.Debug = &PositionDebugResponse{
			ProfilesTried:     diag.ProfilesTried,
			ProfilesFailed:    diag.ProfilesFailed,
			ObligationsFound:  diag.ObligationsFound,
			ObligationsParsed: diag.ObligationsParsed,
			ProfileErrors:     collectProfileErrors(discoveryDiag),
		}
	}
	return result, nil
}

func collectProfileErrors(d model.DiscoveryDiagnostics) []string {
	out := make([]string, 0)
	for _, p := range d.Profiles {
		if p.Error == "" {
			continue
		}
		out = append(out, fmt.Sprintf("%s: %s", p.Profile, p.Error))
	}
	return out
}

func calculateWalletHealthFactor(obligations []solanapkg.ParsedObligationAccount) (float64, bool) {
	if len(obligations) == 0 {
		return 0, false
	}

	// If a wallet has multiple obligations, we expose the riskiest one (lowest health factor).
	minHF := math.Inf(1)
	found := false

	for _, obligation := range obligations {
		if obligation.BorrowFactorAdjustedDebtSF.IsZero() {
			continue
		}

		// Kamino UI definition: HF = Liq. LTV / LTV.
		// Using obligation fields:
		// Liq. LTV = unhealthy_borrow_value / deposited_value
		// LTV      = borrowed_assets_market_value / deposited_value
		// => HF    = unhealthy_borrow_value / borrowed_assets_market_value
		numerator := new(big.Float).SetInt(obligation.UnhealthyBorrowValueSF.BigInt())
		denominator := new(big.Float).SetInt(obligation.BorrowedAssetsMarketValueSF.BigInt())
		if denominator.Sign() == 0 {
			continue
		}

		hfBig := new(big.Float).Quo(numerator, denominator)
		hf, _ := hfBig.Float64()
		if math.IsNaN(hf) || math.IsInf(hf, 0) {
			continue
		}

		if hf < minHF {
			minHF = hf
		}
		found = true
	}

	if !found {
		return 0, false
	}
	return roundFloat(minHF, 2), true
}

func roundFloat(value float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(value*pow) / pow
}
