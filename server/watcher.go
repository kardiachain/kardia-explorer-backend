// Package server
package server

import (
	"context"

	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/kardia"
	"github.com/kardiachain/explorer-backend/server/cache"
	"github.com/kardiachain/explorer-backend/server/db"
)

type ValidatorWatcher interface {
	SyncValidators(ctx context.Context) error
}

type watcher struct {
	dbClient    db.Client
	cacheClient cache.Client
	kaiClient   kardia.ClientInterface

	lgr *zap.Logger
}

func NewValidatorWatcher(cfg Config) (ValidatorWatcher, error) {
	return &watcher{}, nil

}

//SyncValidators fetch validators info from network and update to storage and cache
func (s *watcher) SyncValidators(ctx context.Context) error {
	validators, err := s.kaiClient.Validators(ctx)
	if err != nil {
		return err
	}

	if err := s.dbClient.UpsertValidators(ctx, validators.Validators); err != nil {
		return err
	}

	if err := s.cacheClient.UpdateValidators(ctx, validators); err != nil {
		return err
	}
	return nil
}
