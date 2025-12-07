package chain

import "time"

// Block represents a single block in the chain
type Block struct {
	Number    uint64    `json:"number"`
	Timestamp int64     `json:"timestamp"`
	Data      BlockData `json:"data"`
}

// BlockData contains the state updates included in the block
type BlockData struct {
	StateUpdates map[string]string `json:"stateUpdates,omitempty"`
}

// NewBlock creates a new block with the given number
func NewBlock(number uint64) *Block {
	return &Block{
		Number:    number,
		Timestamp: time.Now().Unix(),
		Data: BlockData{
			StateUpdates: make(map[string]string),
		},
	}
}

