package types

type TokenHolder struct {
	TokenName       string `json:"tokenName" bson:"tokenName"`
	TokenSymbol     string `json:"tokenSymbol" bson:"tokenSymbol"`
	TokenDecimals   int64  `json:"tokenDecimals" bson:"-"`
	Logo            string `json:"logo" bson:"-"`
	ContractAddress string `json:"contractAddress" bson:"contractAddress"`
	HolderAddress   string `json:"holderAddress" bson:"holderAddress"`
	BalanceString   string `json:"balance" bson:"balance"`
	BalanceFloat    int64  `json:"-" bson:"balanceFloat"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}
