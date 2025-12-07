package ftso

import (
	"encoding/json"
	"net/http"
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

