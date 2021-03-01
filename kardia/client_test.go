// Package kardia
package kardia

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/kardiachain/go-kardia/lib/abi"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/stretchr/testify/assert"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func TestInteractSMC(t *testing.T) {
	ctx := context.Background()
	dbClient, err := SetupMGOClient()
	assert.Nil(t, err)
	contract, err := dbClient.Contract(ctx, "0xFFA0Ba96DDd604DCA35F008Cf0aC6A0453670533")
	assert.Nil(t, err)

	fmt.Println("Contract", contract)
	//abiStr := fmt.Sprintf("%v", contract.ABI)
	//assert.Equal(t, testABIStr, abiStr)
	dataBytes, err := base64.StdEncoding.DecodeString(contract.ABI)
	assert.Nil(t, err)

	//r := bytes.NewBufferString(abiStr)
	r := bytes.NewReader(dataBytes)
	wheelABI, err := abi.JSON(r)
	assert.Nil(t, err)
	assert.NotNil(t, wheelABI)

	node, err := SetupNodeClient()
	assert.Nil(t, err)

	smc := SmcUtil{
		Abi:             &wheelABI,
		ContractAddress: common.HexToAddress("0xFFA0Ba96DDd604DCA35F008Cf0aC6A0453670533"),
	}

	payload, err := smc.Abi.Pack("reward", common.HexToAddress("0x4f36A53DC32272b97Ae5FF511387E2741D727bdb"))
	assert.Nil(t, err)
	res, err := node.KardiaCall(ctx, types.CallArgsJSON(constructCallArgs("0xFFA0Ba96DDd604DCA35F008Cf0aC6A0453670533", payload)))
	assert.Nil(t, err)
	type reward struct {
		Reward *big.Int
	}
	var result reward

	if err := smc.Abi.UnpackIntoInterface(&result, "reward", res); err != nil {
		fmt.Println("err", err)
		return
	}
	fmt.Println("A", result.Reward.Uint64())
	return
	//
	//pubKey, privateKey, err := SetupTestAccount()
	//assert.Nil(t, err)
	//fromAddress := crypto.PubkeyToAddress(*pubKey)
	//
	///*
	//var result uint64
	//	err := n.client.CallContext(ctx, &result, "account_nonce", common.HexToAddress(account))
	//	return result, err
	// */
	//// Now we can read the nonce that we should use for the account's transaction.
	//nonce, err := node.NonceAt(context.Background(), fromAddress.Hex())
	//assert.Nil(t, err)
	//gasLimit := uint64(3000000)
	//gasPrice := big.NewInt(1)
	//auth := NewKeyedTransactor(privateKey)
	//auth.Nonce = nonce
	//auth.Value = big.NewInt(0) // in wei
	//auth.GasLimit = gasLimit   // in units
	//auth.GasPrice = gasPrice
	//
	//_, err = smc.Transact(auth, "spin")
	//assert.Nil(t, err)
	////assert.Nil(b, err)
	////r := strings.NewReader(wheelABI)
	////abiData, err := abi.JSON(r)
	////assert.Nil(b, err)
	////
	////smc := NewContract(node, &abiData, common.HexToAddress(WheelSMCAddr))
	////
	////// run the Fib function b.N times
	////fmt.Println("TotalRun", b.N)
	////for n := 0; n < b.N; n++ {
	////	pubKey, privateKey, err := SetupTestAccount()
	////	assert.Nil(b, err)
	////	fromAddress := crypto.PubkeyToAddress(*pubKey)
	////	// Now we can read the nonce that we should use for the account's transaction.
	////	nonce, err := node.NonceAt(context.Background(), fromAddress.Hex())
	////	assert.Nil(b, err)
	////	gasLimit := uint64(3000000)
	////	gasPrice := big.NewInt(1)
	////	auth := NewKeyedTransactor(privateKey)
	////	auth.Nonce = nonce
	////	auth.Value = big.NewInt(0) // in wei
	////	auth.GasLimit = gasLimit   // in units
	////	auth.GasPrice = gasPrice
	////
	////	_, err = smc.Transact(auth, "spin")
	////	assert.Nil(b, err)
	//
	//// }
}

type SMCCallArgs struct {
	From     string   `json:"from"`     // the sender of the 'transaction'
	To       *string  `json:"to"`       // the destination contract (nil for contract creation)
	Gas      uint64   `json:"gas"`      // if 0, the call executes with near-infinite gas
	GasPrice *big.Int `json:"gasPrice"` // HYDRO <-> gas exchange ratio
	Value    *big.Int `json:"value"`    // amount of HYDRO sent along with the call
	Data     string   `json:"data"`     // input data, usually an ABI-encoded contract method invocation
}

func constructCallArgs(address string, payload []byte) SMCCallArgs {
	return SMCCallArgs{
		From:     address,
		To:       &address,
		Gas:      100000000,
		GasPrice: big.NewInt(0),
		Value:    big.NewInt(0),
		Data:     common.Bytes(payload).String(),
	}
}
