package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	"kamino-simulador/internal/model"
	solanapkg "kamino-simulador/pkg/solana"
)

var ErrAccountNotFound = errors.New("account not found on Solana")
var ErrKaminoDiscoveryFailed = errors.New("kamino discovery failed for all profiles")

var (
	kaminoLendingProgramID         = solana.MustPublicKeyFromBase58("KLend2g3cP87fffoy8q1mQqGKjrxjC8boSyAYavgmjD")
	kaminoLendingStagingProgramID  = solana.MustPublicKeyFromBase58("SLendK7ySfcEzyaFqy93gDnD3RtrpXJcnRwb6zFHJSh")
	kaminoFlexLendProgramID        = solana.MustPublicKeyFromBase58("FL3X2pRsQ9zHENpZSKDRREtccwJuei8yg9fwDu9UN69Q")
	obligationAccountDiscriminator = solana.Base58("VEdzkJnDweW") // sha256("account:Obligation")[:8]
)

const (
	obligationAccountDataSizeV1 = uint64(1792) // 8-byte discriminator + 1784-byte account
	obligationAccountDataSizeV2 = uint64(3344) // 8-byte discriminator + 3336-byte account
	obligationOwnerOffsetV1     = uint64(64)   // discriminator(8) + tag(8) + last_update(16) + lending_market(32)
	obligationOwnerOffsetV0     = uint64(56)   // fallback in case accounts are stored without discriminator
)

type SolanaRepository struct {
	client *solanapkg.RPCClient
}

func NewSolanaRepository(client *solanapkg.RPCClient) *SolanaRepository {
	return &SolanaRepository{client: client}
}

func (r *SolanaRepository) GetAccountDataByWallet(ctx context.Context, wallet solana.PublicKey) ([]byte, error) {
	res, err := r.client.GetAccountInfoWithOpts(ctx, wallet, &solanarpc.GetAccountInfoOpts{
		Commitment: solanarpc.CommitmentFinalized,
		Encoding:   solana.EncodingBase64,
	})
	if err != nil {
		return nil, fmt.Errorf("RPC getAccountInfo failed: %w", err)
	}
	if res == nil || res.Value == nil {
		return nil, ErrAccountNotFound
	}

	data := res.Value.Data.GetBinary()
	return data, nil
}

