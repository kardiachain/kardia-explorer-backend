package server

import (
	"time"

	coreTypes "github.com/kardiachain/go-kardia/types"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

type PagingResponse struct {
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Total uint64      `json:"total"`
	Data  interface{} `json:"data"`
}

type Blocks []SimpleBlock

type SimpleBlock struct {
	Height          uint64    `json:"height,omitempty"`
	Hash            string    `json:"hash,omitempty"`
	Time            time.Time `json:"time,omitempty"`
	ProposerAddress string    `json:"proposerAddress,omitempty"`
	ProposerName    string    `json:"proposerName"`
	NumTxs          uint64    `json:"numTxs"`
	GasLimit        uint64    `json:"gasLimit,omitempty"`
	GasUsed         uint64    `json:"gasUsed"`
	Rewards         string    `json:"rewards"`
}

type Block struct {
	types.Block
	ProposerName string `json:"proposerName"`
}

type Transactions []SimpleTransaction

type SimpleTransaction struct {
	Hash               string              `json:"hash"`
	BlockNumber        uint64              `json:"blockNumber"`
	Time               time.Time           `json:"time"`
	From               string              `json:"from"`
	FromName           string              `json:"fromName,omitempty"`
	To                 string              `json:"to"`
	ToName             string              `json:"toName,omitempty"`
	IsInValidatorsList bool                `json:"isInValidatorsList"`
	Role               int                 `json:"role"`
	ContractAddress    string              `json:"contractAddress,omitempty"`
	Value              string              `json:"value"`
	TxFee              string              `json:"txFee"`
	Status             uint                `json:"status"`
	DecodedInputData   *types.FunctionCall `json:"decodedInputData,omitempty"`
	InputData          string              `json:"input"`
}

type Transaction struct {
	BlockHash   string `json:"blockHash"`
	BlockNumber uint64 `json:"blockNumber"`

	Hash               string                 `json:"hash"`
	From               string                 `json:"from"`
	FromName           string                 `json:"fromName,omitempty"`
	To                 string                 `json:"to"`
	ToName             string                 `json:"toName,omitempty"`
	IsInValidatorsList bool                   `json:"isInValidatorsList"`
	Role               int                    `json:"role"`
	Status             uint                   `json:"status"`
	ContractAddress    string                 `json:"contractAddress"`
	Value              string                 `json:"value"`
	GasPrice           uint64                 `json:"gasPrice"`
	GasLimit           uint64                 `json:"gas"`
	GasUsed            uint64                 `json:"gasUsed"`
	TxFee              string                 `json:"txFee"`
	Nonce              uint64                 `json:"nonce"`
	Time               time.Time              `json:"time"`
	InputData          string                 `json:"input"`
	DecodedInputData   *types.FunctionCall    `json:"decodedInputData,omitempty"`
	Logs               []*InternalTransaction `json:"logs"`
	TransactionIndex   uint                   `json:"transactionIndex"`
	LogsBloom          coreTypes.Bloom        `json:"logsBloom"`
	Root               string                 `json:"root"`
	RevertReason       string                 `json:"revertReason"`
}

type NodeInfo struct {
	ID         string `json:"id"`
	Moniker    string `json:"moniker"`
	PeersCount int    `json:"peersCount"`
}

type Addresses []SimpleAddress

type SimpleAddress struct {
	Address            string `json:"address"` // low precise balance for sorting purposes
	BalanceString      string `json:"balance"` // high precise balance for API
	IsContract         bool   `json:"isContract"`
	Name               string `json:"name"`
	IsInValidatorsList bool   `json:"isInValidatorsList"`
	Role               int    `json:"role"`
	Rank               uint64 `json:"rank"`
}

type valInfoResponse struct {
	Name string
	Role int
}

type KRCTokenInfo struct {
	Name          string `json:"name" bson:"name,omitempty"`
	Address       string `json:"address" bson:"address,omitempty"`
	Bytecode      string `json:"bytecode,omitempty" bson:"bytecode,omitempty"`
	ABI           string `json:"abi,omitempty" bson:"abi,omitempty"`
	OwnerAddress  string `json:"ownerAddress,omitempty" bson:"ownerAddress,omitempty"`
	TxHash        string `json:"txHash,omitempty" bson:"txHash,omitempty"`
	CreatedAt     int64  `json:"createdAt" bson:"createdAt,omitempty"`
	Type          string `json:"type" bson:"type,omitempty"`
	BalanceString string `json:"balance" bson:"balanceString,omitempty"` // high precise balance for API
	Info          string `json:"info" bson:"info,omitempty"`             // additional info of this address
	Logo          string `json:"logo" bson:"logo,omitempty"`

	// SMC
	IsContract bool `json:"isContract,omitempty" bson:"isContract,omitempty"`

	// Token
	TokenName   string `json:"tokenName,omitempty" bson:"tokenName,omitempty"`
	TokenSymbol string `json:"tokenSymbol,omitempty" bson:"tokenSymbol,omitempty"`
	Decimals    int64  `json:"decimals" bson:"decimals,omitempty"`
	TotalSupply string `json:"totalSupply,omitempty" bson:"totalSupply,omitempty"`

	// Stats
	TxCount         int `json:"txCount,omitempty" bson:"txCount,omitempty"`
	HolderCount     int `json:"holderCount,omitempty" bson:"holderCount,omitempty"`
	InternalTxCount int `json:"internalTxCount,omitempty" bson:"internalTxCount,omitempty"`
	TokenTxCount    int `json:"tokenTxCount,omitempty" bson:"tokenTxCount,omitempty"`

	UpdatedAt int64 `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

type SimpleKRCTokenInfo struct {
	Name        string `json:"name,omitempty" bson:"name"`
	Address     string `json:"address,omitempty" bson:"address"`
	Info        string `json:"info,omitempty" bson:"info"`
	Type        string `json:"type,omitempty" bson:"type"`
	TokenSymbol string `json:"tokenSymbol,omitempty" bson:"tokenSymbol"`
	TotalSupply string `json:"totalSupply,omitempty" bson:"totalSupply"`
	Decimal     int64  `json:"decimal" bson:"decimal"`
	Logo        string `json:"logo,omitempty" bson:"logo"`
	IsVerified  bool   `json:"isVerified" bson:"isVerified"`
}

type InternalTransaction struct {
	*types.Log
	*types.KRCTokenInfo
	From     string `json:"from,omitempty"`
	FromName string `json:"fromName,omitempty"`
	To       string `json:"to,omitempty"`
	ToName   string `json:"toName,omitempty"`
	Value    string `json:"value,omitempty"`
}
