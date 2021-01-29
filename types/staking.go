package types

type ProposalDetail struct {
	ProposalMetadata
	VoteYes     uint64            `json:"voteYes"`
	VoteNo      uint64            `json:"voteNo"`
	VoteAbstain uint64            `json:"voteAbstain"`
	Params      map[string]string `json:"params"`
}

type ProposalMetadata struct {
	ID        uint64 `json:"id"`
	Proposer  string `json:"nominator"`
	StartTime uint64 `json:"startTime"`
	EndTime   uint64 `json:"endTime"`
	Deposit   string `json:"deposit"`
	Status    uint8  `json:"status"`
}
