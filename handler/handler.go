// Package handler
package handler

import (
	"github.com/kardiachain/go-kaiclient/kardia"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
)

type Config struct {
	// Ecosystem node
	Nodes []string
	//Trusted node for long-poll API/action
	TrustedNodes []string

	Logger *zap.Logger
}

type Handler interface {
	IStakingHandler
}

type handler struct {
	ecoNode     kardia.Node
	trustedNode kardia.Node

	// Internal
	db    db.Client
	cache cache.Client

	logger *zap.Logger
}
