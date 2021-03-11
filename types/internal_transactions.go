package types

// TokenTransfer represents a Transfer event emitted from an ERC20 or ERC721.
type TokenTransfer struct {
	TransactionHash string `json:"transactionHash" bson:"transactionHash"`
	Contract        string `json:"contractAddress" bson:"contractAddress"`

	From        string `json:"fromAddress" bson:"fromAddress"`
	To          string `json:"toAddress" bson:"toAddress"`
	Value       string `json:"value" bson:"value"`
	BlockHeight uint64 `json:"blockHeight" bson:"blockHeight"`

	UpdatedAt int64 `json:"updatedAt" bson:"updatedAt"`
}

type KRCTokenInfo struct {
	Address     string `json:"-"`
	TokenName   string `json:"tokenName"`
	TokenType   string `json:"tokenType"`
	TokenSymbol string `json:"tokenSymbol"`
	TotalSupply string `json:"totalSupply"`
	Decimals    int64  `json:"decimals"`
	Logo        string `json:"logo"`
}
