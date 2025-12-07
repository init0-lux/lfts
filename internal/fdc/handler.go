package fdc

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// HandleFeed handles GET /fdc/feed?name=<feed_name>
func HandleFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feedName := r.URL.Query().Get("name")
	if feedName == "" {
		http.Error(w, "Missing name parameter", http.StatusBadRequest)
		return
	}

	feed, err := GetFeed(feedName)
	if err != nil {
		http.Error(w, "Error retrieving feed", http.StatusInternalServerError)
		return
	}

	if feed == nil {
		http.Error(w, "Feed not found: "+feedName, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}

// HandleInjectFeed handles POST /fdc/inject?name=<feed_name>
func HandleInjectFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feedName := r.URL.Query().Get("name")
	if feedName == "" {
		http.Error(w, "Missing name parameter", http.StatusBadRequest)
		return
	}

	// Read JSON data from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		http.Error(w, "Empty request body", http.StatusBadRequest)
		return
	}

	// Parse JSON data
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Invalid JSON data: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = SetFeed(feedName, data)
	if err != nil {
		http.Error(w, "Error setting feed", http.StatusInternalServerError)
		return
	}

	feed, _ := GetFeed(feedName)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feed)
}

// HandleFeedHistory handles GET /fdc/history?name=<feed_name>&limit=<limit>
func HandleFeedHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feedName := r.URL.Query().Get("name")
	if feedName == "" {
		http.Error(w, "Missing name parameter", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}

		history, err := GetFeedHistory(feedName)
		if err != nil {
			http.Error(w, "Error retrieving feed history", http.StatusInternalServerError)
			return
		}

		if history == nil || len(history.History) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]FeedPoint{})
			return
		}

		// Return last N entries
		start := len(history.History) - limit
		if start < 0 {
			start = 0
		}
		result := history.History[start:]

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Return full history
	history, err := GetFeedHistory(feedName)
	if err != nil {
		http.Error(w, "Error retrieving feed history", http.StatusInternalServerError)
		return
	}

	if history == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FDCFeedHistory{
			FeedName: feedName,
			History:  []FeedPoint{},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// HandleListFeeds handles GET /fdc/list
func HandleListFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	feeds, err := GetAllFeeds()
	if err != nil {
		http.Error(w, "Error retrieving feeds", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"feeds": feeds,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

