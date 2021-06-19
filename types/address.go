package types

type AddressInfo struct {
	Hash         string  `json:"hash" bson:"hash"`
	Balance      string  `json:"balance" bson:"balance"`
	BalanceIndex float64 `json:"-" bson:"balanceIndex"`
	Name         string  `json:"name" bson:"name"`

	// AddressType [Wallet, KRC20, Contract, Staking, Validator, Params]
	AddressType string `json:"type" bson:"type"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}

type Address struct {
	Address        string  `json:"address" bson:"address,omitempty"`
	Rank           uint64  `json:"rank,omitempty"`
	IndexByBalance float64 `json:"-" bson:"indexByBalance,omitempty"`
	BalanceFloat   float64 `json:"-" bson:"balanceFloat,omitempty"`        // low precise balance for sorting purposes
	BalanceString  string  `json:"balance" bson:"balanceString,omitempty"` // high precise balance for API
	Name           string  `json:"name" bson:"name,omitempty"`             // alias of an address
	Info           string  `json:"info,omitempty" bson:"-"`                // additional info of this address
	Logo           string  `json:"logo" bson:"-"`

	// Token
	TokenName   string `json:"tokenName" bson:"-"`
	TokenSymbol string `json:"tokenSymbol" bson:"-"`
	Decimals    int64  `json:"decimals" bson:"-"`
	TotalSupply string `json:"totalSupply" bson:"-"`

	// SMC
	IsContract   bool   `json:"isContract" bson:"-"`
	KrcTypes     string `json:"type" bson:"type,omitempty"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress"`

	// Stats
	TxCount         int `json:"txCount,omitempty" bson:"-"`
	HolderCount     int `json:"holderCount,omitempty" bson:"-"`
	InternalTxCount int `json:"internalTxCount,omitempty" bson:"-"`
	TokenTxCount    int `json:"tokenTxCount,omitempty" bson:"-"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
}

type UpdateAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
