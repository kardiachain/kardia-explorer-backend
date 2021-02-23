// Package types
package types

//Contract define simple information about a SMC in kardia system
type Contract struct {
	Address      string `json:"address" bson:"address"`
	Bytecode     string `json:"bytecode" bson:"bytecode"`
	ABI          string `json:"abi" bson:"abi"`
	OwnerAddress string `json:"ownerAddress" bson:"ownerAddress"`
	CreatedAt    string `json:"createdAt" bson:"createdAt"`
}
