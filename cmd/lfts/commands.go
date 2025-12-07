package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lfts/internal/autoupdate"
	"lfts/internal/chain"
	"lfts/internal/fdc"
	"lfts/internal/ftso"
	"lfts/internal/rpc"
	"lfts/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	blockTime      int
	rpcPort        string
	autoUpdateFTSO bool
	updateInterval int
	updatePattern  string
	updateAssets   []string
	volatility     float64
)

var rootCmd = &cobra.Command{
	Use:   "lfts",
	Short: "Local Flare Testnet Sandbox - FTSO MVP",
	Long:  "A lightweight, standalone Flare-like local testnet sandbox with FTSO mock simulation",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the chain node",
	Long:  "Starts the chain engine and RPC server",
	Run:   runStart,
}

var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Inject data into the chain",
	Long:  "Inject FTSO prices or other data",
}

var injectFTSOCmd = &cobra.Command{
	Use:   "ftso <asset> <price>",
	Short: "Inject an FTSO price",
	Long:  "Inject a fake FTSO price for the given asset",
	Args:  cobra.ExactArgs(2),
	Run:   runInjectFTSO,
}

var injectFDCCmd = &cobra.Command{
	Use:   "fdc <feed_name> <json_data>",
	Short: "Inject FDC feed data",
	Long:  "Inject JSON data for a FDC feed. Example: lfts inject fdc weather '{\"temp\":25,\"humidity\":60}'",
	Args:  cobra.ExactArgs(2),
	Run:   runInjectFDC,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show chain status",
	Long:  "Shows the current chain status including block height and FTSO prices",
	Run:   runStatus,
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show price history",
	Long:  "Shows price history for FTSO assets",
}

var historyFTSOCmd = &cobra.Command{
	Use:   "ftso <asset>",
	Short: "Show FTSO price history",
	Long:  "Shows price history for a specific asset",
	Args:  cobra.ExactArgs(1),
	Run:   runHistoryFTSO,
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query data feeds",
	Long:  "Query FTSO prices or FDC feeds",
}

var queryFDCCmd = &cobra.Command{
	Use:   "fdc <feed_name>",
	Short: "Query FDC feed",
	Long:  "Query the latest data for a FDC feed",
	Args:  cobra.ExactArgs(1),
	Run:   runQueryFDC,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available feeds",
	Long:  "List all available FTSO prices or FDC feeds",
}

var listFDCCmd = &cobra.Command{
	Use:   "fdc",
	Short: "List FDC feeds",
	Long:  "List all available FDC feeds",
	Run:   runListFDC,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the chain",
	Long:  "Stops the running chain loop",
	Run:   runStop,
}

