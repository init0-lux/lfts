package contracts

import (
	"encoding/json"
	"fmt"
	"lfts/internal/chain"
	"net/http"
)

// JSONRPCRequest represents a JSON-RPC request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// EthCallParams represents parameters for eth_call
type EthCallParams struct {
	To   string `json:"to"`
	Data string `json:"data"`
}

// HandleJSONRPC handles JSON-RPC requests
func HandleJSONRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, nil, -32700, "Parse error", err.Error())
		return
	}

	var resp JSONRPCResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "eth_call":
		resp.Result = handleEthCall(req.Params)
	case "eth_blockNumber":
		resp.Result = handleEthBlockNumber()
	case "eth_getBlockByNumber":
		resp.Result = handleEthGetBlockByNumber(req.Params)
	default:
		resp.Error = &RPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleEthCall processes eth_call requests
func handleEthCall(params json.RawMessage) interface{} {
	var callParams []interface{}
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil
	}

	if len(callParams) < 1 {
		return "0x"
	}

	// Parse call object
	callObj, ok := callParams[0].(map[string]interface{})
	if !ok {
		return "0x"
	}

	to, _ := callObj["to"].(string)
	data, _ := callObj["data"].(string)

	call := ContractCall{
		To:   to,
		Data: data,
	}

	response, err := HandleContractCall(call)
	if err != nil {
		return "0x"
	}

	if response.Error != "" {
		return "0x"
	}

	return response.Result
}

// handleEthBlockNumber returns current block number
func handleEthBlockNumber() interface{} {
	chainInstance := chain.GetInstance()
	if chainInstance == nil {
		return "0x0"
	}

	height := chainInstance.GetHeight()
	return fmt.Sprintf("0x%x", height)
}

// handleEthGetBlockByNumber returns block information
func handleEthGetBlockByNumber(params json.RawMessage) interface{} {
	var blockParams []interface{}
	if err := json.Unmarshal(params, &blockParams); err != nil {
		return nil
	}

	if len(blockParams) < 1 {
		return nil
	}

	blockNumStr, _ := blockParams[0].(string)
	chainInstance := chain.GetInstance()
	if chainInstance == nil {
		return nil
	}

	block := chainInstance.GetLatestBlock()
	if block == nil {
		return nil
	}

	// Check if requesting latest block
	if blockNumStr == "latest" || blockNumStr == "0x" {
		return map[string]interface{}{
			"number":       fmt.Sprintf("0x%x", block.Number),
			"timestamp":    fmt.Sprintf("0x%x", block.Timestamp),
			"transactions": []interface{}{},
		}
	}

	return nil
}

// sendError sends a JSON-RPC error response
func sendError(w http.ResponseWriter, id interface{}, code int, message, data string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

