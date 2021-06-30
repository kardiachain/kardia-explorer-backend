// Package types
package types

// Contract define simple information about a SMC in kardia system
type Contract struct {
	Name         string `json:"name" bson:"name,omitempty"`
	Address      string `json:"address" bson:"address,omitempty"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress,omitempty"`
	TxHash       string `json:"txHash,omitempty" bson:"txHash,omitempty"`
	Type         string `json:"type" bson:"type,omitempty"`
	Info         string `json:"info" bson:"info,omitempty"`

	// TokenInfo
	Symbol      string `json:"symbol" bson:"symbol,omitempty"`
	TotalSupply string `json:"totalSupply" bson:"totalSupply,omitempty"`
	Decimals    int64  `json:"decimals" bson:"decimals,omitempty"`
	Logo        string `json:"logo" bson:"logo"`

	// Addition information
	IsVerified bool `json:"isVerified" bson:"isVerified"`
	Status     int  `json:"status" bson:"status,omitempty"`

	// Source information
	Bytecode        string `json:"bytecode,omitempty" bson:"bytecode,omitempty"`
	ABI             string `json:"abi" bson:"abi,omitempty"`
	Source          string `json:"source" bson:"source,omitempty"`
	CompilerVersion string `json:"compilerVersion" bson:"compilerVersion,omitempty"`
	IsOptimize      bool   `json:"isOptimize" bson:"isOptimize,omitempty"`

	CreatedAt int64 `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
}

type ContractABI struct {
	Type string `json:"type" bson:"type"`
	ABI  string `json:"abi" bson:"abi"`
}