func init() {
	startCmd.Flags().IntVarP(&blockTime, "block-time", "b", 1000, "Block generation interval in milliseconds")
	startCmd.Flags().StringVarP(&rpcPort, "port", "p", "9650", "RPC server port")
	startCmd.Flags().BoolVar(&autoUpdateFTSO, "auto-update-ftso", false, "Enable automatic FTSO price updates")
	startCmd.Flags().IntVar(&updateInterval, "update-interval", 1800, "Auto-update interval in milliseconds (default: 1800ms)")
	startCmd.Flags().StringVar(&updatePattern, "update-pattern", "random", "Update pattern: random, sine, crash, spike, stable")
	startCmd.Flags().StringSliceVar(&updateAssets, "update-assets", []string{"BTC", "ETH"}, "Assets to auto-update (comma-separated)")
	startCmd.Flags().Float64Var(&volatility, "volatility", 1.0, "Price volatility percentage (default: 1.0%)")

	injectCmd.PersistentFlags().StringVarP(&rpcPort, "port", "p", "9650", "RPC server port")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(injectCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(stopCmd)

	injectCmd.AddCommand(injectFTSOCmd)
	injectCmd.AddCommand(injectFDCCmd)
	historyCmd.AddCommand(historyFTSOCmd)
	queryCmd.AddCommand(queryFDCCmd)
	listCmd.AddCommand(listFDCCmd)
}

func runStart(cmd *cobra.Command, args []string) {
	utils.Info("Starting Local Flare Testnet Sandbox...")
	utils.Info("Block time: %d ms", blockTime)
	utils.Info("RPC port: %s", rpcPort)

	// Create and set chain instance
	chainInstance := chain.NewChain(blockTime)
	chain.SetInstance(chainInstance)

	// Start chain
	chainInstance.Start()
	chain.StartLoop(chainInstance)

	// Start RPC server
	rpcServer := rpc.NewServer(rpcPort)
	go func() {
		if err := rpcServer.Start(); err != nil {
			utils.Error("RPC server error: %v", err)
			os.Exit(1)
		}
	}()

	// Start auto-update if enabled
	stopChan := make(chan struct{})
	if autoUpdateFTSO {
		// Parse assets
		assets := updateAssets
		if len(assets) == 0 {
			assets = []string{"BTC", "ETH"}
		}

		// Get base prices from current state
		basePrices := make(map[string]float64)
		for _, asset := range assets {
			price, err := ftso.GetPrice(asset)
			if err == nil && price != nil {
				basePrices[asset] = price.Price
			}
		}

		autoupdate.StartAutoUpdate(autoupdate.Config{
			Enabled:    true,
			Interval:   time.Duration(updateInterval) * time.Millisecond,
			Pattern:    autoupdate.Pattern(updatePattern),
			Assets:     assets,
			BasePrices: basePrices,
			Volatility: volatility,
			StopChan:   stopChan,
		})
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	utils.Info("Chain is running. Press Ctrl+C to stop.")
	<-sigChan

	utils.Info("Shutting down...")
	close(stopChan) // Stop auto-update
	chainInstance.Stop()
	utils.Info("Chain stopped")
}

func runInjectFTSO(cmd *cobra.Command, args []string) {
	asset := args[0]
	priceStr := args[1]

	// Try to inject via RPC if chain is running
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s/ftso/inject?asset=%s&price=%s", rpcPort, asset, priceStr)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		utils.Error("Failed to create request: %v", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Chain might not be running, fall back to local injection
		price, parseErr := strconv.ParseFloat(priceStr, 64)
		if parseErr != nil {
			utils.Error("Invalid price: %v", parseErr)
			os.Exit(1)
		}

		err = ftso.SetPrice(asset, price)
		if err != nil {
			utils.Error("Failed to inject price (chain not running?): %v", err)
			os.Exit(1)
		}
		utils.Info("Injected FTSO price locally: %s = %.2f (Note: Chain must be running for RPC access)", asset, price)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 256)
		resp.Body.Read(body)
		utils.Error("Failed to inject price via RPC: status %d - %s", resp.StatusCode, string(body))
		os.Exit(1)
	}

	utils.Info("Injected FTSO price: %s = %s", asset, priceStr)
}

