package autoupdate

import (
	"lfts/internal/chain"
	"lfts/internal/ftso"
	"lfts/internal/utils"
	"math"
	"math/rand"
	"time"
)

// Pattern represents the update pattern type
type Pattern string

const (
	PatternRandom  Pattern = "random"
	PatternSine    Pattern = "sine"
	PatternCrash   Pattern = "crash"
	PatternSpike   Pattern = "spike"
	PatternStable  Pattern = "stable"
)

// Config holds auto-update configuration
type Config struct {
	Enabled    bool
	Interval   time.Duration
	Pattern    Pattern
	Assets     []string
	BasePrices map[string]float64
	Volatility float64
	StopChan   <-chan struct{}
}

// StartAutoUpdate starts the automatic price update loop
func StartAutoUpdate(config Config) {
	if !config.Enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(config.Interval)
		defer ticker.Stop()

		utils.Info("Auto-update started: pattern=%s, interval=%v, assets=%v", config.Pattern, config.Interval, config.Assets)

		// Initialize prices if not set
		for _, asset := range config.Assets {
			if _, exists := config.BasePrices[asset]; !exists {
				// Get current price or use default
				current, err := ftso.GetPrice(asset)
				if err == nil && current != nil {
					config.BasePrices[asset] = current.Price
				} else {
					// Default prices for common assets
					defaultPrices := map[string]float64{
						"BTC": 50000.0,
						"ETH": 3000.0,
						"XRP": 0.5,
					}
					if price, ok := defaultPrices[asset]; ok {
						config.BasePrices[asset] = price
					} else {
						config.BasePrices[asset] = 100.0
					}
				}
			}
		}

		startTime := time.Now()
		updateCount := 0

		for {
			select {
			case <-config.StopChan:
				utils.Info("Auto-update stopped")
				return
			case <-ticker.C:
				chainInstance := chain.GetInstance()
				if chainInstance == nil || !chainInstance.IsRunning() {
					continue
				}

				// Update each asset
				for _, asset := range config.Assets {
					basePrice := config.BasePrices[asset]
					newPrice := calculateNewPrice(config.Pattern, basePrice, config.Volatility, startTime, updateCount)
					
					err := ftso.SetPrice(asset, newPrice)
					if err == nil {
						config.BasePrices[asset] = newPrice
						utils.Info("Auto-updated %s: %.2f", asset, newPrice)
					}
				}

				updateCount++
			}
		}
	}()
}

// calculateNewPrice calculates the new price based on the pattern
func calculateNewPrice(pattern Pattern, basePrice, volatility float64, startTime time.Time, updateCount int) float64 {
	switch pattern {
	case PatternRandom:
		// Random walk: change by random percentage
		changePercent := (rand.Float64() - 0.5) * volatility * 2 // -volatility to +volatility
		return basePrice * (1 + changePercent/100)

	case PatternSine:
		// Sine wave: oscillating price
		elapsed := time.Since(startTime).Seconds()
		amplitude := basePrice * volatility / 100
		period := 60.0 // 60 second period
		oscillation := math.Sin(2*math.Pi*elapsed/period) * amplitude
		return basePrice + oscillation

	case PatternCrash:
		// Crash scenario: gradual decline
		if updateCount < 10 {
			// First 10 updates: gradual decline
			declinePercent := float64(updateCount) * 2.0 // 2% per update
			return basePrice * (1 - declinePercent/100)
		}
		// After crash: slight recovery
		return basePrice * 0.8 * (1 + float64(updateCount-10)*0.1/100)

	case PatternSpike:
		// Spike scenario: sudden increase then gradual decline
		if updateCount < 5 {
			// Spike up
			spikePercent := float64(updateCount) * 10.0 // 10% per update
			return basePrice * (1 + spikePercent/100)
		}
		// Gradual decline after spike
		declinePercent := float64(updateCount-5) * 1.0
		return basePrice * 1.5 * (1 - declinePercent/100)

	case PatternStable:
		// Stable: minimal random changes
		changePercent := (rand.Float64() - 0.5) * 0.1 // ±0.05%
		return basePrice * (1 + changePercent/100)

	default:
		return basePrice
	}
}

