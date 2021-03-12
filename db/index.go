// Package db
package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func createDelegatorCollectionIndexes() []mongo.IndexModel {
	return []mongo.IndexModel{
		{Keys: bson.M{"address": 1}, Options: options.Index().SetSparse(true)},
		{Keys: bson.M{"validatorSMCAddress": 1}, Options: options.Index().SetSparse(true)},
	}
}
