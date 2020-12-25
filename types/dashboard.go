// Package types
package types

type TokenInfo struct {
	Name                     string  `json:"name"`
	Symbol                   string  `json:"symbol"`
	Decimal                  int64   `json:"decimal"`
	ERC20TotalSupply         int64   `json:"erc20_total_supply"`
	ERC20CirculatingSupply   int64   `json:"erc20_circulating_supply"`
	MainnetTotalSupply       int64   `json:"mainnet_total_supply"`
	MainnetCirculatingSupply int64   `json:"mainnet_circulating_supply"`
	Price                    float64 `json:"price"`
	Volume24h                float64 `json:"volume_24h"`
	Change1h                 float64 `json:"change_1h"`
	Change24h                float64 `json:"change_24h"`
	Change7d                 float64 `json:"change_7d"`
	ERC20MarketCap           float64 `json:"erc20_market_cap"`
	MainnetMarketCap         float64 `json:"mainnet_market_cap"`
}

type SupplyInfo struct {
	ERC20CirculatingSupply int64 `json:"erc20CirculatingSupply"`
	ERC20TotalSupply       int64 `json:"erc20TotalSupply"`
	MainnetTotalSupply     int64 `json:"mainnetTotalSupply"`
	MainnetGenesisAmount   int64 `json:"mainnetGenesisAmount"`
}
