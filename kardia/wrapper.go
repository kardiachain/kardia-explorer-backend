// Package kardia
package kardia

import (
	"context"
	"fmt"
	"sync"

	"github.com/kardiachain/go-kaiclient/kardia"
	"github.com/kardiachain/go-kardia/lib/common"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

type WrapperConfig struct {
	TrustedNodes []string
	PublicNodes  []string
	WSNodes      []string
	Logger       *zap.Logger
}

type Wrapper struct {
	trustedNodes []kardia.Node
	publicNodes  []kardia.Node
	wsNodes      []kardia.Node
	logger       *zap.Logger
}

const (
	WorkerRate = 4
)

func NewWrapper(cfg WrapperConfig) (*Wrapper, error) {
	w := &Wrapper{
		logger: cfg.Logger,
	}
	lgr := w.logger
	for _, url := range cfg.TrustedNodes {
		lgr.Info("Setup trusted node:", zap.String("url", url))
		node, err := kardia.NewNode(url, cfg.Logger)
		if err != nil {
			return nil, err
		}
		w.trustedNodes = append(w.trustedNodes, node)
	}
	for _, url := range cfg.PublicNodes {
		lgr.Info("Setup public node:", zap.String("url", url))
		node, err := kardia.NewNode(url, cfg.Logger)
		if err != nil {
			return nil, err
		}
		w.publicNodes = append(w.publicNodes, node)
	}
	for _, url := range cfg.WSNodes {
		lgr.Info("Setup ws node:", zap.String("url", url))
		node, err := kardia.NewNode(url, cfg.Logger)
		if err != nil {
			return nil, err
		}
		w.wsNodes = append(w.wsNodes, node)
	}
	return w, nil
}

func (w *Wrapper) WSNode() kardia.Node {
	return w.wsNodes[0]
}

func (w *Wrapper) TrustedNode() kardia.Node {
	return w.trustedNodes[0]
}

func (w *Wrapper) PublicNode() kardia.Node {
	return w.publicNodes[0]
}

func (w *Wrapper) pickTrusted() kardia.Node {
	return w.trustedNodes[0]
}

func Worker() {

}

func (w *Wrapper) ValidatorsWithWorker(ctx context.Context) ([]*types.Validator, error) {
	lgr := w.logger.With(zap.String("method", "ValidatorsWithWorker"))
	validatorSMCAddresses, err := w.pickTrusted().ValidatorSMCAddresses(ctx)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	nodesSize := len(w.publicNodes)
	lgr.Info("Start load validators with node size", zap.Int("Size", nodesSize))
	type inputArgs struct {
		node    kardia.Node
		smcAddr string
	}

	var validators []*types.Validator
	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	p, _ := ants.NewPoolWithFunc(nodesSize*WorkerRate, func(i interface{}) {
		defer wg.Done()
		input := i.(inputArgs)
		v, err := w.validatorWithNode(ctx, input.smcAddr, input.node)
		if err != nil {
			return
		}
		validators = append(validators, v)

	})
	defer p.Release()
	// Submit tasks one by one.
	for id, smcAddr := range validatorSMCAddresses {
		wg.Add(1)
		nodeId := id % nodesSize
		node := w.publicNodes[nodeId]
		input := inputArgs{
			node:    node,
			smcAddr: smcAddr.Hex(),
		}
		_ = p.Invoke(input)
	}
	wg.Wait()
	return validators, nil
}

func getValidatorRole(proposers []common.Address, validator *types.Validator) int {
	for _, val := range proposers {
		if val.Hex() == validator.Address {
			return 2
		}
	}
	// else if his node is started, he is a normal validator
	if validator.Status == 2 {
		return 1
	}
	// otherwise he is a candidate
	return 0
}

func (w *Wrapper) Validators(ctx context.Context) ([]*types.Validator, error) {
	validatorSMCAddresses, err := w.pickTrusted().ValidatorSMCAddresses(ctx)
	if err != nil {
		return nil, err
	}
	var validators []*types.Validator
	for _, smcAddr := range validatorSMCAddresses {
		v, err := w.Validator(ctx, smcAddr.Hex())
		if err != nil {
			return nil, err
		}
		validators = append(validators, v)
	}

	return validators, nil
}

func (w *Wrapper) Validator(ctx context.Context, validatorSMCAddress string) (*types.Validator, error) {
	lgr := w.logger.With(zap.String("method", "Validator"))
	proposers, err := w.TrustedNode().ValidatorSets(ctx)
	if err != nil {
		lgr.Error("cannot load proposer set", zap.Error(err))
		return nil, err
	}

	nValidator, err := w.pickTrusted().ValidatorInfo(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get validator info", zap.Error(err))
		return nil, err
	}

	commission, err := w.pickTrusted().ValidatorCommission(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get validator commission", zap.Error(err))
		return nil, err
	}

	delegatorAddresses, err := w.pickTrusted().DelegatorAddresses(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get delegators addresses", zap.Error(err))
		return nil, err
	}

	signingInfo, err := w.pickTrusted().SigningInfo(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get signing info", zap.Error(err))
		return nil, err
	}

	commissionRate, err := convertBigIntToPercentage(commission.Rate.String())
	if err != nil {
		return nil, err
	}

	maxRate, err := convertBigIntToPercentage(commission.MaxRate.String())
	if err != nil {
		return nil, err
	}

	maxChangeRate, err := convertBigIntToPercentage(commission.MaxChangeRate.String())
	if err != nil {
		return nil, err
	}

	v := &types.Validator{
		Address:               nValidator.Signer.String(),
		SmcAddress:            validatorSMCAddress,
		Status:                nValidator.Status,
		Role:                  int(nValidator.Status),
		Jailed:                nValidator.Jailed,
		Name:                  validatorNameInString(nValidator.Name),
		StakedAmount:          nValidator.Tokens.String(),
		AccumulatedCommission: nValidator.AccumulatedCommission.String(),
		CommissionRate:        commissionRate,
		TotalDelegators:       len(delegatorAddresses),
		MaxRate:               maxRate,
		MaxChangeRate:         maxChangeRate,
		SigningInfo: &types.SigningInfo{
			StartHeight:        signingInfo.StartHeight.Uint64(),
			IndexOffset:        signingInfo.IndexOffset.Uint64(),
			Tombstoned:         signingInfo.Tombstoned,
			MissedBlockCounter: signingInfo.MissedBlockCounter.Uint64(),
			IndicatorRate:      100 - float64(signingInfo.MissedBlockCounter.Uint64())/100,
			JailedUntil:        signingInfo.JailedUntil.Uint64(),
		},
	}
	v.Role = getValidatorRole(proposers, v)
	return v, nil
}

func (w *Wrapper) validatorWithNode(ctx context.Context, validatorSMCAddress string, node kardia.Node) (*types.Validator, error) {
	lgr := w.logger.With(zap.String("method", "validatorWithNode"))

	proposers, err := w.TrustedNode().ValidatorSets(ctx)
	if err != nil {
		lgr.Error("cannot load proposer set", zap.Error(err))
		return nil, err
	}

	nValidator, err := node.ValidatorInfo(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get validator info", zap.Error(err))
		return nil, err
	}

	commission, err := node.ValidatorCommission(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get validator commission", zap.Error(err))
		return nil, err
	}

	delegatorAddresses, err := node.DelegatorAddresses(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get delegator address", zap.Error(err))
		return nil, err
	}

	signingInfo, err := node.SigningInfo(ctx, validatorSMCAddress)
	if err != nil {
		lgr.Error("cannot get signing info", zap.Error(err))
		return nil, err
	}

	commissionRate, err := convertBigIntToPercentage(commission.Rate.String())
	if err != nil {
		return nil, err
	}

	maxRate, err := convertBigIntToPercentage(commission.MaxRate.String())
	if err != nil {
		return nil, err
	}

	maxChangeRate, err := convertBigIntToPercentage(commission.MaxChangeRate.String())
	if err != nil {
		return nil, err
	}

	v := &types.Validator{
		Address:               nValidator.Signer.String(),
		SmcAddress:            validatorSMCAddress,
		Status:                nValidator.Status,
		Jailed:                nValidator.Jailed,
		Name:                  validatorNameInString(nValidator.Name),
		StakedAmount:          nValidator.Tokens.String(),
		AccumulatedCommission: nValidator.AccumulatedCommission.String(),
		CommissionRate:        commissionRate,
		TotalDelegators:       len(delegatorAddresses),
		MaxRate:               maxRate,
		MaxChangeRate:         maxChangeRate,
		SigningInfo: &types.SigningInfo{
			StartHeight:        signingInfo.StartHeight.Uint64(),
			IndexOffset:        signingInfo.IndexOffset.Uint64(),
			Tombstoned:         signingInfo.Tombstoned,
			MissedBlockCounter: signingInfo.MissedBlockCounter.Uint64(),
			IndicatorRate:      100 - float64(signingInfo.MissedBlockCounter.Uint64())/100,
			JailedUntil:        signingInfo.JailedUntil.Uint64(),
		},
	}
	v.Role = getValidatorRole(proposers, v)
	return v, nil
}

func (w *Wrapper) DelegatorsWithWorker(ctx context.Context, validatorSMC string) ([]*types.Delegator, error) {
	lgr := w.logger.With(zap.String("method", "DelegatorsWithWorker"))
	delegatorAddresses, err := w.pickTrusted().DelegatorAddresses(ctx, validatorSMC)
	if err != nil {
		lgr.Error("cannot get delegator address", zap.Error(err))
		return nil, err
	}

	var wg sync.WaitGroup
	nodesSize := len(w.publicNodes)
	type inputArgs struct {
		node             kardia.Node
		validatorSMCAddr string
		delegatorAddr    string
	}
	var delegators []*types.Delegator
	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	p, _ := ants.NewPoolWithFunc(nodesSize*WorkerRate, func(i interface{}) {
		defer wg.Done()
		input := i.(inputArgs)
		v, err := w.DelegatorWithNode(ctx, input.node, input.validatorSMCAddr, input.delegatorAddr)
		if err != nil {
			lgr.Debug("cannot get delegator", zap.Error(err))
			return
		}
		delegators = append(delegators, v)

	})
	defer p.Release()
	// Submit tasks one by one.
	for id, addr := range delegatorAddresses {
		wg.Add(1)
		nodeId := id % nodesSize
		node := w.publicNodes[nodeId]
		input := inputArgs{
			node:             node,
			delegatorAddr:    addr.Hex(),
			validatorSMCAddr: validatorSMC,
		}
		_ = p.Invoke(input)
	}
	wg.Wait()
	return delegators, nil
}

func (w *Wrapper) Delegator(ctx context.Context, validatorSMCAddress string, delegatorAddress string) (*types.Delegator, error) {
	reward, err := w.pickTrusted().DelegationRewards(ctx, validatorSMCAddress, delegatorAddress)
	if err != nil {
		return nil, err
	}

	stakedAmount, err := w.pickTrusted().DelegatorStakedAmount(ctx, validatorSMCAddress, delegatorAddress)
	if err != nil {
		return nil, err
	}
	d := &types.Delegator{
		Address:             delegatorAddress,
		StakedAmount:        stakedAmount.String(),
		ValidatorSMCAddress: validatorSMCAddress,
		Reward:              reward.String(),
	}

	return d, nil
}

func (w *Wrapper) DelegatorWithNode(ctx context.Context, node kardia.Node, validatorSMCAddress string, delegatorAddress string) (*types.Delegator, error) {
	lgr := w.logger.With(zap.String("method", "DelegatorWithNode"))
	reward, err := node.DelegationRewards(ctx, validatorSMCAddress, delegatorAddress)
	if err != nil {
		lgr.Error("cannot get delegation rewards", zap.Error(err), zap.Any("node", node.Url()))
		return nil, err
	}

	stakedAmount, err := node.DelegatorStakedAmount(ctx, validatorSMCAddress, delegatorAddress)
	if err != nil {
		lgr.Error("cannot get delegation staked amount", zap.Error(err))
		return nil, err
	}
	d := &types.Delegator{
		Address:             delegatorAddress,
		ValidatorSMCAddress: validatorSMCAddress,
		StakedAmount:        stakedAmount.String(),
		Reward:              reward.String(),
	}

	return d, nil
}

func (w *Wrapper) Delegators(ctx context.Context, validatorSMCAddr string) ([]*types.Delegator, error) {
	delegatorAddresses, err := w.pickTrusted().DelegatorAddresses(ctx, validatorSMCAddr)
	if err != nil {
		return nil, err
	}

	var delegators []*types.Delegator
	for _, addr := range delegatorAddresses {
		d, err := w.Delegator(ctx, validatorSMCAddr, addr.Hex())
		if err != nil {
			return nil, err
		}
		delegators = append(delegators, d)
	}

	return delegators, nil
}

func (w *Wrapper) UnbondedRecords(ctx context.Context, validatorSMCAddress, delegatorAddress string) ([]*types.UnbondedRecord, error) {
	nRecord, err := w.pickTrusted().UnbondedRecords(ctx, validatorSMCAddress, delegatorAddress)
	if err != nil {
		return nil, err
	}
	fmt.Println("re", nRecord)
	var records []*types.UnbondedRecord

	//for id := range nRecord.CompletionTimes {
	//	records = append(records, &types.UnbondedRecord{
	//		Balance:        nRecord.Balances[id],
	//		CompletionTime: nRecord.CompletionTimes[id],
	//	})
	//}
	//return records, nil
	//// Preprocess unbonded records
	//var ubdRecords []*types.UnbondedRecord
	//totalUnbondedAmount := new(big.Int).SetInt64(0)
	//totalWithdrawableAmount := new(big.Int).SetInt64(0)
	//now := new(big.Int).SetInt64(time.Now().Unix())
	//for _, r := range unbondedRecords {
	//	if r.CompletionTime.Cmp(now) == -1 {
	//		totalWithdrawableAmount = new(big.Int).Add(totalWithdrawableAmount, r.Balance)
	//	} else {
	//		totalUnbondedAmount = new(big.Int).Add(totalUnbondedAmount, r.Balance)
	//	}
	//
	//	ubdRecords = append(ubdRecords, &types.UnbondedRecord{
	//		Balances:        r.Balance.String(),
	//		CompletionTimes: r.CompletionTime.String(),
	//	})
	//}
	return records, nil
}

func (w *Wrapper) ValidatorOfDelegator(ctx context.Context, delegatorAddress string) {
	return
}
