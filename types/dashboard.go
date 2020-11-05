// Package types
package types

type TokenInfo struct {
	Name              string  `json:"name"`
	Symbol            string  `json:"symbol"`
	Decimal           int     `json:"decimal"`
	TotalSupply       int     `json:"total_supply"`
	CirculatingSupply int     `json:"circulating_supply"`
	Price             float64 `json:"price"`
	Volume24h         float64 `json:"volume_24h"`
	Change1h          float64 `json:"change_1h"`
	Change24h         float64 `json:"change_24h"`
	Change7d          float64 `json:"change_7d"`
	MarketCap         float64 `json:"market_cap"`
}
