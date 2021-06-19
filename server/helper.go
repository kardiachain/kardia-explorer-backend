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

func filterUniqueAddressesAndContracts(txs []*types.Transaction) ([]*types.Address, []*types.Contract) {
	addrs := make(map[string]*types.Address)
	var contracts []*types.Contract
	for _, tx := range txs {
		addrs[tx.From] = &types.Address{
			Address:    tx.From,
			IsContract: false,
		}
		addrs[tx.To] = &types.Address{
			Address:    tx.To,
			IsContract: false,
		}

		if tx.ContractAddress != "" && tx.ContractAddress != "0x" {
			addrs[tx.ContractAddress] = &types.Address{
				OwnerAddress: tx.From,
				Address:      tx.ContractAddress,
				IsContract:   true,
			}
			c := &types.Contract{
				Address:      tx.ContractAddress,
				OwnerAddress: tx.From,
				TxHash:       tx.Hash,
				ContractType: types.ContractTypeNormal,
				IsVerified:   false,
				CreatedAt:    tx.Time.Unix(),
			}
			contracts = append(contracts, c)
		}
	}
	delete(addrs, "")
	delete(addrs, "0x")
	var addresses []*types.Address

	for _, addr := range addrs {
		addresses = append(addresses, addr)
	}
	return addresses, contracts
}
