// Package types
package types

const (
	ContractTypeNormal     = 1
	ContractTypeStaking    = 2
	ContractTypeParams     = 3
	ContractTypeValidators = 4
)

// Contract define simple information about a SMC in kardia system
type Contract struct {
	Name         string `json:"name" bson:"name"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress"`
	Address      string `json:"address" bson:"address"`
	Bytecode     string `json:"bytecode,omitempty" bson:"bytecode"`
	ABI          string `json:"abi" bson:"abi"`
	TxHash       string `json:"txHash,omitempty" bson:"txHash"`
	Info         string `json:"info" bson:"info"`
	Logo         string `json:"logo" bson:"logo"`
	IsVerified   bool   `json:"isVerified" bson:"isVerified"`
	Type         string `json:"type" bson:"type"`
	ContractType int    `json:"contractType" bson:"contractType"`
	CreatedAt    int64  `json:"createdAt" bson:"createdAt"`
}

type ContractABI struct {
	Type string `json:"type" bson:"type"`
	ABI  string `json:"abi" bson:"abi"`
}
