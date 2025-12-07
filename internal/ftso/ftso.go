package ftso

import (
	"encoding/json"
	"lfts/internal/state"
	"time"
)

// FTSOPrice represents a price feed from the FTSO oracle
type FTSOPrice struct {
	Asset     string  `json:"asset"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

// SetPrice stores a price for the given asset
func SetPrice(asset string, price float64) error {
	ftsoPrice := FTSOPrice{
		Asset:     asset,
		Price:     price,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(ftsoPrice)
	if err != nil {
		return err
	}

	key := "ftso:" + asset
	return state.Set(key, data)
}

// GetPrice retrieves the latest price for the given asset
func GetPrice(asset string) (*FTSOPrice, error) {
	key := "ftso:" + asset
	data, err := state.Get(key)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var price FTSOPrice
	err = json.Unmarshal(data, &price)
	if err != nil {
		return nil, err
	}

	return &price, nil
}

// GetAllPrices retrieves all FTSO prices from state
func GetAllPrices() (map[string]*FTSOPrice, error) {
	allKeys := state.GlobalState.GetAllKeys()
	prices := make(map[string]*FTSOPrice)

	for _, key := range allKeys {
		if len(key) > 5 && key[:5] == "ftso:" {
			asset := key[5:]
			price, err := GetPrice(asset)
			if err == nil && price != nil {
				prices[asset] = price
			}
		}
	}

	return prices, nil
}

