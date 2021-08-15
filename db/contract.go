// Package db
package db

import (
	"context"
	"time"

	"github.com/kardiachain/kardia-explorer-backend/cfg"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

var (
	cContract = "Contracts"
	cABI      = "ContractABIs"
)

type IContract interface {
	InsertContract(ctx context.Context, contract *types.Contract, addrInfo *types.Address) error
	Contract(ctx context.Context, contractAddr string) (*types.Contract, *types.Address, error)
	UpdateContract(ctx context.Context, contract *types.Contract, addrInfo *types.Address) error
	UpdateKRCTotalSupply(ctx context.Context, krcTokenAddress, totalSupply string) error
	Contracts(ctx context.Context, filter *types.ContractsFilter) ([]*types.Contract, uint64, error)

	CountContracts(ctx context.Context) (int64, error)
	AllContracts(ctx context.Context) ([]*types.Contract, error)
	ContractByType(ctx context.Context, contractType string) ([]*types.Contract, error)
	UpsertSMCABIByType(ctx context.Context, smcType, abi string) error
	SMCABIByType(ctx context.Context, smcType string) (string, error)

	// Remove
	RemoveContract(ctx context.Context, contractAddress string) error
	RemoveContracts(ctx context.Context) error
}

func (m *mongoDB) CountContracts(ctx context.Context) (int64, error) {
	total, err := m.wrapper.C(cContract).Count(bson.M{})
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (m *mongoDB) InsertContract(ctx context.Context, contract *types.Contract, addrInfo *types.Address) error {
	if contract != nil {
		contract.CreatedAt = time.Now().Unix()
		if _, err := m.wrapper.C(cContract).Insert(contract); err != nil {
			return err
		}
	}
	//if addrInfo != nil {
	//	addrInfo.UpdatedAt = time.Now().Unix()
	//	if _, err := m.wrapper.C(cAddresses).Insert(addrInfo); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (m *mongoDB) RemoveContract(ctx context.Context, contractAddress string) error {
	if _, err := m.wrapper.C(cContract).Remove(bson.M{"address": contractAddress}); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) RemoveContracts(ctx context.Context) error {
	if _, err := m.wrapper.C(cContract).Remove(bson.M{"address": ""}); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) AllContracts(ctx context.Context) ([]*types.Contract, error) {
	var contracts []*types.Contract
	var opts []*options.FindOptions
	cursor, err := m.wrapper.C(cContract).Find(bson.M{}, opts...)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

func (m *mongoDB) ContractByType(ctx context.Context, contractType string) ([]*types.Contract, error) {
	var contracts []*types.Contract
	var opts []*options.FindOptions
	cursor, err := m.wrapper.C(cContract).Find(bson.M{"type": contractType}, opts...)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(ctx, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

func (m *mongoDB) Contracts(ctx context.Context, filter *types.ContractsFilter) ([]*types.Contract, uint64, error) {
	var (
		contracts []*types.Contract
		crit      = bson.M{}
	)
	critBytes, err := bson.Marshal(filter)
	if err != nil {
		m.logger.Warn("Cannot marshal contract filter criteria", zap.Error(err))
	}
	err = bson.Unmarshal(critBytes, &crit)
	if err != nil {
		m.logger.Warn("Cannot unmarshal contract filter criteria", zap.Error(err))
	}
	switch filter.Status {
	case "Verified":
		crit["status"] = types.ContractStatusVerified
	case "Unverified":
		crit["status"] = types.ContractStatusUnverified
	default:

	}
	opts := []*options.FindOptions{
		options.Find().SetHint(bson.M{"type": 1}),
		options.Find().SetSort(bson.M{"createdAt": -1}),
	}
	if filter.Pagination != nil {
		filter.Pagination.Sanitize()
		opts = append(opts, options.Find().SetSkip(int64(filter.Pagination.Skip)), options.Find().SetLimit(int64(filter.Pagination.Limit)))
	}
	cursor, err := m.wrapper.C(cContract).Find(crit, opts...)
	if err != nil {
		return nil, 0, err
	}

	if err := cursor.All(ctx, &contracts); err != nil {
		return nil, 0, err
	}

	total, err := m.wrapper.C(cContract).Count(crit)
	if err != nil {
		return nil, 0, err
	}

	return contracts, uint64(total), nil
}

func (m *mongoDB) Contract(ctx context.Context, contractAddr string) (*types.Contract, *types.Address, error) {
	var (
		contract *types.Contract
		addr     *types.Address
	)
	err := m.wrapper.C(cContract).FindOne(bson.M{"address": contractAddr}).Decode(&contract)
	if err != nil {
		return nil, nil, err
	}
	if contract.ABI == "" && contract.Type != "" && contract.Type != cfg.SMCTypeNormal {
		var smcABI *types.ContractABI
		err = m.wrapper.C(cABI).FindOne(bson.M{"type": contract.Type}).Decode(&smcABI)
		if err != nil {
			return nil, nil, err
		}
		contract.ABI = smcABI.ABI
	}
	addr, err = m.AddressByHash(ctx, contractAddr)
	if err != nil {
		return nil, nil, err
	}
	return contract, addr, nil
}

func (m *mongoDB) UpdateContract(ctx context.Context, contract *types.Contract, addrInfo *types.Address) error {
	contract.CreatedAt = time.Now().Unix()
	if _, err := m.wrapper.C(cContract).Upsert(bson.M{"address": contract.Address}, contract); err != nil {
		return err
	}
	if addrInfo != nil {
		addrInfo.UpdatedAt = time.Now().Unix()
		if _, err := m.wrapper.C(cAddresses).Upsert(bson.M{"address": addrInfo.Address}, addrInfo); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDB) UpdateKRCTotalSupply(ctx context.Context, krcTokenAddress, totalSupply string) error {
	addrInfo, err := m.AddressByHash(ctx, krcTokenAddress)
	if err != nil {
		return err
	}
	addrInfo.TotalSupply = totalSupply
	addrInfo.UpdatedAt = time.Now().Unix()
	if _, err := m.wrapper.C(cAddresses).Upsert(bson.M{"address": addrInfo.Address}, addrInfo); err != nil {
		return err
	}
	return nil
}

func (m *mongoDB) UpsertSMCABIByType(ctx context.Context, smcType, abi string) error {
	upsertModel := []mongo.WriteModel{
		mongo.NewUpdateOneModel().SetUpsert(true).SetFilter(bson.M{"type": smcType}).SetUpdate(bson.M{"$set": &types.ContractABI{
			Type: smcType,
			ABI:  abi,
		}}),
	}
	if _, err := m.wrapper.C(cABI).BulkWrite(upsertModel); err != nil {
		m.logger.Warn("cannot upsert new abi", zap.Error(err))
		return err
	}
	return nil
}

func (m *mongoDB) SMCABIByType(ctx context.Context, smcType string) (string, error) {
	var currABI *types.ContractABI
	if err := m.wrapper.C(cABI).FindOne(bson.M{"type": smcType}).Decode(&currABI); err != nil {
		return "", err
	}
	return currABI.ABI, nil
}
