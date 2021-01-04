// Package db
package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/kardiachain/explorer-backend/types"
)

const (
	cBlocks = "Blocks"
)

type FindBlockFilter struct {
	StartDate string
	EndDate   string
}

func (m *mongoDB) FindBlock(ctx context.Context, filter FindBlockFilter) ([]*types.Block, error) {
	const (
		layoutISO = "2006-01-02"
	)
	startDate, err := time.Parse(layoutISO, filter.StartDate)
	if err != nil {
		m.logger.Debug("cannot parse", zap.Error(err))
		return nil, err
	}
	endDate, err := time.Parse(layoutISO, filter.EndDate)
	if err != nil {
		m.logger.Debug("cannot parse", zap.Error(err))
		return nil, err
	}

	fmt.Println("StartDate", startDate)
	fmt.Println("EndDate", endDate)
	startDateCond := bson.M{"time": bson.M{"$gte": startDate}}
	endDateCond := bson.M{"time": bson.M{"$lt": endDate}}

	cond := bson.M{"$and": []bson.M{startDateCond, endDateCond}}
	cursor, err := m.wrapper.C(cBlocks).Find(cond)
	if err != nil {
		m.logger.Info("err", zap.Error(err))
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			return
		}
	}()
	var blocks []*types.Block
	for cursor.Next(ctx) {
		var b types.Block
		if err := cursor.Decode(&b); err != nil {
			m.logger.Info("cannot decode block info", zap.Error(err))
			return nil, err
		}

		blocks = append(blocks, &b)
	}

	return blocks, nil
}