func (r *SolanaRepository) DiscoverKaminoObligationsByWallet(
	ctx context.Context,
	wallet solana.PublicKey,
	fullScan bool,
) ([]model.KaminoObligationAccount, model.DiscoveryDiagnostics, error) {
	type queryProfile struct {
		programID    solana.PublicKey
		dataSize     uint64
		ownerOffset  uint64
		useAnchorDis bool
	}

	baseProfiles := []queryProfile{
		{programID: kaminoLendingProgramID, dataSize: obligationAccountDataSizeV2, ownerOffset: obligationOwnerOffsetV1, useAnchorDis: true},
		{programID: kaminoLendingProgramID, dataSize: obligationAccountDataSizeV1, ownerOffset: obligationOwnerOffsetV1, useAnchorDis: true},
	}
	extendedProfiles := []queryProfile{
		{programID: kaminoLendingProgramID, dataSize: obligationAccountDataSizeV2, ownerOffset: obligationOwnerOffsetV0, useAnchorDis: false},
		{programID: kaminoLendingProgramID, dataSize: obligationAccountDataSizeV1, ownerOffset: obligationOwnerOffsetV0, useAnchorDis: false},
		{programID: kaminoLendingStagingProgramID, dataSize: obligationAccountDataSizeV2, ownerOffset: obligationOwnerOffsetV1, useAnchorDis: true},
		{programID: kaminoLendingStagingProgramID, dataSize: obligationAccountDataSizeV1, ownerOffset: obligationOwnerOffsetV1, useAnchorDis: true},
		{programID: kaminoLendingStagingProgramID, dataSize: obligationAccountDataSizeV2, ownerOffset: obligationOwnerOffsetV0, useAnchorDis: false},
		{programID: kaminoLendingStagingProgramID, dataSize: obligationAccountDataSizeV1, ownerOffset: obligationOwnerOffsetV0, useAnchorDis: false},
		{programID: kaminoFlexLendProgramID, dataSize: obligationAccountDataSizeV2, ownerOffset: obligationOwnerOffsetV1, useAnchorDis: true},
		{programID: kaminoFlexLendProgramID, dataSize: obligationAccountDataSizeV1, ownerOffset: obligationOwnerOffsetV1, useAnchorDis: true},
		{programID: kaminoFlexLendProgramID, dataSize: obligationAccountDataSizeV2, ownerOffset: obligationOwnerOffsetV0, useAnchorDis: false},
		{programID: kaminoFlexLendProgramID, dataSize: obligationAccountDataSizeV1, ownerOffset: obligationOwnerOffsetV0, useAnchorDis: false},
	}
	profiles := baseProfiles
	if fullScan {
		profiles = append(profiles, extendedProfiles...)
	}

	diagnostics := model.DiscoveryDiagnostics{
		Profiles: make([]model.DiscoveryProfileResult, 0, len(profiles)),
	}
	dedup := make(map[string]struct{})
	result := make([]model.KaminoObligationAccount, 0)
	for _, profile := range profiles {
		diagnostics.ProfilesTried++
		profileName := fmt.Sprintf("program=%s size=%d ownerOffset=%d discriminator=%t", profile.programID.String(), profile.dataSize, profile.ownerOffset, profile.useAnchorDis)
		filters := []solanarpc.RPCFilter{
			{DataSize: profile.dataSize},
			{
				Memcmp: &solanarpc.RPCFilterMemcmp{
					Offset: profile.ownerOffset,
					Bytes:  solana.Base58(wallet.String()),
				},
			},
		}
		if profile.useAnchorDis {
			filters = append([]solanarpc.RPCFilter{
				{
					Memcmp: &solanarpc.RPCFilterMemcmp{
						Offset: 0,
						Bytes:  obligationAccountDiscriminator,
					},
				},
			}, filters...)
		}

		accounts, err := r.getProgramAccountsWithRetry(ctx, profile.programID, &solanarpc.GetProgramAccountsOpts{
			Commitment: solanarpc.CommitmentFinalized,
			Encoding:   solana.EncodingBase64,
			Filters:    filters,
		})
		if err != nil {
			diagnostics.ProfilesFailed++
			diagnostics.Profiles = append(diagnostics.Profiles, model.DiscoveryProfileResult{
				Profile: profileName,
				Matches: 0,
				Error:   err.Error(),
			})
			continue
		}
		diagnostics.Profiles = append(diagnostics.Profiles, model.DiscoveryProfileResult{
			Profile: profileName,
			Matches: len(accounts),
		})

		for _, account := range accounts {
			if account == nil || account.Account == nil {
				continue
			}
			key := account.Pubkey.String()
			if _, exists := dedup[key]; exists {
				continue
			}
			dedup[key] = struct{}{}
			result = append(result, model.KaminoObligationAccount{
				Address: account.Pubkey,
				Data:    account.Account.Data.GetBinary(),
			})
		}
	}

	if diagnostics.ProfilesFailed == diagnostics.ProfilesTried {
		return nil, diagnostics, fmt.Errorf("%w: rpc errors on %d/%d profiles", ErrKaminoDiscoveryFailed, diagnostics.ProfilesFailed, diagnostics.ProfilesTried)
	}

	return result, diagnostics, nil
}

func (r *SolanaRepository) getProgramAccountsWithRetry(
	ctx context.Context,
	programID solana.PublicKey,
	opts *solanarpc.GetProgramAccountsOpts,
) (solanarpc.GetProgramAccountsResult, error) {
	var lastErr error
	wait := 300 * time.Millisecond
	for range 4 {
		accounts, err := r.client.GetProgramAccountsWithOpts(ctx, programID, opts)
		if err == nil {
			return accounts, nil
		}
		lastErr = err
		if !isTooManyRequestsError(err) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		wait *= 2
	}
	return nil, lastErr
}

func isTooManyRequestsError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "429") || strings.Contains(msg, "too many requests")
}
