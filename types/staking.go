package types

type ProposalDetail struct {
	ProposalMetadata
	VoteYes             uint64           `json:"voteYes" bson:"voteYes"`
	VoteNo              uint64           `json:"voteNo" bson:"voteNo"`
	VoteAbstain         uint64           `json:"voteAbstain" bson:"voteAbstain"`
	Params              []*NetworkParams `json:"params" bson:"params"`
	NumberOfVoteYes     uint64           `json:"numberOfVoteYes" bson:"numberOfVoteYes"`
	NumberOfVoteNo      uint64           `json:"numberOfVoteNo" bson:"numberOfVoteNo"`
	NumberOfVoteAbstain uint64           `json:"numberOfVoteAbstain" bson:"numberOfVoteAbstain"`
	UpdateTime          int64            `json:"updateTime" bson:"updateTime"`
}

type ProposalMetadata struct {
	ID        uint64 `json:"id" bson:"id"`
	Proposer  string `json:"nominator" bson:"nominator"`
	StartTime uint64 `json:"startTime" bson:"startTime"`
	EndTime   uint64 `json:"endTime" bson:"endTime"`
	Deposit   string `json:"deposit" bson:"deposit"`
	Status    uint8  `json:"status" bson:"status"`
}

type NetworkParams struct {
	LabelName string      `json:"labelName" bson:"labelName"`
	FromValue interface{} `json:"fromValue" bson:"fromValue"`
	ToValue   interface{} `json:"toValue" bson:"toValue"`
}
