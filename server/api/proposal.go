// Package api
package api

import (
	"context"
	"fmt"
	"math/big"

	"github.com/labstack/echo"
	"go.uber.org/zap"
)

func (s *Server) GetProposalsList(c echo.Context) error {
	ctx := context.Background()
	pagination, page, limit := getPagingOption(c)
	dbResult, dbTotal, dbErr := s.dbClient.GetListProposals(ctx, pagination)
	if dbErr != nil {
		return Invalid.Build(c)
	}
	rpcResult, rpcTotal, rpcErr := s.kaiClient.GetProposals(ctx, pagination)
	if rpcErr != nil {
		fmt.Println("GetProposals err: ", rpcErr)
		return Invalid.Build(c)
	}
	if dbTotal != rpcTotal { // try to find out and insert missing proposals to db
		isFound := false
		for _, rpcProposal := range rpcResult {
			isFound = false
			for _, dbProposal := range dbResult {
				if dbProposal.ID == rpcProposal.ID {
					isFound = true
					break
				}
			}
			if isFound {
				continue
			}
			dbResult = append(dbResult, rpcProposal) // include new proposal in response
			s.logger.Info("Inserting new proposal", zap.Any("proposal", rpcProposal))
			err := s.dbClient.UpsertProposal(ctx, rpcProposal) // insert missing proposal to db
			if err != nil {
				s.logger.Debug("Cannot insert new proposal to DB", zap.Error(err))
			}
		}
	}
	return OK.SetData(PagingResponse{
		Page:  page,
		Limit: limit,
		Data:  dbResult,
		Total: rpcTotal,
	}).Build(c)
}

func (s *Server) GetProposalDetails(c echo.Context) error {
	ctx := context.Background()
	proposalID, ok := new(big.Int).SetString(c.Param("id"), 10)
	if !ok {
		return Invalid.Build(c)
	}
	// Force get details from chain
	//result, err := s.dbClient.ProposalInfo(ctx, proposalID.Uint64())
	//if err == nil {
	//	return OK.SetData(result).Build(c)
	//}
	result, err := s.kaiClient.GetProposalDetails(ctx, proposalID)
	if err != nil {
		fmt.Println("GetProposalDetails err: ", err)
		return Invalid.Build(c)
	}
	s.logger.Info("Proposal details", zap.Any("Details", result))
	return OK.SetData(result).Build(c)
}

func (s *Server) GetParams(c echo.Context) error {
	ctx := context.Background()
	params, err := s.kaiClient.GetParams(ctx)
	if err != nil {
		return Invalid.Build(c)
	}
	result := make(map[string]interface{})
	for _, param := range params {
		result[param.LabelName] = param.FromValue
	}
	return OK.SetData(result).Build(c)
}
