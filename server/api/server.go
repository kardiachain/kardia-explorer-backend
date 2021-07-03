// Package api
package api

import (
	kClient "github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/kardia-explorer-backend/cache"
	"github.com/kardiachain/kardia-explorer-backend/db"
	s3 "github.com/kardiachain/kardia-explorer-backend/driver/aws"
	"github.com/kardiachain/kardia-explorer-backend/kardia"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type Server struct {
	HttpRequestSecret string

	node        kClient.Node
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	s3.ConfigUploader
	fileStorage s3.FileStorage

	logger *zap.Logger
}

func (s *Server) Stats(c echo.Context) error {
	panic("implement me")
}

func (s *Server) SetLogger(logger *zap.Logger) *Server {
	s.logger = logger
	return s
}

func (s *Server) SetStorage(db db.Client) *Server {
	s.dbClient = db
	return s
}

func (s *Server) SetCache(cache cache.Client) *Server {
	s.cacheClient = cache
	return s
}

func (s *Server) SetNode(node kClient.Node) *Server {
	s.node = node
	return s
}
