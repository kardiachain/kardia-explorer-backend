// Package server
package server

import (
	"context"
	"math/big"
	"strings"

	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"github.com/kardiachain/kardia-explorer-backend/types"
)

func (s *infoServer) FilterProposalEvent(ctx context.Context, txs []*types.Transaction) error {
	lgr := s.logger.With(zap.String("method", "filterProposalEvent"))

	for _, tx := range txs {
		decoded := tx.DecodedInputData
		if !strings.EqualFold(tx.To, cfg.ParamsContractAddr) {
			continue
		}

		if tx.Status != 1 {
			continue
		}

		if tx.DecodedInputData == nil {
			continue
		}
		if decoded.MethodName != "addVote" && decoded.MethodName != "confirmProposal" {
			lgr.Debug("new proposal event, but skipped", zap.Any("Decoded", decoded))
			return nil
		}
		// get proposal info
		proposalID, ok := new(big.Int).SetString(decoded.Arguments["proposalId"].(string), 10)
		if !ok {
			lgr.Debug("Cannot set proposalID")
		}
		proposalDetail := &types.ProposalDetail{}
		proposal, err := s.dbClient.ProposalInfo(ctx, proposalID.Uint64())
		if err == nil {
			proposalDetail = proposal
		}
		rpcProposal, err := s.kaiClient.GetProposalDetails(ctx, proposalID)
		if err != nil {
			s.logger.Warn("cannot get proposal by ID from RPC", zap.Any("proposal", proposalID), zap.Error(err))
		}

		proposalDetail.VoteYes = rpcProposal.VoteYes
		proposalDetail.VoteNo = rpcProposal.VoteNo
		proposalDetail.VoteAbstain = rpcProposal.VoteAbstain

		// insert to db
		if decoded.MethodName == "addVote" {
			voteOption := new(big.Int).SetInt64(int64(decoded.Arguments["option"].(uint8)))
			err = s.dbClient.AddVoteToProposal(ctx, proposal, voteOption.Uint64())
			if err != nil {
				s.logger.Warn("cannot add vote to new proposal in db", zap.Any("decoded", decoded), zap.Error(err))
			}
		} else if decoded.MethodName == "confirmProposal" {
			err = s.dbClient.UpsertProposal(ctx, proposal)
			if err != nil {
				s.logger.Warn("cannot confirm proposal in db", zap.Any("decoded", decoded), zap.Error(err))
			}
		}
	}

	return nil
}
