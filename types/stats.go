package types

import "time"

type Stats struct {
	UpdatedAt         time.Time `json:"updatedAt" bson:"updatedAt"`
	UpdatedAtBlock    uint64    `json:"updatedAtBlock" bson:"updatedAtBlock"`
	TotalTransactions uint64    `json:"totalTransactions" bson:"totalTransactions"`
	TotalAddresses    uint64    `json:"totalAddresses" bson:"totalAddresses"`
	TotalContracts    uint64    `json:"totalContracts" bson:"totalContracts"`
}