func runStatus(cmd *cobra.Command, args []string) {
	// Try to get status via RPC first
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s/status", rpcPort)

	resp, err := client.Get(url)
	if err != nil {
		// Chain might not be running, try local access
		chainInstance := chain.GetInstance()
		if chainInstance == nil {
			fmt.Println("Chain is not running. Start the chain with 'lfts start' first.")
			return
		}

		fmt.Println("=== Chain Status ===")
		fmt.Printf("Running: %v\n", chainInstance.IsRunning())
		fmt.Printf("Block Height: %d\n", chainInstance.GetHeight())
		fmt.Printf("Last Block Time: %d\n", chainInstance.GetLastBlockTime())

		// Get all FTSO prices
		prices, err := ftso.GetAllPrices()
		if err != nil {
			utils.Error("Error retrieving FTSO prices: %v", err)
			return
		}

		fmt.Println("\n=== FTSO Prices ===")
		if len(prices) == 0 {
			fmt.Println("No FTSO prices available")
		} else {
			for asset, price := range prices {
				fmt.Printf("%s: %.2f (timestamp: %d)\n", asset, price.Price, price.Timestamp)
			}
		}
		return
	}
	defer resp.Body.Close()

	// Parse RPC response
	var statusData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusData); err != nil {
		utils.Error("Error parsing status response: %v", err)
		return
	}

	fmt.Println("=== Chain Status ===")
	fmt.Printf("Running: %v\n", statusData["running"])
	fmt.Printf("Block Height: %.0f\n", statusData["height"])
	fmt.Printf("Last Block Time: %.0f\n", statusData["lastBlockTime"])

	// Get FTSO prices via RPC
	pricesUrl := fmt.Sprintf("http://localhost:%s/ftso/prices", rpcPort)
	pricesResp, err := client.Get(pricesUrl)
	if err == nil {
		defer pricesResp.Body.Close()
		var pricesData map[string]interface{}
		if err := json.NewDecoder(pricesResp.Body).Decode(&pricesData); err == nil {
			fmt.Println("\n=== FTSO Prices ===")
			if prices, ok := pricesData["prices"].(map[string]interface{}); ok {
				for asset, priceData := range prices {
					if priceMap, ok := priceData.(map[string]interface{}); ok {
						fmt.Printf("%s: %.2f (timestamp: %.0f)\n",
							asset,
							priceMap["price"],
							priceMap["timestamp"])
					}
				}
			} else {
				fmt.Println("No FTSO prices available")
			}
		}
	} else {
		// Fallback: try to get prices individually (if /ftso/prices doesn't exist)
		fmt.Println("\n=== FTSO Prices ===")
		fmt.Println("(Query individual prices with: curl http://localhost:" + rpcPort + "/ftso/price?asset=<asset>)")
	}
}

func runHistoryFTSO(cmd *cobra.Command, args []string) {
	asset := args[0]

	// Try to get history via RPC
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s/ftso/history?asset=%s&limit=10", rpcPort, asset)

	resp, err := client.Get(url)
	if err != nil {
		// Chain might not be running, try local access
		history, err := ftso.GetPriceHistory(asset)
		if err != nil {
			utils.Error("Error retrieving price history: %v", err)
			return
		}

		if history == nil || len(history.History) == 0 {
			fmt.Printf("No price history available for %s\n", asset)
			return
		}

		fmt.Printf("=== Price History for %s ===\n", asset)
		fmt.Printf("Total entries: %d\n\n", len(history.History))

		// Show last 10 entries
		start := len(history.History) - 10
		if start < 0 {
			start = 0
		}

		for i := len(history.History) - 1; i >= start; i-- {
			point := history.History[i]
			fmt.Printf("[%d] Price: %.2f, Timestamp: %d, Block: %d\n",
				i+1, point.Price, point.Timestamp, point.BlockNum)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Error("Failed to retrieve history: status %d", resp.StatusCode)
		return
	}

	var historyPoints []ftso.PricePoint
	if err := json.NewDecoder(resp.Body).Decode(&historyPoints); err != nil {
		utils.Error("Error parsing history response: %v", err)
		return
	}

	fmt.Printf("=== Price History for %s (Last 10) ===\n", asset)
	if len(historyPoints) == 0 {
		fmt.Println("No price history available")
		return
	}

		for i := len(historyPoints) - 1; i >= 0; i-- {
		point := historyPoints[i]
		fmt.Printf("Price: %.2f, Timestamp: %d, Block: %d\n",
			point.Price, point.Timestamp, point.BlockNum)
	}
}

func runInjectFDC(cmd *cobra.Command, args []string) {
	feedName := args[0]
	jsonDataStr := args[1]

	// Parse JSON data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonDataStr), &data); err != nil {
		utils.Error("Invalid JSON data: %v", err)
		os.Exit(1)
	}

	// Try to inject via RPC if chain is running
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s/fdc/inject?name=%s", rpcPort, feedName)

	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		utils.Error("Failed to create request: %v", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// Chain might not be running, fall back to local injection
		err = fdc.SetFeed(feedName, data)
		if err != nil {
			utils.Error("Failed to inject feed (chain not running?): %v", err)
			os.Exit(1)
		}
		utils.Info("Injected FDC feed locally: %s (Note: Chain must be running for RPC access)", feedName)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 256)
		resp.Body.Read(body)
		utils.Error("Failed to inject feed via RPC: status %d - %s", resp.StatusCode, string(body))
		os.Exit(1)
	}

	utils.Info("Injected FDC feed: %s", feedName)
}

