// Package types
package types

import (
	"math/big"

	"github.com/kardiachain/go-kardia/lib/common"
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

type Validators struct {
	TotalValidators            int          `json:"totalValidators" bson:"totalValidators"`
	TotalProposers             int          `json:"totalProposers" bson:"totalProposers"`
	TotalCandidates            int          `json:"totalCandidates" bson:"totalCandidates"`
	TotalDelegators            int          `json:"totalDelegators" bson:"totalDelegators"`
	TotalStakedAmount          string       `json:"totalStakedAmount" bson:"totalStakedAmount"`
	TotalValidatorStakedAmount string       `json:"totalValidatorStakedAmount" bson:"totalValidatorStakedAmount"`
	TotalDelegatorStakedAmount string       `json:"totalDelegatorStakedAmount" bson:"totalDelegatorStakedAmount"`
	Validators                 []*Validator `json:"validators"`
}

type Validator struct {
	Address               common.Address `json:"address" bson:"address"`
	SmcAddress            common.Address `json:"smcAddress" bson:"smcAddress"`
	Status                uint8          `json:"status" bson:"status"`
	Role                  int            `json:"role" bson:"role"`
	Jailed                bool           `json:"jailed" bson:"jailed"`
	Name                  string         `json:"name,omitempty" bson:"name"`
	VotingPowerPercentage string         `json:"votingPowerPercentage" bson:"votingPowerPercentage"`
	StakedAmount          string         `json:"stakedAmount" bson:"stakedAmount"`
	AccumulatedCommission string         `json:"accumulatedCommission" bson:"accumulatedCommission"`
	UpdateTime            uint64         `json:"updateTime" bson:"updateTime"`
	CommissionRate        string         `json:"commissionRate" bson:"commissionRate"`
	TotalDelegators       int            `json:"totalDelegators" bson:"totalDelegators"`
	MaxRate               string         `json:"maxRate" bson:"maxRate"`
	MaxChangeRate         string         `json:"maxChangeRate" bson:"maxChangeRate"`
	SigningInfo           *SigningInfo   `json:"signingInfo" bson:"signingInfo"`
	Delegators            []*Delegator   `json:"delegators,omitempty" bson:"delegators"`
}

type RPCValidator struct {
	Name                  [32]uint8       `json:"name"`
	ValAddr               common.Address  `json:"validatorAddress"`
	ValStakingSmc         common.Address  `json:"valStakingSmc"`
	Tokens                *big.Int        `json:"tokens"`
	Jailed                bool            `json:"jailed"`
	DelegationShares      *big.Int        `json:"delegationShares"`
	AccumulatedCommission *big.Int        `json:"accumulatedCommission"`
	UbdEntryCount         *big.Int        `json:"ubdEntryCount"`
	UpdateTime            *big.Int        `json:"updateTime"`
	Status                uint8           `json:"status"`
	Role                  int             `json:"role"`
	UnbondingTime         *big.Int        `json:"unbondingTime"`
	UnbondingHeight       *big.Int        `json:"unbondingHeight"`
	CommissionRate        *big.Int        `json:"commissionRate,omitempty"`
	MaxRate               *big.Int        `json:"maxRate,omitempty"`
	MaxChangeRate         *big.Int        `json:"maxChangeRate,omitempty"`
	SigningInfo           *SigningInfo    `json:"signingInfo"`
	Delegators            []*RPCDelegator `json:"delegators,omitempty"`
}

type RPCDelegator struct {
	Address      common.Address `json:"address"`
	StakedAmount *big.Int       `json:"stakedAmount"`
	Reward       *big.Int       `json:"reward"`
}

type Delegator struct {
	Address      common.Address `json:"address" bson:"address"`
	Name         string         `json:"name,omitempty" bson:"name"`
	StakedAmount string         `json:"stakedAmount" bson:"stakedAmount"`
	Reward       string         `json:"reward" bson:"reward"`
}

type SlashEvents struct {
	Period   string `json:"period" bson:"period"`
	Fraction string `json:"fraction" bson:"fraction"`
	Height   string `json:"height" bson:"height"`
}

type SigningInfo struct {
	StartHeight        uint64  `json:"startHeight" bson:"startHeight"`
	IndexOffset        uint64  `json:"indexOffset" bson:"indexOffset"`
	Tombstoned         bool    `json:"tombstoned" bson:"tombstoned"`
	MissedBlockCounter uint64  `json:"missedBlockCounter" bson:"missedBlockCounter"`
	IndicatorRate      float64 `json:"indicatorRate" bson:"indicatorRate"`
	JailedUntil        uint64  `json:"jailedUntil" bson:"jailedUntil"`
}

type RPCPeerInfo struct {
	NodeInfo         *NodeInfo `json:"node_info"`
	IsOutbound       bool      `json:"is_outbound"`
	ConnectionStatus struct {
		Duration uint64 `json:"Duration"`
	} `json:"connection_status"`
	RemoteIP string `json:"remote_ip"`
}

type PeerInfo struct {
	Duration uint64 `json:"Duration,omitempty"`
	Moniker  string `json:"moniker,omitempty"` // arbitrary moniker
	RemoteIP string `json:"remote_ip,omitempty"`
}

type ProtocolVersion struct {
	P2P   uint64 `json:"p2p"`
	Block uint64 `json:"block"`
	App   uint64 `json:"app"`
}

type DefaultNodeInfoOther struct {
	TxIndex    string `json:"tx_index"`
	RPCAddress string `json:"rpc_address"`
}

type NodeInfo struct {
	ProtocolVersion ProtocolVersion      `json:"protocol_version"`
	ID              string               `json:"id"`              // authenticated identifier
	ListenAddr      string               `json:"listen_addr"`     // accepting incoming
	Network         string               `json:"network"`         // network/chain ID
	Version         string               `json:"version"`         // major.minor.revision
	Moniker         string               `json:"moniker"`         // arbitrary moniker
	Peers           map[string]*PeerInfo `json:"peers,omitempty"` // peers details
	Other           DefaultNodeInfoOther `json:"other"`           // other application specific data
}

type ValidatorsByDelegator struct {
	Name                  string         `json:"name"`
	Validator             common.Address `json:"validator"`
	ValidatorContractAddr common.Address `json:"validatorContractAddr"`
	ValidatorStatus       uint8          `json:"validatorStatus"`
	ValidatorRole         int            `json:"validatorRole"`
	StakedAmount          string         `json:"stakedAmount"`
	ClaimableRewards      string         `json:"claimableRewards"`
	UnbondedAmount        string         `json:"unbondedAmount"`
	WithdrawableAmount    string         `json:"withdrawableAmount"`
}
