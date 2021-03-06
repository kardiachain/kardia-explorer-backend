package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func TestMGO_InternalTxs(t *testing.T) {
	db, err := GetMgo()
	assert.Nil(t, err)
	for i := 0; i < 100; i++ {
		//token/txs?page=1&limit=100&address=0xF5aEd64137C0fCaA596D4aF9dd2e33980a402901&contractAddress=0xb3b39589Cf5ECf173e5191cdef3563f7677E3703
		//https://backend-dex.kardiachain.io/api/v1/token/txs?address=0xd258f28642e8AEa592A2D914c1975bcA495FD931&contractAddress=0xb3b39589Cf5ECf173e5191cdef3563f7677E3703&page=1&limit=100
		start := time.Now()
		txs, total, err := db.GetListInternalTxs(context.Background(), &types.InternalTxsFilter{
			Pagination: &types.Pagination{
				Skip:  0,
				Limit: 25,
			},
			//Contract:        "0xB1a2F2A95Bc565bBd02634864F733f5FcC6615A7",
			Address: "0xAde9A316f1E430c7a6F7BE4eD42367979db8AaA0",
		})
		assert.Nil(t, err)
		fmt.Println("TotalTime", time.Now().Sub(start))
		fmt.Println("Total", total)
		fmt.Println("Txs", txs)
	}

}