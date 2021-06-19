// Package types
package types

type KRC20 struct {
	Address     string
	Owner       string
	Name        string
	Logo        string
	Symbol      string
	Info        string
	ABI         string
	Bytecode    string
	Decimals    uint64
	TotalSupply string
	Status      int // 0 = New, 1 = Verified
}
