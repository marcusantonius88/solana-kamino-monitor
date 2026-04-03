package solana

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/gagliardetto/solana-go"
)

var (
	errInvalidObligationDataSize = errors.New("invalid Kamino obligation data size")
	errInvalidDiscriminator      = errors.New("invalid Kamino obligation discriminator")
)

var obligationDiscriminator = [8]byte{0xa8, 0xce, 0x8d, 0x6a, 0x58, 0x4c, 0xac, 0xa7}

const (
	obligationMinimumDataSize = obligationUnhealthyBorrowValueSFOffset + 16

	obligationOwnerOffset = 64

	obligationDepositedValueSFOffset           = 1192
	obligationBorrowFactorAdjustedDebtSFOffset = 2168
	obligationBorrowedAssetsValueSFOffset      = 2184
	obligationAllowedBorrowValueSFOffset       = 2200
	obligationUnhealthyBorrowValueSFOffset     = 2216
)

type ParsedObligationAccount struct {
	Address                     solana.PublicKey
	Owner                       solana.PublicKey
	DepositedValueSF            Uint128
	BorrowFactorAdjustedDebtSF  Uint128
	BorrowedAssetsMarketValueSF Uint128
	AllowedBorrowValueSF        Uint128
	UnhealthyBorrowValueSF      Uint128
}

type Uint128 struct {
	Lo uint64
	Hi uint64
}

func (u Uint128) IsZero() bool {
	return u.Lo == 0 && u.Hi == 0
}

func (u Uint128) BigInt() *big.Int {
	hi := new(big.Int).SetUint64(u.Hi)
	hi.Lsh(hi, 64)
	lo := new(big.Int).SetUint64(u.Lo)
	return hi.Add(hi, lo)
}

func ParseKaminoObligationAccount(address solana.PublicKey, data []byte) (ParsedObligationAccount, error) {
	if len(data) < obligationMinimumDataSize {
		return ParsedObligationAccount{}, fmt.Errorf("%w: expected at least %d, got %d", errInvalidObligationDataSize, obligationMinimumDataSize, len(data))
	}

	if !matchesDiscriminator(data[:8], obligationDiscriminator[:]) {
		return ParsedObligationAccount{}, errInvalidDiscriminator
	}

	owner := solana.PublicKeyFromBytes(data[obligationOwnerOffset : obligationOwnerOffset+32])

	return ParsedObligationAccount{
		Address:                     address,
		Owner:                       owner,
		DepositedValueSF:            readUint128LE(data, obligationDepositedValueSFOffset),
		BorrowFactorAdjustedDebtSF:  readUint128LE(data, obligationBorrowFactorAdjustedDebtSFOffset),
		BorrowedAssetsMarketValueSF: readUint128LE(data, obligationBorrowedAssetsValueSFOffset),
		AllowedBorrowValueSF:        readUint128LE(data, obligationAllowedBorrowValueSFOffset),
		UnhealthyBorrowValueSF:      readUint128LE(data, obligationUnhealthyBorrowValueSFOffset),
	}, nil
}

func readUint128LE(data []byte, offset int) Uint128 {
	return Uint128{
		Lo: binary.LittleEndian.Uint64(data[offset : offset+8]),
		Hi: binary.LittleEndian.Uint64(data[offset+8 : offset+16]),
	}
}

func matchesDiscriminator(data []byte, expected []byte) bool {
	if len(data) != len(expected) {
		return false
	}
	for i := range data {
		if data[i] != expected[i] {
			return false
		}
	}
	return true
}
