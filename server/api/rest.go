// Package api
package api

import (
	"github.com/labstack/echo"
)

// RestServer define all API expose
type RestServer interface {
	IPrivate
	IContract
	IBlock
	ITx
	IAddress
	IKrc721
	IKrc20

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
}
