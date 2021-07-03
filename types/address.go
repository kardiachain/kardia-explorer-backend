package types

const (
	AddressTypeWallet  = 1
	AddressTypeKRC20   = 2
	AddressTypeKRC721  = 3
	AddressTypeStaking = 4
)

type Address struct {
	Address       string  `json:"address" bson:"address,omitempty"`
	Rank          uint64  `json:"rank,omitempty"`
	BalanceFloat  float64 `json:"-" bson:"balanceFloat,omitempty"`        // low precise balance for sorting purposes
	BalanceString string  `json:"balance" bson:"balanceString,omitempty"` // high precise balance for API
	Name          string  `json:"name" bson:"name,omitempty"`             // alias of an address
	Info          string  `json:"info,omitempty" bson:"info,omitempty"`   // additional info of this address
	Logo          string  `json:"logo" bson:"logo,omitempty"`

	// Token
	TokenName   string `json:"tokenName" bson:"tokenName,omitempty"`
	TokenSymbol string `json:"tokenSymbol" bson:"tokenSymbol,omitempty"`
	Decimals    int64  `json:"decimals" bson:"decimals,omitempty"`
	TotalSupply string `json:"totalSupply" bson:"totalSupply,omitempty"`

	// SMC
	IsContract   bool   `json:"isContract" bson:"isContract,omitempty"`
	KrcTypes     string `json:"type" bson:"type,omitempty"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress,omitempty"`

	// Stats
	TxCount         int `json:"txCount,omitempty" bson:"txCount,omitempty"`
	HolderCount     int `json:"holderCount,omitempty" bson:"holderCount,omitempty"`
	InternalTxCount int `json:"internalTxCount,omitempty" bson:"internalTxCount,omitempty"`
	TokenTxCount    int `json:"tokenTxCount,omitempty" bson:"tokenTxCount,omitempty"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
}

type UpdateAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
