// Package db
package db

import (
	"context"
)

const (
	cTxsDailyStats = "TransactionDailyStats"
)

type TxsDaily interface {
}

func (m *mongoDB) UpsertDailyTxs(ctx context.Context, time, txs int64, amount string) error {
	return nil
}

func (m *mongoDB) TxsDailyStats(ctx context.Context) {

}
