package kardia

import (
	"context"
	"fmt"

	kai "github.com/kardiachain/go-kardia"
	"github.com/kardiachain/go-kardia/rpc"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

// NewLogsFilter
func (ec *Client) NewLogsFilter(ctx context.Context, query kai.FilterQuery) (*rpc.ID, error) {
	return nil, nil
}

// UninstallFilter
func (ec *Client) UninstallFilter(ctx context.Context, filterID *rpc.ID) error {
	return nil
}

// GetFilterChanges
func (ec *Client) GetFilterChanges(ctx context.Context, filterID *rpc.ID) ([]*types.Log, error) {
	return nil, nil
}

// GetLogs
func (ec *Client) GetLogs(ctx context.Context, query kai.FilterQuery) ([]*types.Log, error) {
	var result []*types.Log
	arg, err := toFilterArg(query)
	if err != nil {
		return nil, err
	}
	err = ec.defaultClient.c.CallContext(ctx, &result, "kai_getLogs", arg)
	return result, err
}

func toFilterArg(q kai.FilterQuery) (interface{}, error) {
	arg := map[string]interface{}{
		"address": q.Addresses,
		"topics":  q.Topics,
	}
	if q.BlockHash != nil {
		arg["blockHash"] = *q.BlockHash
		if q.FromBlock != 0 || q.ToBlock != 0 {
			return nil, fmt.Errorf("cannot specify both BlockHash and FromBlock/ToBlock")
		}
	} else {
		if q.FromBlock == 0 {
			arg["fromBlock"] = uint64(1)
		} else {
			arg["fromBlock"] = q.FromBlock
		}
		arg["toBlock"] = q.ToBlock
		if q.ToBlock == 0 {
			arg["toBlock"] = "latest"
		}
	}
	return arg, nil
}
