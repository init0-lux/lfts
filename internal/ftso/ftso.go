package ftso

import (
	"encoding/json"
	"lfts/internal/chain"
	"lfts/internal/state"
	"sort"
	"time"
)

const (
	// MaxHistoryEntries limits the number of historical prices stored per asset
	MaxHistoryEntries = 1000
)

// FTSOPrice represents a price feed from the FTSO oracle
type FTSOPrice struct {
	Asset     string  `json:"asset"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	BlockNum  uint64  `json:"blockNum,omitempty"`
}

// PricePoint represents a single price point in history
type PricePoint struct {
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	BlockNum  uint64  `json:"blockNum"`
}

// FTSOPriceHistory represents the full price history for an asset
type FTSOPriceHistory struct {
	Asset  string       `json:"asset"`
	Latest *FTSOPrice   `json:"latest"`
	History []PricePoint `json:"history"`
}

// SetPrice stores a price for the given asset and adds it to history
func SetPrice(asset string, price float64) error {
	chainInstance := chain.GetInstance()
	var blockNum uint64
	if chainInstance != nil {
		blockNum = chainInstance.GetHeight()
	}

	now := time.Now().Unix()
	ftsoPrice := FTSOPrice{
		Asset:     asset,
		Price:     price,
		Timestamp: now,
		BlockNum:  blockNum,
	}

	// Store latest price
	latestKey := "ftso:" + asset + ":latest"
	latestData, err := json.Marshal(ftsoPrice)
	if err != nil {
		return err
	}
	if err := state.Set(latestKey, latestData); err != nil {
		return err
	}

	// Add to history
	historyKey := "ftso:" + asset + ":history"
	historyData, err := state.Get(historyKey)
	if err != nil {
		return err
	}

	var history FTSOPriceHistory
	if historyData != nil {
		if err := json.Unmarshal(historyData, &history); err != nil {
			// If unmarshal fails, start fresh
			history = FTSOPriceHistory{
				Asset:   asset,
				History: []PricePoint{},
			}
		}
	} else {
		history = FTSOPriceHistory{
			Asset:   asset,
			History: []PricePoint{},
		}
	}

	// Add new price point
	history.Latest = &ftsoPrice
	history.History = append(history.History, PricePoint{
		Price:     price,
		Timestamp: now,
		BlockNum:  blockNum,
	})

	// Limit history size
	if len(history.History) > MaxHistoryEntries {
		history.History = history.History[len(history.History)-MaxHistoryEntries:]
	}

	// Store updated history
	historyData, err = json.Marshal(history)
	if err != nil {
		return err
	}

	return state.Set(historyKey, historyData)
}

// GetPrice retrieves the latest price for the given asset
func GetPrice(asset string) (*FTSOPrice, error) {
	key := "ftso:" + asset + ":latest"
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

// GetPriceAt retrieves the price at or before the given timestamp
func GetPriceAt(asset string, timestamp int64) (*PricePoint, error) {
	history, err := GetPriceHistory(asset)
	if err != nil {
		return nil, err
	}

	if history == nil || len(history.History) == 0 {
		return nil, nil
	}

	// Find the price point at or before the timestamp
	// History is stored chronologically, so we can do a binary search
	idx := sort.Search(len(history.History), func(i int) bool {
		return history.History[i].Timestamp > timestamp
	})

	if idx == 0 {
		// All prices are after the requested timestamp
		return nil, nil
	}

	// Return the price point just before the first one after the timestamp
	return &history.History[idx-1], nil
}

// GetPriceHistory retrieves the full price history for an asset
func GetPriceHistory(asset string) (*FTSOPriceHistory, error) {
	key := "ftso:" + asset + ":history"
	data, err := state.Get(key)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var history FTSOPriceHistory
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}

	return &history, nil
}

// GetPriceHistoryRange retrieves price history within a time range
func GetPriceHistoryRange(asset string, fromTimestamp, toTimestamp int64) ([]PricePoint, error) {
	history, err := GetPriceHistory(asset)
	if err != nil {
		return nil, err
	}

	if history == nil {
		return []PricePoint{}, nil
	}

	var result []PricePoint
	for _, point := range history.History {
		if point.Timestamp >= fromTimestamp && point.Timestamp <= toTimestamp {
			result = append(result, point)
		}
	}

	return result, nil
}

// GetAllPrices retrieves all FTSO prices from state
func GetAllPrices() (map[string]*FTSOPrice, error) {
	allKeys := state.GlobalState.GetAllKeys()
	prices := make(map[string]*FTSOPrice)

	for _, key := range allKeys {
		if len(key) > 13 && key[:13] == "ftso:" && key[len(key)-7:] == ":latest" {
			asset := key[5 : len(key)-7]
			price, err := GetPrice(asset)
			if err == nil && price != nil {
				prices[asset] = price
			}
		}
	}

	return prices, nil
}

