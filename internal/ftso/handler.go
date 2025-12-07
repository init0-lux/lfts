package ftso

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// HandlePrice handles GET /ftso/price?asset=<asset>
func HandlePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := r.URL.Query().Get("asset")
	if asset == "" {
		http.Error(w, "Missing asset parameter", http.StatusBadRequest)
		return
	}

	// Check if timestamp parameter is provided for historical price
	timestampStr := r.URL.Query().Get("timestamp")
	if timestampStr != "" {
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
			return
		}

		pricePoint, err := GetPriceAt(asset, timestamp)
		if err != nil {
			http.Error(w, "Error retrieving price", http.StatusInternalServerError)
			return
		}

		if pricePoint == nil {
			http.Error(w, "Price not found for asset at timestamp: "+asset, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pricePoint)
		return
	}

	price, err := GetPrice(asset)
	if err != nil {
		http.Error(w, "Error retrieving price", http.StatusInternalServerError)
		return
	}

	if price == nil {
		http.Error(w, "Price not found for asset: "+asset, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(price)
}

// HandlePriceHistory handles GET /ftso/history?asset=<asset>&from=<timestamp>&to=<timestamp>
func HandlePriceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	asset := r.URL.Query().Get("asset")
	if asset == "" {
		http.Error(w, "Missing asset parameter", http.StatusBadRequest)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	limitStr := r.URL.Query().Get("limit")

	// If limit is provided, return last N prices
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}

		history, err := GetPriceHistory(asset)
		if err != nil {
			http.Error(w, "Error retrieving price history", http.StatusInternalServerError)
			return
		}

		if history == nil || len(history.History) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]PricePoint{})
			return
		}

		// Return last N prices
		start := len(history.History) - limit
		if start < 0 {
			start = 0
		}
		result := history.History[start:]

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// If from/to are provided, return range
	if fromStr != "" && toStr != "" {
		from, err := strconv.ParseInt(fromStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid from timestamp", http.StatusBadRequest)
			return
		}

		to, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid to timestamp", http.StatusBadRequest)
			return
		}

		result, err := GetPriceHistoryRange(asset, from, to)
		if err != nil {
			http.Error(w, "Error retrieving price history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// If no parameters, return full history
	history, err := GetPriceHistory(asset)
	if err != nil {
		http.Error(w, "Error retrieving price history", http.StatusInternalServerError)
		return
	}

	if history == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FTSOPriceHistory{
			Asset:   asset,
			History: []PricePoint{},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

