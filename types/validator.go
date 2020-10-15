// Package types
package types

type Validator struct {
	Address     string  `json:"address"`
	VotingPower float64 `json:"votingPower"`
}

type PeerInfo struct {
	ENR     string   `json:"enr,omitempty"` // Ethereum Node Record
	Enode   string   `json:"enode"`         // Node URL
	ID      string   `json:"id"`            // Unique node identifier
	Name    string   `json:"name"`          // Name of the node, including client type, version, OS, custom data
	Caps    []string `json:"caps"`          // Protocols advertised by this peer
	Address string   `json:"address"`       // Coinbase address
	Network struct {
		LocalAddress  string `json:"localAddress"`  // Local endpoint of the TCP data connection
		RemoteAddress string `json:"remoteAddress"` // Remote endpoint of the TCP data connection
		Inbound       bool   `json:"inbound"`
		Trusted       bool   `json:"trusted"`
		Static        bool   `json:"static"`
	} `json:"network"`
	Protocols map[string]interface{} `json:"protocols"` // Sub-protocol specific metadata fields
}

type NodeInfo struct {
	ID        string `json:"id"`        // Unique node identifier (also the encryption key)
	Name      string `json:"name"`      // Name of the node, including client type, version, OS, custom data
	Enode     string `json:"enode"`     // Enode URL for adding this peer from remote peers
	ENR       string `json:"enr"`       // Ethereum Node Record
	IP        string `json:"ip"`        // IP address of the node
	Address   string `json:"address"`   // Coinbase address
	PeerCount int    `json:"peerCount"` // Number of other peers connecting to this peer
	Ports     struct {
		Discovery int `json:"discovery"` // UDP listening port for discovery protocol
		Listener  int `json:"listener"`  // TCP listening port for RLPx
	} `json:"ports"`
	ListenAddr string                 `json:"listenAddr"`
	Protocols  map[string]interface{} `json:"protocols"`
}
