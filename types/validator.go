// Package types
package types

import (
	"github.com/kardiachain/go-kardiamain/lib/common"
)

type Validator struct {
	Address         common.Address `json:"address"`
	VotingPower     int64          `json:"votingPower"`
	StakedAmount    string         `json:"stakedAmount"`
	Commission      string         `json:"commission"`
	CommissionRate  string         `json:"commissionRate"`
	TotalDelegators int            `json:"totalDelegators"`
	MaxRate         string         `json:"maxRate"`
	MaxChangeRate   string         `json:"maxChangeRate"`
	Delegators      []*Delegator   `json:"delegators,omitempty"`
}

type Delegator struct {
	Address      common.Address `json:"address"`
	StakedAmount string         `json:"stakedAmount"`
	Reward       string         `json:"reward"`
}

type PeerInfo struct {
	NodesInfo []*NodeInfo `json:"node_info"`
	// IsOutbound bool     `json:"is_outbound"`
	// RemoteIP   string   `json:"remote_ip"`
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
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	ID              string          `json:"id"`          // authenticated identifier
	ListenAddr      string          `json:"listen_addr"` // accepting incoming
	// Network         string               `json:"network"`     // network/chain ID
	// Version         string               `json:"version"`     // major.minor.revision
	// Channels        []byte               `json:"channels"`    // channels this node knows about
	Moniker string               `json:"moniker"` // arbitrary moniker
	Other   DefaultNodeInfoOther `json:"other"`   // other application specific data
}
