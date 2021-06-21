package types

import "time"

type InternalTransaction struct {
	From            string    `json:"from" bson:"from"`
	To              string    `json:"to" bson:"to"`
	TokenAddress    string    `json:"tokenAddress" bson:"tokenAddress"`
	Value           string    `json:"value" bson:"value"`
	TransactionHash string    `json:"transactionHash" bson:"transactionHash"`
	BlockHeight     string    `json:"blockHeight" bson:"blockHeight"`
	Time            time.Time `json:"time" bson:"time"`
}

// TokenTransfer represents a Transfer event emitted from an ERC20 or ERC721.
type TokenTransfer struct {
	TransactionHash string `json:"txHash" bson:"txHash"`
	BlockHeight     uint64 `json:"blockHeight" bson:"blockHeight"`
	Contract        string `json:"contractAddress" bson:"contractAddress"`

	From     string      `json:"from" bson:"from"`
	To       string      `json:"to" bson:"to"`
	Value    string      `json:"value" bson:"value"`
	Time     time.Time   `json:"time" bson:"time"`
	LogIndex interface{} `json:"logIndex" bson:"logIndex"`
}

type KRCTokenInfo struct {
	Address     string `json:"tokenAddress"`
	TokenName   string `json:"tokenName"`
	TokenType   string `json:"tokenType"`
	TokenSymbol string `json:"tokenSymbol"`
	TotalSupply string `json:"totalSupply"`
	Decimals    int64  `json:"decimals"`
	Logo        string `json:"logo"`
	IsVerified  bool   `json:"isVerified" bson:"isVerified"`
}
