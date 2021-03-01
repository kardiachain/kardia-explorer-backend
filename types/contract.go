// Package types
package types

//Contract define simple information about a SMC in kardia system
type Contract struct {
	Name         string `json:"name" bson:"name"`
	Address      string `json:"address" bson:"address"`
	Bytecode     string `json:"bytecode" bson:"bytecode"`
	ABI          string `json:"abi" bson:"abi"`
	OwnerAddress string `json:"ownerAddress" bson:"ownerAddress"`
	TxHash       string `json:"txHash" bson:"txHash"`
	CreatedAt    string `json:"createdAt" bson:"createdAt"`
}
