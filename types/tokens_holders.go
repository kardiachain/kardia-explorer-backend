package types

type TokenHolder struct {
	TokenName       string  `json:"tokenName,omitempty" bson:"tokenName"`
	TokenSymbol     string  `json:"tokenSymbol,omitempty" bson:"tokenSymbol"`
	TokenDecimals   int64   `json:"tokenDecimals" bson:"tokenDecimals"`
	Logo            string  `json:"logo,omitempty" bson:"-"`
	ContractAddress string  `json:"contractAddress,omitempty" bson:"contractAddress"`
	HolderAddress   string  `json:"holderAddress" bson:"holderAddress"`
	HolderName      string  `json:"holderName" bson:"-"`
	BalanceString   string  `json:"balance" bson:"balance"`
	BalanceFloat    float64 `json:"-" bson:"balanceFloat"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}
