package repository

import (
	"encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
)

func TestRPCFilterMemcmpUsesBase58EncodedWalletBytes(t *testing.T) {
	t.Parallel()

	wallet := solana.MustPublicKeyFromBase58("HX7qXRFZhgBFmJdE46BnsLEvtLdb14cBh1rMZiAA1x8C")
	filter := solanarpc.RPCFilterMemcmp{
		Offset: obligationOwnerOffsetV1,
		Bytes:  solana.Base58(wallet.Bytes()),
	}

	payload, err := json.Marshal(filter)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}

	expected := `{"offset":64,"bytes":"HX7qXRFZhgBFmJdE46BnsLEvtLdb14cBh1rMZiAA1x8C"}`
	if string(payload) != expected {
		t.Fatalf("unexpected JSON payload: got %s want %s", payload, expected)
	}
}
