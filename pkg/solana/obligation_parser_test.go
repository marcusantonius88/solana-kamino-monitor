package solana

import (
	"testing"

	"github.com/gagliardetto/solana-go"
)

func TestParseKaminoObligationAccount_V1Layout(t *testing.T) {
	t.Parallel()

	data := make([]byte, obligationLayoutV1.minimumDataSize)
	copy(data[:8], obligationDiscriminator[:])

	owner := solana.MustPublicKeyFromBase58("HX7qXRFZhgBFmJdE46BnsLEvtLdb14cBh1rMZiAA1x8C")
	copy(data[obligationOwnerOffset:obligationOwnerOffset+32], owner.Bytes())

	putUint128(data, obligationDepositedValueSFOffset, 111)
	putUint128(data, obligationBorrowFactorAdjustedDebtSFOffsetV1, 222)
	putUint128(data, obligationBorrowedAssetsValueSFOffsetV1, 333)
	putUint128(data, obligationAllowedBorrowValueSFOffsetV1, 444)
	putUint128(data, obligationUnhealthyBorrowValueSFOffsetV1, 555)

	parsed, err := ParseKaminoObligationAccount(solana.PublicKey{}, data)
	if err != nil {
		t.Fatalf("ParseKaminoObligationAccount returned error: %v", err)
	}

	if parsed.Owner != owner {
		t.Fatalf("owner mismatch: got %s want %s", parsed.Owner, owner)
	}
	assertUint128Value(t, parsed.DepositedValueSF, 111)
	assertUint128Value(t, parsed.BorrowFactorAdjustedDebtSF, 222)
	assertUint128Value(t, parsed.BorrowedAssetsMarketValueSF, 333)
	assertUint128Value(t, parsed.AllowedBorrowValueSF, 444)
	assertUint128Value(t, parsed.UnhealthyBorrowValueSF, 555)
}

func TestParseKaminoObligationAccount_V2Layout(t *testing.T) {
	t.Parallel()

	data := make([]byte, obligationLayoutV2.minimumDataSize)
	copy(data[:8], obligationDiscriminator[:])

	owner := solana.MustPublicKeyFromBase58("HX7qXRFZhgBFmJdE46BnsLEvtLdb14cBh1rMZiAA1x8C")
	copy(data[obligationOwnerOffset:obligationOwnerOffset+32], owner.Bytes())

	putUint128(data, obligationDepositedValueSFOffset, 111)
	putUint128(data, obligationBorrowFactorAdjustedDebtSFOffsetV2, 222)
	putUint128(data, obligationBorrowedAssetsValueSFOffsetV2, 333)
	putUint128(data, obligationAllowedBorrowValueSFOffsetV2, 444)
	putUint128(data, obligationUnhealthyBorrowValueSFOffsetV2, 555)

	parsed, err := ParseKaminoObligationAccount(solana.PublicKey{}, data)
	if err != nil {
		t.Fatalf("ParseKaminoObligationAccount returned error: %v", err)
	}

	if parsed.Owner != owner {
		t.Fatalf("owner mismatch: got %s want %s", parsed.Owner, owner)
	}
	assertUint128Value(t, parsed.DepositedValueSF, 111)
	assertUint128Value(t, parsed.BorrowFactorAdjustedDebtSF, 222)
	assertUint128Value(t, parsed.BorrowedAssetsMarketValueSF, 333)
	assertUint128Value(t, parsed.AllowedBorrowValueSF, 444)
	assertUint128Value(t, parsed.UnhealthyBorrowValueSF, 555)
}

func putUint128(data []byte, offset int, value uint64) {
	for i := range 8 {
		data[offset+i] = byte(value >> (8 * i))
	}
}

func assertUint128Value(t *testing.T, got Uint128, want uint64) {
	t.Helper()

	if got.Hi != 0 || got.Lo != want {
		t.Fatalf("uint128 mismatch: got hi=%d lo=%d want lo=%d", got.Hi, got.Lo, want)
	}
}
