// Package db
package db

import (
	"context"
	"fmt"

	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	RemoveDuplicateEvents(ctx context.Context) ([]*types.Log, error)
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

func (m *mongoDB) RemoveDuplicateEvents(ctx context.Context) ([]*types.Log, error) {
	groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id",
		Value: bson.D{{Key: "address", Value: "$address"},
			{Key: "methodName", Value: "$methodName"},
			{Key: "argumentsName", Value: "$argumentsName"},
			{Key: "arguments", Value: "$arguments"},
			{Key: "topics", Value: "$topics"},
			{Key: "data", Value: "$data"},
			{Key: "blockHeight", Value: "$blockHeight"},
			{Key: "time", Value: "$time"},
			{Key: "transactionHash", Value: "$transactionHash"},
			{Key: "transactionIndex", Value: "$transactionIndex"},
			{Key: "blockHash", Value: "$blockHash"},
			{Key: "logIndex", Value: "$logIndex"},
			{Key: "removed", Value: "$removed"}}},
		{Key: "uniqueIds",
			Value: bson.D{{Key: "$addToSet", Value: "$_id"}}},
		{Key: "count",
			Value: bson.D{{Key: "$sum", Value: 1}}},
	}}}
	matchStage := bson.D{{Key: "$match", Value: bson.D{{Key: "count", Value: bson.D{{Key: "$gt", Value: 1}}}}}}
	opts := []*options.AggregateOptions{
		options.Aggregate().SetAllowDiskUse(true),
	}
	row, err := m.wrapper.C(cEvents).Aggregate(mongo.Pipeline{groupStage, matchStage}, opts...)
	if err != nil {
		return nil, err
	}
	type DataDuplicateResponse struct {
		UniqueIds     []string               `json:"uniqueIds"`
		Count         int64                  `json:"count"`
		Address       string                 `json:"address"`
		MethodName    string                 `json:"methodName"`
		ArgumentsName string                 `json:"argumentsName"`
		Arguments     map[string]interface{} `json:"arguments"`
		Topics        []string               `json:"topics"`
		Data          string                 `json:"data"`
		BlockHeight   uint64                 `json:"blockHeight"`
		Time          time.Time              `json:"time"`
		TxHash        string                 `json:"transactionHash"`
		TxIndex       uint                   `json:"transactionIndex"`
		BlockHash     string                 `json:"blockHash"`
		Index         uint                   `json:"logIndex"`
		Removed       bool                   `json:"removed"`
	}

	var groupIDRowDuplicates []primitive.ObjectID

	var dataDuplicate DataDuplicateResponse
	var events []*types.Log
	for row.Next(ctx) {
		errDecode := row.Decode(&dataDuplicate)
		if errDecode != nil {
			return nil, errDecode
		}
		event := &types.Log{
			Address:       dataDuplicate.Address,
			MethodName:    dataDuplicate.MethodName,
			ArgumentsName: dataDuplicate.ArgumentsName,
			Arguments:     dataDuplicate.Arguments,
			Topics:        dataDuplicate.Topics,
			Data:          dataDuplicate.Data,
			BlockHeight:   dataDuplicate.BlockHeight,
			Time:          dataDuplicate.Time,
			TxHash:        dataDuplicate.TxHash,
			TxIndex:       dataDuplicate.TxIndex,
			BlockHash:     dataDuplicate.BlockHash,
			Index:         dataDuplicate.Index,
			Removed:       dataDuplicate.Removed,
		}
		events = append(events, event)
		for index, e := range dataDuplicate.UniqueIds {
			if index > 0 {
				idRowDuplicate, _ := primitive.ObjectIDFromHex(e)
				groupIDRowDuplicates = append(groupIDRowDuplicates, idRowDuplicate)
			}
		}
	}

	if len(groupIDRowDuplicates) > 0 {
		_, err = m.wrapper.C(cEvents).RemoveAll(bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: groupIDRowDuplicates}}}})
		if err != nil {
			return nil, err
		}
	}

	return events, nil
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
