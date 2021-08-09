// Package kardia
package kardia

import (
	"go.uber.org/zap"
)

func SetupTestClient() (ClientInterface, error) {
	lgr, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	cfg := Config{
		rpcURL:            []string{"https://kai-internal-1.kardiachain.io"},
		trustedNodeRPCURL: []string{"https://kai-internal-1.kardiachain.io"},
		lgr:               lgr,
	}
	node, err := NewKaiClient(&cfg)
	if err != nil {
		return nil, err
	}

	return node, nil
}
