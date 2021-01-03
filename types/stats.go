package types

import (
	"time"
)

type StakingStats struct {
	TotalValidators            int    `json:"totalValidators" bson:"totalValidators"`
	TotalProposers             int    `json:"totalProposers" bson:"totalProposers"`
	TotalCandidates            int    `json:"totalCandidates" bson:"totalCandidates"`
	TotalDelegators            int    `json:"totalDelegators" bson:"totalDelegators"`
	TotalStakedAmount          string `json:"totalStakedAmount" bson:"totalStakedAmount"`
	TotalValidatorStakedAmount string `json:"totalValidatorStakedAmount" bson:"totalValidatorStakedAmount"`
	TotalDelegatorStakedAmount string `json:"totalDelegatorStakedAmount" bson:"totalDelegatorStakedAmount"`
}

type Stats struct {
	UpdatedAt         time.Time `json:"updatedAt" bson:"updatedAt"`
	UpdatedAtBlock    uint64    `json:"updatedAtBlock" bson:"updatedAtBlock"`
	TotalTransactions uint64    `json:"totalTransactions" bson:"totalTransactions"`
	TotalAddresses    uint64    `json:"totalAddresses" bson:"totalAddresses"`
	TotalContracts    uint64    `json:"totalContracts" bson:"totalContracts"`
}
