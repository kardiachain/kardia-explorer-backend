// Package db
package db

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.uber.org/zap"
	"gotest.tools/assert"

	"github.com/kardiachain/explorer-backend/types"
)

var (
	dPool  *dockertest.Pool
	mgoRes *dockertest.Resource
)

func SetupMGO(lgr *zap.Logger) (*mongoDB, error) {
	var err error
	var mgo *mongoDB

	dPool, err = dockertest.NewPool("")
	if err != nil {
		lgr.Fatal("Could not connect to docker: %s", zap.Error(err))
	}

	runOpts := &dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "latest",
	}
	mgoRes, err = dPool.RunWithOptions(runOpts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		lgr.Fatal("Could not start resource: %s", zap.Error(err))
	}

	if err := mgoRes.Expire(60); err != nil {
		return nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := dPool.Retry(func() error {
		url := fmt.Sprintf("mongodb://localhost:%s", mgoRes.GetPort("27017/tcp"))
		cfg := Config{
			URL:     url,
			Logger:  lgr,
			MinConn: 1,
			MaxConn: 4,
			DbName:  "explorer",
		}
		mgo, err = newMongoDB(cfg)
		if err != nil {
			lgr.Fatal("cannot setup", zap.Error(err))
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return mgo, nil

	// When you're done, kill and remove the container

}

func StopMGO() {
	if err := dPool.Purge(mgoRes); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}
}

func TestMGO_Insert(t *testing.T) {
	ctx := context.Background()
	lgr, err := zap.NewDevelopment()
	assert.NilError(t, err)
	mgo, err := SetupMGO(lgr)
	defer StopMGO()
	assert.NilError(t, err)

	blocks, err := mgo.Blocks(ctx, &types.Pagination{Skip: 0, Limit: 10})
	assert.NilError(t, err)
	fmt.Printf("Blocks: %+v \n", blocks)

}
