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
	InsertEvents(event *types.FunctionCall) error
	GetListEvents(ctx context.Context, pagination *types.Pagination, contractAddress string, methodName string) ([]*types.FunctionCall, uint64, error)
}

func (m *mongoDB) createEventsCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"txHash": 1}, Options: options.Index().SetUnique(true).SetSparse(true)},
		{Keys: bson.D{{Key: "contractAddress", Value: 1}, {Key: "timestamp", Value: -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.D{{Key: "methodName", Value: 1}, {Key: "timestamp", Value: -1}}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"timestamp": -1}, Options: options.Index().SetSparse(true)},
	}
}

func (m *mongoDB) InsertEvents(event *types.FunctionCall) error {
	if _, err := m.wrapper.C(cEvents).Insert(event); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) GetListEvents(ctx context.Context, pagination *types.Pagination, contractAddress string, methodName string) ([]*types.FunctionCall, uint64, error) {
	var (
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.M{"timestamp": -1}),
			options.Find().SetHint(bson.M{"txHash": 1}),
			options.Find().SetHint(bson.D{{Key: "contractAddress", Value: 1}, {Key: "timestamp", Value: -1}}),
			options.Find().SetHint(bson.D{{Key: "methodName", Value: 1}, {Key: "timestamp", Value: -1}}),
			options.Find().SetSort(bson.M{"timestamp": -1}),
		}
		events []*types.FunctionCall
	)
	if pagination != nil {
		opts = append(opts, options.Find().SetSkip(int64(pagination.Skip)))
		opts = append(opts, options.Find().SetLimit(int64(pagination.Limit)))
	}
	var filter []bson.M
	if contractAddress != "" {
		filter = append(filter, bson.M{"contractAddress": contractAddress})
	}
	if methodName != "" {
		filter = append(filter, bson.M{"methodName": methodName})
	}
	cursor, err := m.wrapper.C(cEvents).
		Find(bson.M{"$and": filter}, opts...)
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
		event := &types.FunctionCall{}
		if err := cursor.Decode(event); err != nil {
			return nil, 0, err
		}
		events = append(events, event)
	}
	total, err := m.wrapper.C(cEvents).Count(bson.M{"$and": filter})
	if err != nil {
		return nil, 0, err
	}
	return events, uint64(total), nil
}
