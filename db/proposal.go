// Package db
package db

import (
	"context"
	"fmt"

	"github.com/kardiachain/go-kardia/types/time"
	"github.com/kardiachain/kardia-explorer-backend/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type IProposal interface {
	AddVoteToProposal(ctx context.Context, proposalInfo *types.ProposalDetail, voteOption uint64) error
	UpsertProposal(ctx context.Context, proposalInfo *types.ProposalDetail) error
	ProposalInfo(ctx context.Context, proposalID uint64) (*types.ProposalDetail, error)
	GetListProposals(ctx context.Context, pagination *types.Pagination) ([]*types.ProposalDetail, uint64, error)
}

func (m *mongoDB) AddVoteToProposal(ctx context.Context, proposalInfo *types.ProposalDetail, voteOption uint64) error {
	m.logger.Warn("AddVoteToProposal", zap.Any("proposal", proposalInfo))
	currentProposal, _ := m.ProposalInfo(ctx, proposalInfo.ID)
	if currentProposal == nil {
		currentProposal = proposalInfo
	}
	// update number of vote choices
	switch voteOption {
	case 0:
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain + 1
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo
	case 1:
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes + 1
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo
	case 2:
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo + 1
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes
	}
	if err := m.upsertProposal(proposalInfo); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertProposal(ctx context.Context, proposalInfo *types.ProposalDetail) error {
	m.logger.Warn("UpsertProposal", zap.Any("proposal", proposalInfo))
	currentProposal, _ := m.ProposalInfo(ctx, proposalInfo.ID)
	if currentProposal != nil { // need to update these stats from db first
		proposalInfo.NumberOfVoteAbstain = currentProposal.NumberOfVoteAbstain
		proposalInfo.NumberOfVoteYes = currentProposal.NumberOfVoteYes
		proposalInfo.NumberOfVoteNo = currentProposal.NumberOfVoteNo
	}
	if err := m.upsertProposal(proposalInfo); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) ProposalInfo(ctx context.Context, proposalID uint64) (*types.ProposalDetail, error) {
	var result *types.ProposalDetail
	err := m.wrapper.C(cProposal).FindOne(bson.M{"id": proposalID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *mongoDB) upsertProposal(proposalInfo *types.ProposalDetail) error {
	proposalInfo.UpdateTime = time.Now().Unix()
	m.logger.Warn("upsertProposal", zap.Any("proposal", proposalInfo))
	model := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"id": proposalInfo.ID}).SetUpdate(bson.M{"$set": proposalInfo}).SetHint(bson.M{"id": -1}),
	}
	if _, err := m.wrapper.C(cProposal).BulkWrite(model); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) GetListProposals(ctx context.Context, pagination *types.Pagination) ([]*types.ProposalDetail, uint64, error) {
	var (
		opts      []*options.FindOptions
		proposals []*types.ProposalDetail
	)
	if pagination != nil {
		opts = []*options.FindOptions{
			options.Find().SetHint(bson.M{"id": -1}),
			options.Find().SetSort(bson.M{"id": 1}),
			options.Find().SetSkip(int64(pagination.Skip)),
			options.Find().SetLimit(int64(pagination.Limit)),
		}
	}
	cursor, err := m.wrapper.C(cProposal).
		Find(bson.M{}, opts...)
	defer func() {
		err = cursor.Close(ctx)
		if err != nil {
			m.logger.Warn("Error when close cursor", zap.Error(err))
		}
	}()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get list proposals: %v", err)
	}
	for cursor.Next(ctx) {
		proposal := &types.ProposalDetail{}
		if err := cursor.Decode(proposal); err != nil {
			return nil, 0, err
		}
		proposals = append(proposals, proposal)
	}
	// get total transaction in block in database
	total, err := m.wrapper.C(cProposal).Count(bson.M{})
	if err != nil {
		return nil, 0, err
	}
	return proposals, uint64(total), nil
}
