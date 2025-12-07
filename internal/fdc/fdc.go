package fdc

import (
	"encoding/json"
	"lfts/internal/chain"
	"lfts/internal/state"
	"time"
)

const (
	// MaxHistoryEntries limits the number of historical feed entries stored per feed
	MaxHistoryEntries = 1000
)

// FDCFeed represents a data feed from the FDC connector
type FDCFeed struct {
	FeedName  string                 `json:"feedName"`
	Data      map[string]interface{} `json:"data"` // Arbitrary JSON data
	Timestamp int64                   `json:"timestamp"`
	BlockNum  uint64                  `json:"blockNum,omitempty"`
}

// FeedPoint represents a single feed entry in history
type FeedPoint struct {
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
	BlockNum  uint64                 `json:"blockNum"`
}

// FDCFeedHistory represents the full feed history
type FDCFeedHistory struct {
	FeedName string      `json:"feedName"`
	Latest   *FDCFeed    `json:"latest"`
	History  []FeedPoint `json:"history"`
}

// SetFeed stores a feed entry for the given feed name
func SetFeed(feedName string, data map[string]interface{}) error {
	chainInstance := chain.GetInstance()
	var blockNum uint64
	if chainInstance != nil {
		blockNum = chainInstance.GetHeight()
	}

	now := time.Now().Unix()
	feed := FDCFeed{
		FeedName:  feedName,
		Data:      data,
		Timestamp: now,
		BlockNum:  blockNum,
	}

	// Store latest feed
	latestKey := "fdc:" + feedName + ":latest"
	latestData, err := json.Marshal(feed)
	if err != nil {
		return err
	}
	if err := state.Set(latestKey, latestData); err != nil {
		return err
	}

	// Add to history
	historyKey := "fdc:" + feedName + ":history"
	historyData, err := state.Get(historyKey)
	if err != nil {
		return err
	}

	var history FDCFeedHistory
	if historyData != nil {
		if err := json.Unmarshal(historyData, &history); err != nil {
			// If unmarshal fails, start fresh
			history = FDCFeedHistory{
				FeedName: feedName,
				History:  []FeedPoint{},
			}
		}
	} else {
		history = FDCFeedHistory{
			FeedName: feedName,
			History:  []FeedPoint{},
		}
	}

	// Add new feed point
	history.Latest = &feed
	history.History = append(history.History, FeedPoint{
		Data:      data,
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

// GetFeed retrieves the latest feed for the given feed name
func GetFeed(feedName string) (*FDCFeed, error) {
	key := "fdc:" + feedName + ":latest"
	data, err := state.Get(key)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var feed FDCFeed
	err = json.Unmarshal(data, &feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}

// GetFeedHistory retrieves the full feed history
func GetFeedHistory(feedName string) (*FDCFeedHistory, error) {
	key := "fdc:" + feedName + ":history"
	data, err := state.Get(key)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var history FDCFeedHistory
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, err
	}

	return &history, nil
}

// GetAllFeeds retrieves all FDC feeds from state
func GetAllFeeds() (map[string]*FDCFeed, error) {
	allKeys := state.GlobalState.GetAllKeys()
	feeds := make(map[string]*FDCFeed)

	for _, key := range allKeys {
		if len(key) > 4 && key[:4] == "fdc:" && key[len(key)-7:] == ":latest" {
			feedName := key[4 : len(key)-7]
			feed, err := GetFeed(feedName)
			if err == nil && feed != nil {
				feeds[feedName] = feed
			}
		}
	}

	return feeds, nil
}

