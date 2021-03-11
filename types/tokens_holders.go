package types

type TokenHolder struct {
	TokenName       string `json:"tokenName,omitempty" bson:"tokenName"`
	TokenSymbol     string `json:"tokenSymbol,omitempty" bson:"tokenSymbol"`
	TokenDecimals   int64  `json:"tokenDecimals,omitempty" bson:"tokenDecimals"`
	Logo            string `json:"logo,omitempty" bson:"-"`
	ContractAddress string `json:"contractAddress,omitempty" bson:"contractAddress"`
	HolderAddress   string `json:"holderAddress" bson:"holderAddress"`
	BalanceString   string `json:"balance" bson:"balance"`
	BalanceFloat    int64  `json:"-" bson:"balanceFloat"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}
