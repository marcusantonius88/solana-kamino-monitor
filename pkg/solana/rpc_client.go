package solana

import (
	"context"

	"github.com/gagliardetto/solana-go"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
)

type RPCClient struct {
	client *solanarpc.Client
}

func NewRPCClient(endpoint string) *RPCClient {
	return &RPCClient{
		client: solanarpc.New(endpoint),
	}
}

func (c *RPCClient) GetAccountInfoWithOpts(
	ctx context.Context,
	account solana.PublicKey,
	opts *solanarpc.GetAccountInfoOpts,
) (*solanarpc.GetAccountInfoResult, error) {
	return c.client.GetAccountInfoWithOpts(ctx, account, opts)
}

func (c *RPCClient) GetProgramAccountsWithOpts(
	ctx context.Context,
	programID solana.PublicKey,
	opts *solanarpc.GetProgramAccountsOpts,
) (solanarpc.GetProgramAccountsResult, error) {
	return c.client.GetProgramAccountsWithOpts(ctx, programID, opts)
}
