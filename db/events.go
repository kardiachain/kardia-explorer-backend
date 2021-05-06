// Package db
package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cEvents = "Events"
)

type IEvents interface {
	createEventsCollectionIndexes() []mongo.IndexModel
	InsertEvents(events []types.Log) error
	GetListEvents(ctx context.Context, filter *types.EventsFilter) ([]*types.Log, uint64, error)
	DeleteEmptyEvents(ctx context.Context, contractAddress string) error
}

func (m *mongoDB) createEventsCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"transactionHash": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{Key: "address", Value: 1}, {Key: "timestamp", Value: -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{Key: "methodName", Value: 1}, {Key: "timestamp", Value: -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"blockHeight": -1}, Options: options.Index().SetSparse(true)},
	}
}

func (m *mongoDB) InsertEvents(events []types.Log) error {
	eventsBulkWriter := make([]mongo.WriteModel, len(events))
	for i := range events {
		txModel := mongo.NewInsertOneModel().SetDocument(events[i])
		eventsBulkWriter[i] = txModel
	}
	if len(eventsBulkWriter) > 0 {
		if _, err := m.wrapper.C(cEvents).BulkWrite(eventsBulkWriter); err != nil {
			return err
		}
	}

	return nil
}

func (m *mongoDB) GetListEvents(ctx context.Context, filter *types.EventsFilter) ([]*types.Log, uint64, error) {
	var (
		opts = []*options.FindOptions{
			//options.Find().SetHint(bson.M{"blockHeight": -1}),
			//options.Find().SetHint(bson.M{"transactionHash": 1}),
			//options.Find().SetHint(bson.D{{Key: "address", Value: 1}, {Key: "timestamp", Value: -1}}),
			//options.Find().SetHint(bson.D{{Key: "methodName", Value: 1}, {Key: "timestamp", Value: -1}}),
			options.Find().SetSort(bson.M{"blockHeight": -1}),
		}
		events []*types.Log
		crit   bson.M
	)
	critBytes, err := bson.Marshal(filter)
	if err != nil {
		m.logger.Warn("Cannot marshal events filter criteria", zap.Error(err))
	}
	err = bson.Unmarshal(critBytes, &crit)
	if err != nil {
		m.logger.Warn("Cannot unmarshal events filter criteria", zap.Error(err))
	}
	if filter.Pagination != nil {
		opts = append(opts, options.Find().SetSkip(int64(filter.Pagination.Skip)))
		opts = append(opts, options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cEvents).
		Find(crit, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get contract events: %v", err)
	}
	for cursor.Next(ctx) {
		event := &types.Log{}
		if err := cursor.Decode(event); err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}
	total, err := m.wrapper.C(cEvents).Count(crit)
	if err != nil {
		return nil, 0, err
	}
	return events, uint64(total), nil
}

func (m *mongoDB) DeleteEmptyEvents(ctx context.Context, contractAddress string) error {
	_, err := m.wrapper.C(cEvents).RemoveAll(bson.M{"address": contractAddress, "methodName": ""})
	return err
}
