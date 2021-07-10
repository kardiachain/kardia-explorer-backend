package types

type KRC20Holder struct {
	TokenName       string  `json:"tokenName,omitempty" bson:"tokenName,omitempty"`
	TokenSymbol     string  `json:"tokenSymbol,omitempty" bson:"tokenSymbol,omitempty"`
	TokenDecimals   int64   `json:"tokenDecimals" bson:"tokenDecimals,omitempty"`
	Logo            string  `json:"logo,omitempty" bson:"-"`
	ContractAddress string  `json:"contractAddress,omitempty" bson:"contractAddress,omitempty"`
	HolderAddress   string  `json:"holderAddress" bson:"holderAddress,omitempty"`
	HolderName      string  `json:"holderName" bson:"-"`
	BalanceString   string  `json:"balance" bson:"balance,omitempty"`
	BalanceFloat    float64 `json:"-" bson:"balanceFloat,omitempty"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
}

type Holder struct {
	TokenName       string  `json:"tokenName,omitempty" bson:"tokenName,omitempty"`
	TokenSymbol     string  `json:"tokenSymbol,omitempty" bson:"tokenSymbol,omitempty"`
	TokenDecimals   int64   `json:"tokenDecimals" bson:"tokenDecimals,omitempty"`
	Logo            string  `json:"logo,omitempty" bson:"-"`
	ContractAddress string  `json:"contractAddress,omitempty" bson:"contractAddress,omitempty"`
	HolderAddress   string  `json:"holderAddress" bson:"holderAddress,omitempty"`
	HolderName      string  `json:"holderName" bson:"-"`
	BalanceString   string  `json:"balance" bson:"balance,omitempty"`
	BalanceFloat    float64 `json:"-" bson:"balanceFloat,omitempty"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
}

type KRC721Holder struct {
	HolderID        string `json:"holderID" bson:"holderID"`
	Address         string `json:"address" bson:"address"`
	ContractAddress string `json:"contractAddress" bson:"contractAddress"`
	TokenID         string `json:"tokenID" bson:"tokenID"`

	CreatedAt int64 `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt,omitempty"`
}
