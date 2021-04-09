package types

import "time"

// TokenTransfer represents a Transfer event emitted from an ERC20 or ERC721.
type TokenTransfer struct {
	TransactionHash string `json:"txHash" bson:"txHash"`
	Contract        string `json:"contractAddress" bson:"contractAddress"`

	From  string    `json:"from" bson:"from"`
	To    string    `json:"to" bson:"to"`
	Value string    `json:"value" bson:"value"`
	Time  time.Time `json:"time" bson:"time"`
}

type KRCTokenInfo struct {
	Address     string `json:"-"`
	TokenName   string `json:"tokenName"`
	TokenType   string `json:"tokenType"`
	TokenSymbol string `json:"tokenSymbol"`
	TotalSupply string `json:"totalSupply"`
	Decimals    int64  `json:"decimals"`
	Logo        string `json:"logo"`
	IsVerified  bool   `json:"isVerified" bson:"isVerified"`
}
