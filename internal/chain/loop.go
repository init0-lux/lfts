package chain

import (
	"lfts/internal/utils"
	"time"
)

// StartLoop begins the block generation loop in a goroutine
func StartLoop(chain *Chain) {
	go func() {
		ticker := time.NewTicker(chain.GetBlockTime())
		defer ticker.Stop()

		// Create genesis block immediately
		block := chain.CreateBlock()
		utils.LogBlock(block.Number, block.Timestamp)

		for {
			select {
			case <-chain.GetStopChan():
				utils.Info("Chain loop stopped")
				return
			case <-ticker.C:
				if chain.IsRunning() {
					block := chain.CreateBlock()
					utils.LogBlock(block.Number, block.Timestamp)
				}
			}
		}
	}()
}

