// Package types
package types

type Validator struct {
	Address     string          `json:"address"`
	VotingPower float64         `json:"votingPower"`
	Name        string          `json:"name"`
	Protocols   ProtocolVersion `json:"protocol"`
	PeerCount   int             `json:"peerCount"`
	RpcUrl      string          `json:"rpcUrl"`
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
