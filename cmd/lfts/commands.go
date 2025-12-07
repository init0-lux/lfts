package main

import (
	"fmt"
	"lfts/internal/chain"
	"lfts/internal/ftso"
	"lfts/internal/rpc"
	"lfts/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	blockTime int
	rpcPort   string
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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show chain status",
	Long:  "Shows the current chain status including block height and FTSO prices",
	Run:   runStatus,
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

	injectCmd.PersistentFlags().StringVarP(&rpcPort, "port", "p", "9650", "RPC server port")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(injectCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)

	injectCmd.AddCommand(injectFTSOCmd)
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

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	utils.Info("Chain is running. Press Ctrl+C to stop.")
	<-sigChan

	utils.Info("Shutting down...")
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
	chainInstance := chain.GetInstance()
	if chainInstance == nil {
		fmt.Println("Chain is not initialized. Run 'lfts start' first.")
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

