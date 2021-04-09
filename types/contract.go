// Package types
package types

// Contract define simple information about a SMC in kardia system
type Contract struct {
	Name         string `json:"name" bson:"name"`
	Address      string `json:"address" bson:"address"`
	Bytecode     string `json:"bytecode,omitempty" bson:"bytecode"`
	ABI          string `json:"abi" bson:"abi"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress"`
	TxHash       string `json:"txHash,omitempty" bson:"txHash"`
	CreatedAt    int64  `json:"createdAt" bson:"createdAt"`
	Type         string `json:"type" bson:"type"`
	Info         string `json:"info" bson:"info"`
	Logo         string `json:"logo" bson:"logo"`
	IsVerified   bool   `json:"isVerified" bson:"isVerified"`
}

type ContractABI struct {
	Type string `json:"type" bson:"type"`
	ABI  string `json:"abi" bson:"abi"`
}
