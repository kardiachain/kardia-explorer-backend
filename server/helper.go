// Package server
package server

import (
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func filterNewContractAddresses(txs []*types.Transaction) []string {
	var contractAddresses []string
	for _, tx := range txs {
		if tx.ContractAddress != "" && tx.ContractAddress != "0x" {
			contractAddresses = append(contractAddresses, tx.ContractAddress)
		}
	}
	return contractAddresses
}
