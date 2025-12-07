package rpc

import (
	"encoding/json"
	"lfts/internal/chain"
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

