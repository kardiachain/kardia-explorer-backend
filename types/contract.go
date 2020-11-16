package types

import (
	"time"
)

type Contract struct {
	Address         string    `json:"address" bson:"address"`
	Bytecode        string    `json:"byteCode" bson:"byteCode,omitempty"`
	Valid           bool      `json:"valid" bson:"valid,omitempty"`
	ContractName    string    `json:"contractName" bson:"contractName,omitempty"`
	CompilerVersion string    `json:"compilerVersion" bson:"compilerVersion,omitempty"`
	EVMVersion      string    `json:"kvmVersion" bson:"kvmVersion,omitempty"`
	Optimization    bool      `json:"optimization" bson:"optimization,omitempty"`
	SourceCode      string    `json:"sourceCode" bson:"sourceCode,omitempty"`
	CreatedAt       time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt" bson:"updatedAt"`
	Abi             []AbiItem `json:"abi" bson:"abi"`
}

type AbiItem struct {
	Anonymous       bool          `json:"anonymous" bson:"anonymous"`
	Constant        bool          `json:"constant" bson:"constant"`
	Inputs          []AbiArgument `json:"inputs" bson:"inputs"`
	Name            string        `json:"name" bson:"name"`
	Outputs         []AbiArgument `json:"outputs" bson:"outputs"`
	Payable         bool          `json:"payable" bson:"payable"`
	StateMutability string        `json:"stateMutability" bson:"stateMutability"`
	Type            string        `json:"type" bson:"type"`
}

type AbiArgument struct {
	Name    string `json:"name" bson:"name"`
	Indexed bool   `json:"indexes" bson:"indexes"`
	Type    string `json:"type" bson:"type"`
}
