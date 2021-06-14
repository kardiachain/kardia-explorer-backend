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
	Address        string  `json:"address" bson:"address"`
	Rank           uint64  `json:"rank"`
	IndexByBalance float64 `json:"-" bson:"indexByBalance"`
	BalanceFloat   float64 `json:"-" bson:"balanceFloat"`                // low precise balance for sorting purposes
	BalanceString  string  `json:"balance" bson:"balanceString"`         // high precise balance for API
	Name           string  `json:"name" bson:"name"`                     // alias of an address
	Info           string  `json:"info,omitempty" bson:"info,omitempty"` // additional info of this address
	Logo           string  `json:"logo" bson:"logo"`

	// Token
	TokenName   string `json:"tokenName" bson:"tokenName"`
	TokenSymbol string `json:"tokenSymbol" bson:"tokenSymbol"`
	Decimals    int64  `json:"decimals" bson:"decimals"`
	TotalSupply string `json:"totalSupply" bson:"totalSupply"`

	// SMC
	IsContract   bool   `json:"isContract" bson:"isContract"`
	KrcTypes     string `json:"type" bson:"type"`
	OwnerAddress string `json:"ownerAddress,omitempty" bson:"ownerAddress"`

	// Stats
	TxCount         int `json:"txCount,omitempty" bson:"txCount"`
	HolderCount     int `json:"holderCount,omitempty" bson:"holderCount"`
	InternalTxCount int `json:"internalTxCount,omitempty" bson:"internalTxCount"`
	TokenTxCount    int `json:"tokenTxCount,omitempty" bson:"tokenTxCount"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}

type UpdateAddress struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}
