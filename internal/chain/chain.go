package chain

import (
	"sync"
	"time"
)

// Chain represents the blockchain state
type Chain struct {
	mu            sync.RWMutex
	currentHeight uint64
	latestBlock   *Block
	running       bool
	blockTime     time.Duration
	stopChan      chan struct{}
}

// NewChain creates a new chain instance
func NewChain(blockTimeMs int) *Chain {
	return &Chain{
		currentHeight: 0,
		blockTime:     time.Duration(blockTimeMs) * time.Millisecond,
		stopChan:      make(chan struct{}),
		running:       false,
	}
}

// Start begins the block generation loop
func (c *Chain) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.stopChan = make(chan struct{})
	c.mu.Unlock()
}

// Stop halts the block generation loop
func (c *Chain) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return
	}
	c.running = false
	close(c.stopChan)
}

// IsRunning returns whether the chain is running
func (c *Chain) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// GetHeight returns the current block height
func (c *Chain) GetHeight() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentHeight
}

// GetLatestBlock returns the latest block
func (c *Chain) GetLatestBlock() *Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latestBlock
}

// GetLastBlockTime returns the timestamp of the latest block
func (c *Chain) GetLastBlockTime() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.latestBlock == nil {
		return 0
	}
	return c.latestBlock.Timestamp
}

// GetStopChan returns the stop channel for the loop
func (c *Chain) GetStopChan() <-chan struct{} {
	return c.stopChan
}

// GetBlockTime returns the configured block time
func (c *Chain) GetBlockTime() time.Duration {
	return c.blockTime
}

// CreateBlock creates and stores a new block
func (c *Chain) CreateBlock() *Block {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentHeight++
	block := NewBlock(c.currentHeight)
	c.latestBlock = block
	return block
}

// Global chain instance
var globalChain *Chain

// GetInstance returns the global chain instance
func GetInstance() *Chain {
	return globalChain
}

// SetInstance sets the global chain instance
func SetInstance(chain *Chain) {
	globalChain = chain
}

