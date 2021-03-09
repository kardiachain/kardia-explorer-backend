package types

import (
	"time"
)

const (
	defaultLimit = 50
	MaximumLimit = 100
)

type Pagination struct {
	Skip  int
	Limit int
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
}

type InternalTxsFilter struct {
	Pagination
	TokenTransactions bool
	InternalAddress   string
}

type TxsFilter struct {
	Pagination
	TimeFilter
}

type BlocksFilter struct {
	Pagination
	TimeFilter
}

func (f *TxsFilter) Sanitize() {
	f.Pagination.Sanitize()
	f.TimeFilter.Sanitize()
}
