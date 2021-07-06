// Package api
package api

import (
	"context"
	"math/big"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func (s *Server) Ping(c echo.Context) error {
	type pingStat struct {
		Version string `json:"version"`
	}
	stats := &pingStat{Version: cfg.ServerVersion}
	return OK.SetData(stats).Build(c)
}

func (s *Server) ServerStatus(c echo.Context) error {
	lgr := s.logger
	ctx := context.Background()
	//var status *types.ServerStatus
	//var err error
	status, err := s.cacheClient.ServerStatus(ctx)
	if err != nil {
		lgr.Error("cannot get cache, return default instead")
		status = &types.ServerStatus{
			Status:        "ONLINE",
			AppVersion:    "1.0.0",
			ServerVersion: "1.0.0",
			DexStatus:     "ONLINE",
		}
		// Return default
	}
	return OK.SetData(status).Build(c)
}

func (s *Server) UpdateServerStatus(c echo.Context) error {
	lgr := s.logger.With(zap.String("method", "UpdateServerStatus"))
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		lgr.Warn("Cannot authorization request")
		return Unauthorized.Build(c)
	}
	var serverStatus *types.ServerStatus
	if err := c.Bind(&serverStatus); err != nil {
		lgr.Error("cannot bind server status", zap.Error(err))
		return Invalid.Build(c)
	}
	ctx := context.Background()
	if err := s.cacheClient.UpdateServerStatus(ctx, serverStatus); err != nil {
		lgr.Error("cannot update server status", zap.Error(err))
		return Invalid.Build(c)
	}

	return OK.SetData(nil).Build(c)

}

func (s *Server) Nodes(c echo.Context) error {
	ctx := context.Background()
	nodes, err := s.kaiClient.NodesInfo(ctx)
	if err != nil {
		s.logger.Warn("cannot get nodes info from RPC", zap.Error(err))
		return Invalid.Build(c)
	}
	var result []*NodeInfo
	for _, node := range nodes {
		result = append(result, &NodeInfo{
			ID:         node.ID,
			Moniker:    node.Moniker,
			PeersCount: len(node.Peers),
		})
	}
	customNodes, err := s.dbClient.Nodes(ctx)
	if err != nil {
		// If cannot read nodes from db
		// then return network nodes only
		return OK.SetData(result).Build(c)
	}

	for _, n := range customNodes {
		result = append(result, &NodeInfo{
			ID:         n.ID,
			Moniker:    n.Moniker,
			PeersCount: len(n.Peers),
		})
	}

	return OK.SetData(result).Build(c)
}

func (s *Server) TokenInfo(c echo.Context) error {
	ctx := context.Background()
	if !s.cacheClient.IsRequestToCoinMarket(ctx) {
		tokenInfo, err := s.cacheClient.TokenInfo(ctx)
		if err != nil {
			tokenInfo, err = s.fetchTokenInfo(ctx)
			if err != nil {
				return Invalid.Build(c)
			}
		}
		cirSup, err := s.kaiClient.GetCirculatingSupply(ctx)
		if err != nil {
			return Invalid.Build(c)
		}
		cirSup = new(big.Int).Div(cirSup, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		tokenInfo.MainnetCirculatingSupply = cirSup.Int64() - 4500000000
		return OK.SetData(tokenInfo).Build(c)
	}

	tokenInfo, err := s.fetchTokenInfo(ctx)
	if err != nil {
		return Invalid.Build(c)
	}
	cirSup, err := s.kaiClient.GetCirculatingSupply(ctx)
	if err != nil {
		return Invalid.Build(c)
	}
	cirSup = new(big.Int).Div(cirSup, new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	tokenInfo.MainnetCirculatingSupply = cirSup.Int64() - 4500000000
	return OK.SetData(tokenInfo).Build(c)
}

func (s *Server) UpdateSupplyAmounts(c echo.Context) error {
	ctx := context.Background()
	if c.Request().Header.Get("Authorization") != s.authorizationSecret {
		return Unauthorized.Build(c)
	}
	var supplyInfo *types.SupplyInfo
	if err := c.Bind(&supplyInfo); err != nil {
		return Invalid.Build(c)
	}
	if err := s.cacheClient.UpdateSupplyAmounts(ctx, supplyInfo); err != nil {
		return Invalid.Build(c)
	}
	return OK.SetData(nil).Build(c)
}
