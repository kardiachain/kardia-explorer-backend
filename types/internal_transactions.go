package types

import "time"

// TokenTransfer represents a Transfer event emitted from an ERC20 or ERC721.
type TokenTransfer struct {
	TransferID      string `json:"transferID" bson:"transferID,omitempty"`
	TransactionHash string `json:"txHash" bson:"txHash,omitempty"`
	BlockHeight     uint64 `json:"blockHeight" bson:"blockHeight,omitempty"`
	Contract        string `json:"contractAddress" bson:"contractAddress,omitempty"`

	From     string      `json:"from" bson:"from,omitempty"`
	To       string      `json:"to" bson:"to,omitempty"`
	Value    string      `json:"value" bson:"value,omitempty"`
	TokenID  string      `json:"tokenID" bson:"tokenID,omitempty"`
	Time     time.Time   `json:"time" bson:"time,omitempty"`
	LogIndex interface{} `json:"logIndex" bson:"logIndex,omitempty"`
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
