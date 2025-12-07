package rpc

import (
	"encoding/json"
	"lfts/internal/chain"
	"lfts/internal/contracts"
	"lfts/internal/fdc"
	"lfts/internal/ftso"
	"net/http"
	"strconv"
)

// HandleStatus handles GET /status
func HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chainInstance := chain.GetInstance()
	response := map[string]interface{}{
		"running":       chainInstance.IsRunning(),
		"height":        chainInstance.GetHeight(),
		"lastBlockTime": chainInstance.GetLastBlockTime(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleLatestBlock handles GET /block/latest
func HandleLatestBlock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chainInstance := chain.GetInstance()
	latestBlock := chainInstance.GetLatestBlock()

	if latestBlock == nil {
		response := map[string]interface{}{
			"number":    0,
			"timestamp": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"number":    latestBlock.Number,
		"timestamp": latestBlock.Timestamp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleFTSOPrice delegates to ftso package handler
func HandleFTSOPrice(w http.ResponseWriter, r *http.Request) {
	ftso.HandlePrice(w, r)
}

// HandleFTSOPriceHistory delegates to ftso package handler
func HandleFTSOPriceHistory(w http.ResponseWriter, r *http.Request) {
	ftso.HandlePriceHistory(w, r)
}

// HandleFTSOAllPrices handles GET /ftso/prices - returns all FTSO prices
func HandleFTSOAllPrices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	prices, err := ftso.GetAllPrices()
	if err != nil {
		http.Error(w, "Error retrieving prices", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"prices": prices,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleInjectFTSO handles POST /ftso/inject?asset=<asset>&price=<price>
func HandleInjectFTSO(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := r.URL.Query().Get("asset")
	priceStr := r.URL.Query().Get("price")

	if asset == "" || priceStr == "" {
		http.Error(w, "Missing asset or price parameter", http.StatusBadRequest)
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	err = ftso.SetPrice(asset, price)
	if err != nil {
		http.Error(w, "Error setting price", http.StatusInternalServerError)
		return
	}

	priceObj, _ := ftso.GetPrice(asset)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(priceObj)
}

// HandleFDCFeed delegates to fdc package handler
func HandleFDCFeed(w http.ResponseWriter, r *http.Request) {
	fdc.HandleFeed(w, r)
}

// HandleFDCInject delegates to fdc package handler
func HandleFDCInject(w http.ResponseWriter, r *http.Request) {
	fdc.HandleInjectFeed(w, r)
}

// HandleFDCHistory delegates to fdc package handler
func HandleFDCHistory(w http.ResponseWriter, r *http.Request) {
	fdc.HandleFeedHistory(w, r)
}

// HandleFDCList delegates to fdc package handler
func HandleFDCList(w http.ResponseWriter, r *http.Request) {
	fdc.HandleListFeeds(w, r)
}

// HandleJSONRPC delegates to contracts package handler
func HandleJSONRPC(w http.ResponseWriter, r *http.Request) {
	contracts.HandleJSONRPC(w, r)
}