func runQueryFDC(cmd *cobra.Command, args []string) {
	feedName := args[0]

	// Try to get feed via RPC
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s/fdc/feed?name=%s", rpcPort, feedName)

	resp, err := client.Get(url)
	if err != nil {
		// Chain might not be running, try local access
		feed, err := fdc.GetFeed(feedName)
		if err != nil {
			utils.Error("Error retrieving feed: %v", err)
			return
		}

		if feed == nil {
			fmt.Printf("Feed not found: %s\n", feedName)
			return
		}

		fmt.Printf("=== FDC Feed: %s ===\n", feedName)
		jsonData, _ := json.MarshalIndent(feed.Data, "", "  ")
		fmt.Printf("Data: %s\n", string(jsonData))
		fmt.Printf("Timestamp: %d\n", feed.Timestamp)
		fmt.Printf("Block: %d\n", feed.BlockNum)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Error("Failed to retrieve feed: status %d", resp.StatusCode)
		return
	}

	var feed fdc.FDCFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		utils.Error("Error parsing feed response: %v", err)
		return
	}

	fmt.Printf("=== FDC Feed: %s ===\n", feedName)
	jsonData, _ := json.MarshalIndent(feed.Data, "", "  ")
	fmt.Printf("Data: %s\n", string(jsonData))
	fmt.Printf("Timestamp: %d\n", feed.Timestamp)
	fmt.Printf("Block: %d\n", feed.BlockNum)
}

func runListFDC(cmd *cobra.Command, args []string) {
	// Try to get feeds via RPC
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%s/fdc/list", rpcPort)

	resp, err := client.Get(url)
	if err != nil {
		// Chain might not be running, try local access
		feeds, err := fdc.GetAllFeeds()
		if err != nil {
			utils.Error("Error retrieving feeds: %v", err)
			return
		}

		fmt.Println("=== FDC Feeds ===")
		if len(feeds) == 0 {
			fmt.Println("No FDC feeds available")
		} else {
			for feedName, feed := range feeds {
				fmt.Printf("%s: timestamp %d, block %d\n", feedName, feed.Timestamp, feed.BlockNum)
			}
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.Error("Failed to retrieve feeds: status %d", resp.StatusCode)
		return
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		utils.Error("Error parsing feeds response: %v", err)
		return
	}

	fmt.Println("=== FDC Feeds ===")
	if feeds, ok := response["feeds"].(map[string]interface{}); ok {
		if len(feeds) == 0 {
			fmt.Println("No FDC feeds available")
		} else {
			for feedName, feedData := range feeds {
				if feedMap, ok := feedData.(map[string]interface{}); ok {
					fmt.Printf("%s: timestamp %.0f, block %.0f\n",
						feedName,
						feedMap["timestamp"],
						feedMap["blockNum"])
				}
			}
		}
	} else {
		fmt.Println("No FDC feeds available")
	}
}

func runStop(cmd *cobra.Command, args []string) {
	chainInstance := chain.GetInstance()
	if chainInstance == nil {
		fmt.Println("Chain is not running")
		return
	}

	if !chainInstance.IsRunning() {
		fmt.Println("Chain is not running")
		return
	}

	chainInstance.Stop()
	utils.Info("Chain stopped")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

