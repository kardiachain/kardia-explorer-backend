package types

import (
	"time"

	"github.com/kardiachain/go-kardia/lib/common"
)

const (
	defaultLimit = 50
	MaximumLimit = 100
)

type Pagination struct {
	Skip  int `json:"skip"`
	Limit int `json:"limit"`
}

func (f *Pagination) Sanitize() {
	if f.Skip < 0 {
		f.Skip = 0
	}
	if f.Limit <= 0 {
		f.Limit = defaultLimit
	} else if f.Limit > MaximumLimit {
		f.Limit = MaximumLimit
	}
}

type SortFilter struct {
	SortBy string
	Asc    bool
}

type TimeFilter struct {
	FromTime time.Time
	ToTime   time.Time
}

func (f *TimeFilter) Sanitize() {
	if f.FromTime.IsZero() {
		f.FromTime = time.Unix(0, 0)
	}
	if f.ToTime.IsZero() {
		f.ToTime = time.Now()
	}
}

type ContractsFilter struct {
	Type       string      `bson:"type,omitempty"`
	Pagination *Pagination `bson:"-"`

	ContractName string `bson:"name,omitempty"`
	TokenName    string `bson:"tokenName,omitempty"`
	TokenSymbol  string `bson:"tokenSymbol,omitempty"`
	Status       string `bson:"-"`
}

type InternalTxsFilter struct {
	Pagination *Pagination `bson:"-"`

	TransactionHash string          `bson:"txHash,omitempty"`
	Contract        string          `json:"contractAddress" bson:"contractAddress,omitempty"`
	Address         string          `bson:"address,omitempty"`
	Topics          [][]common.Hash `json:"topics" bson:"-"`
}

type TxsFilter struct {
	Pagination *Pagination `json:"pagination" bson:"-"`
	TimeFilter *TimeFilter `bson:"-"`

	ContractAddress string `json:"contractAddress" bson:"contractAddress,omitempty"`
	To              string `json:"to" bson:"to,omitempty"`
}

type BlocksFilter struct {
	Pagination
	TimeFilter
}

func (f *TxsFilter) Sanitize() {
	f.Pagination.Sanitize()
	f.TimeFilter.Sanitize()
}

type KRC20HolderFilter struct {
	Pagination *Pagination `bson:"-"`

	ContractAddress string `bson:"contractAddress,omitempty"`
	HolderAddress   string `bson:"holderAddress,omitempty"`
}

type KRC721HolderFilter struct {
	Pagination *Pagination `bson:"-"`

	ContractAddress string `bson:"contractAddress,omitempty"`
	HolderAddress   string `bson:"holderAddress,omitempty"`
}

type EventsFilter struct {
	Pagination *Pagination `bson:"-"`

	ContractAddress string `bson:"address,omitempty"`
	MethodName      string `bson:"methodName,omitempty"`
	TxHash          string `bson:"transactionHash,omitempty"`
}
