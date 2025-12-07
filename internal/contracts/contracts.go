package contracts

import (
	"encoding/hex"
	"fmt"
	"lfts/internal/ftso"
	"math/big"
	"strings"
)

// ContractCall represents a contract function call
type ContractCall struct {
	To     string `json:"to"`     // Contract address
	Data   string `json:"data"`   // Function call data (hex encoded)
	Method string `json:"method"` // Function name (for logging)
}

// ContractResponse represents the response from a contract call
type ContractResponse struct {
	Result string `json:"result"` // Return value (hex encoded)
	Error  string `json:"error,omitempty"`
}

// Mock contract addresses
const (
	FTSOContractAddress = "0x0000000000000000000000000000000000000001"
	FDCContractAddress  = "0x0000000000000000000000000000000000000002"
)

// HandleContractCall simulates a contract call
func HandleContractCall(call ContractCall) (*ContractResponse, error) {
	// Route to appropriate contract handler
	switch strings.ToLower(call.To) {
	case FTSOContractAddress:
		return handleFTSOCall(call)
	case FDCContractAddress:
		return handleFDCCall(call)
	default:
		return &ContractResponse{
			Error: "Unknown contract address",
		}, nil
	}
}

// handleFTSOCall handles calls to the mock FTSO contract
func handleFTSOCall(call ContractCall) (*ContractResponse, error) {
	// Parse function selector (first 4 bytes)
	if len(call.Data) < 10 {
		return &ContractResponse{Error: "Invalid call data"}, nil
	}

	selector := call.Data[:10] // 0x + 4 bytes

	// Function selectors (first 4 bytes of keccak256(function signature))
	// getCurrentPrice(address) = 0x893d20e8
	// getPrice(address,uint256) = 0x4b750334
	// getPriceAt(address,uint256) = 0x... (placeholder)

	switch selector {
	case "0x893d20e8": // getCurrentPrice(address)
		return handleGetCurrentPrice(call.Data)
	case "0x4b750334": // getPrice(address,uint256)
		return handleGetPrice(call.Data)
	default:
		return &ContractResponse{Error: "Unknown function selector"}, nil
	}
}

// handleGetCurrentPrice implements getCurrentPrice(address asset) returns (uint256 price, uint256 timestamp)
func handleGetCurrentPrice(data string) (*ContractResponse, error) {
	// Extract asset address from data (skip 0x and selector)
	if len(data) < 74 { // 0x + 4 bytes selector + 32 bytes address
		return &ContractResponse{Error: "Invalid call data"}, nil
	}

	// For simplicity, we'll use the last 20 bytes as asset identifier
	// In real implementation, this would decode the address parameter
	assetHex := data[len(data)-40:] // Last 20 bytes (40 hex chars)
	
	// Map common addresses to asset symbols (simplified)
	asset := addressToAsset(assetHex)

	price, err := ftso.GetPrice(asset)
	if err != nil {
		return &ContractResponse{Error: err.Error()}, nil
	}

	if price == nil {
		return &ContractResponse{Error: "Price not found"}, nil
	}

	// Encode return values: (uint256 price, uint256 timestamp)
	priceBig := big.NewInt(int64(price.Price * 1e8)) // Scale to 8 decimals
	timestampBig := big.NewInt(price.Timestamp)

	// Pack: 32 bytes price + 32 bytes timestamp
	result := fmt.Sprintf("0x%064s%064s",
		priceBig.Text(16),
		timestampBig.Text(16))

	return &ContractResponse{Result: result}, nil
}

// handleGetPrice implements getPrice(address asset, uint256 epoch) returns (uint256 price)
func handleGetPrice(data string) (*ContractResponse, error) {
	// This would decode asset and epoch parameters
	// For now, return error (not fully implemented)
	return &ContractResponse{Error: "getPrice(asset, epoch) not fully implemented"}, nil
}

// handleFDCCall handles calls to the mock FDC contract
func handleFDCCall(call ContractCall) (*ContractResponse, error) {
	if len(call.Data) < 10 {
		return &ContractResponse{Error: "Invalid call data"}, nil
	}

	selector := call.Data[:10]

	// getData(string memory feedName) = 0x... (placeholder)
	switch selector {
	default:
		return &ContractResponse{Error: "Unknown FDC function"}, nil
	}
}

// addressToAsset maps contract addresses to asset symbols (simplified)
func addressToAsset(addressHex string) string {
	// Common asset addresses (simplified mapping)
	// In real implementation, this would be a proper mapping
	addressMap := map[string]string{
		"0000000000000000000000000000000000000001": "BTC",
		"0000000000000000000000000000000000000002": "ETH",
		"0000000000000000000000000000000000000003": "XRP",
	}

	// Normalize address (remove leading zeros, lowercase)
	normalized := strings.ToLower(strings.TrimPrefix(addressHex, "0x"))
	normalized = strings.TrimLeft(normalized, "0")
	if normalized == "" {
		normalized = "0"
	}

	// Try to find in map
	for addr, asset := range addressMap {
		if strings.HasSuffix(normalized, strings.TrimLeft(addr, "0")) {
			return asset
		}
	}

	// Default: use last 4 chars as asset code
	if len(normalized) >= 4 {
		return strings.ToUpper(normalized[len(normalized)-4:])
	}

	return "UNKNOWN"
}

// EncodeString encodes a string parameter for contract calls
func EncodeString(s string) string {
	// Simple encoding: length + string bytes
	bytes := []byte(s)
	length := big.NewInt(int64(len(bytes)))
	return fmt.Sprintf("0x%064s%064s",
		length.Text(16),
		hex.EncodeToString(bytes))
}

// DecodeUint256 decodes a uint256 from hex string
func DecodeUint256(hexStr string) (*big.Int, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	result, ok := new(big.Int).SetString(hexStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid hex string: %s", hexStr)
	}
	return result, nil
}

