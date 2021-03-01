// Package kardia
package kardia

import (
	"crypto/ecdsa"

	"github.com/kardiachain/go-kardia/lib/crypto"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/db"
)

var (
	dbCfg = db.Config{
		DbAdapter: "mgo",
		DbName:    "explorerProd",
		URL:       "mongodb://kardia.ddns.net:27017",
		MinConn:   4,
		MaxConn:   16,
		FlushDB:   false,
	}
	kaiCfg = Config{
		rpcURL:            []string{"https://dev-1.kardiachain.io"},
		trustedNodeRPCURL: []string{"https://dev-1.kardiachain.io"},
	}
)

func SetupTestAccount() (*ecdsa.PublicKey, *ecdsa.PrivateKey, error) {
	privateKey, err := crypto.HexToECDSA("63e16b5334e76d63ee94f35bd2a81c721ebbbb27e81620be6fc1c448c767eed9")
	if err != nil {
		return nil, nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, err
	}

	return publicKeyECDSA, privateKey, nil
}

func SetupNodeClient() (ClientInterface, error) {
	return NewKaiClient(&kaiCfg)
}

func SetupMGOClient() (db.Client, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	dbCfg.Logger = logger
	c, err := db.NewClient(dbCfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}
