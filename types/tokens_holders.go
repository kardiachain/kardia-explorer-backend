package types

type TokenHolder struct {
	HolderAddress string `json:"holderAddress,omitempty" bson:"holderAddress"`
	HolderName    string `json:"holderName,omitempty" bson:"-"`

	TokenAddress string `json:"tokenAddress" bson:"tokenAddress"`
	TokenName    string `json:"tokenName,omitempty" bson:"tokenName"`
	//TokenSymbol     string  `json:"tokenSymbol,omitempty" bson:"tokenSymbol"`
	//TokenDecimals   int64   `json:"tokenDecimals" bson:"tokenDecimals"`
	//Logo            string  `json:"logo,omitempty" bson:"-"`
	ContractAddress string  `json:"contractAddress,omitempty" bson:"contractAddress"`
	BalanceString   string  `json:"balance,omitempty" bson:"balance"`
	BalanceFloat    float64 `json:"-" bson:"balanceFloat"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}
