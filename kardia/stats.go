// Package kardia
package kardia

import (
	"context"

	"github.com/kardiachain/go-kardia/lib/common"

	"github.com/kardiachain/explorer-backend/types"
)

func (ec *Client) KardiaCall(ctx context.Context, args types.CallArgsJSON) (common.Bytes, error) {
	var result common.Bytes
	err := ec.chooseClient().c.CallContext(ctx, &result, "kai_kardiaCall", args, "latest")
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ec *Client) NodesInfo(ctx context.Context) ([]*types.NodeInfo, error) {
	var (
		nodes []*types.NodeInfo
		err   error
	)
	clientList := append(ec.clientList, ec.trustedClientList...)
	nodeMap := make(map[string]*types.NodeInfo, len(clientList)) // list all nodes in network
	peersMap := make(map[string]*types.RPCPeerInfo)              // list all peers details
	for _, client := range clientList {
		var (
			node  *types.NodeInfo
			peers []*types.RPCPeerInfo
		)
		// get current node info then get it's peers
		err = client.c.CallContext(ctx, &node, "node_nodeInfo")
		if err != nil {
			continue
		}
		err := client.c.CallContext(ctx, &peers, "node_peers")
		if err != nil {
			continue
		}
		node.Peers = make(map[string]*types.PeerInfo)
		for _, peer := range peers {
			// append current peer to this node
			node.Peers[peer.NodeInfo.ID] = &types.PeerInfo{
				Moniker:  peer.NodeInfo.Moniker,
				Duration: peer.ConnectionStatus.Duration,
				RemoteIP: peer.RemoteIP,
			}
			peersMap[peer.NodeInfo.ID] = peer
			// try to discover new nodes through these peers
			// init a new node info in list
			if nodeMap[peer.NodeInfo.ID] == nil {
				nodeMap[peer.NodeInfo.ID] = peer.NodeInfo
			}
			if nodeMap[peer.NodeInfo.ID].Peers == nil {
				nodeMap[peer.NodeInfo.ID].Peers = make(map[string]*types.PeerInfo)
			}
			nodeMap[peer.NodeInfo.ID].Peers[node.ID] = &types.PeerInfo{} // mark this peer for re-updating later
		}
		nodeMap[node.ID] = node
	}
	for _, node := range nodeMap {
		// re-update full peers info
		for id, peer := range node.Peers {
			node.Peers[id] = &types.PeerInfo{
				Duration: peer.Duration,
				Moniker:  peer.Moniker,
				RemoteIP: peer.RemoteIP,
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (ec *Client) Datadir(ctx context.Context) (string, error) {
	var result string
	err := ec.chooseClient().c.CallContext(ctx, &result, "node_datadir")
	return result, err
}
