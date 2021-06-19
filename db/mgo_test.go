// Package db
package db

import (
	"go.uber.org/zap"
)

func GetMgo() (*mongoDB, error) {
	lgr, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	mgoCfg := Config{
		DbAdapter: "mgo",
		DbName:    "explorer",
		URL:       "mongodb://10.10.0.43:27019",
		MinConn:   1,
		MaxConn:   4,
		FlushDB:   false,
		Logger:    lgr,
	}

	mgo, err := newMongoDB(mgoCfg)
	if err != nil {
		return nil, err
	}
	return mgo, nil
}
