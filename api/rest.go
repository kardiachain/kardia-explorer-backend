// Package api
package api

import (
	"github.com/labstack/echo"
)

// EchoServer define all API expose
type EchoServer interface {
	// General
	Ping(c echo.Context) error
	ServerStatus(c echo.Context) error
	UpdateServerStatus(c echo.Context) error
	Stats(c echo.Context) error
	TotalHolders(c echo.Context) error
	TokenInfo(c echo.Context) error
	Nodes(c echo.Context) error

	// Staking-related
	StakingStats(c echo.Context) error
	Validator(c echo.Context) error
	ValidatorsByDelegator(c echo.Context) error
	Validators(c echo.Context) error
	Candidates(c echo.Context) error
	MobileValidators(c echo.Context) error
	MobileCandidates(c echo.Context) error

	// Proposal
	GetProposalsList(c echo.Context) error
	GetProposalDetails(c echo.Context) error
	GetParams(c echo.Context) error

	// Blocks
	Blocks(c echo.Context) error
	Block(c echo.Context) error
	BlockTxs(c echo.Context) error
	BlocksByProposer(c echo.Context) error
	PersistentErrorBlocks(c echo.Context) error

	// Addresses
	Addresses(c echo.Context) error
	AddressInfo(c echo.Context) error
	AddressTxs(c echo.Context) error
	AddressHolders(c echo.Context) error

	// Tx
	Txs(c echo.Context) error
	TxByHash(c echo.Context) error

	// Admin sector
	ReloadAddressesBalance(c echo.Context) error
	ReloadValidators(c echo.Context) error
	UpdateAddressName(c echo.Context) error
	UpsertNetworkNodes(c echo.Context) error
	RemoveNetworkNodes(c echo.Context) error
	UpdateSupplyAmounts(c echo.Context) error

	RemoveDuplicateEvents(c echo.Context) error

	IContract

	SearchAddressByName(c echo.Context) error

	GetHoldersListByToken(c echo.Context) error
	GetInternalTxs(c echo.Context) error
	UpdateInternalTxs(c echo.Context) error
}

type IContract interface {
	Contracts(c echo.Context) error
	Contract(c echo.Context) error
	InsertContract(c echo.Context) error
	UpdateContract(c echo.Context) error
	UpdateSMCABIByType(c echo.Context) error

	ContractEvents(c echo.Context) error
}
