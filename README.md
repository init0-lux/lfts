# Local Flare Testnet Sandbox (LFTS) - FTSO MVP

A lightweight, standalone Flare-like local blockchain simulator with FTSO (Flare Time Series Oracle) mock simulation, built in Go.

## Purpose

LFTS provides a minimal local testnet environment for testing FTSO price feeds and blockchain interactions without requiring a full Flare network setup. It's designed to be hackathon-friendly, simple to understand, and easy to extend.

## Features

- **Minimal Chain Engine**: Single-node blockchain simulator with configurable block generation
- **FTSO Mock Oracle**: Simulate price feeds for various assets
- **HTTP RPC API**: Simple REST endpoints for querying chain state and FTSO prices
- **CLI Tool**: Easy-to-use command-line interface for managing the sandbox
- **Docker Support**: Containerized deployment option

## Building

### Prerequisites

- Go 1.22 or later
- Docker (optional, for containerized deployment)

### Build from Source

```bash
# Clone or navigate to the project directory
cd LFTS

# Download dependencies
go mod download

# Build the binary
go build -o lfts ./cmd/lfts
```

## Running

### Start the Chain

```bash
# Start with default settings (1 second block time, port 9650)
./lfts start

# Start with custom block time (in milliseconds)
./lfts start --block-time 2000

# Start with custom RPC port
./lfts start --port 8080
```

The chain will start generating blocks and the RPC server will be available on the specified port.

### Inject FTSO Prices

In a separate terminal:

```bash
# Inject a BTC price
./lfts inject ftso BTC 65000

# Inject an ETH price
./lfts inject ftso ETH 3500

# Inject any asset price
./lfts inject ftso XRP 0.5
```

### Check Status

```bash
# View chain status and FTSO prices
./lfts status
```

### Stop the Chain

Press `Ctrl+C` in the terminal where the chain is running, or use:

```bash
./lfts stop
```

## RPC API Endpoints

The RPC server runs on port `9650` by default.

### GET /status

Returns the current chain status.

**Response:**
```json
{
  "running": true,
  "height": 42,
  "lastBlockTime": 1710000000
}
```

### GET /ftso/price?asset=BTC

Returns the latest FTSO price for the specified asset.

**Response:**
```json
{
  "asset": "BTC",
  "price": 65000.0,
  "timestamp": 1710000000
}
```

**Example:**
```bash
curl http://localhost:9650/ftso/price?asset=BTC
```

### GET /block/latest

Returns the latest block information.

**Response:**
```json
{
  "number": 42,
  "timestamp": 1710000000
}
```

## Docker Quick Start

### Build the Docker Image

```bash
docker build -t lfts .
```

### Run the Container

```bash
docker run -d -p 9650:9650 --name lfts-sandbox lfts
```

### Access the Container

```bash
# Inject a price
docker exec lfts-sandbox ./lfts inject ftso BTC 65000

# Check status
docker exec lfts-sandbox ./lfts status
```

### Stop the Container

```bash
docker stop lfts-sandbox
docker rm lfts-sandbox
```

## Example Workflow

1. **Start the chain:**
   ```bash
   ./lfts start
   ```

2. **In another terminal, inject some prices:**
   ```bash
   ./lfts inject ftso BTC 65000
   ./lfts inject ftso ETH 3500
   ./lfts inject ftso XRP 0.5
   ```

3. **Query prices via RPC:**
   ```bash
   curl http://localhost:9650/ftso/price?asset=BTC
   curl http://localhost:9650/ftso/price?asset=ETH
   ```

4. **Check status:**
   ```bash
   ./lfts status
   ```

## Project Structure

```
lfts/
├── cmd/
│   └── lfts/
│       ├── main.go          # CLI entry point
│       └── commands.go      # Cobra command definitions
├── internal/
│   ├── chain/               # Chain engine
│   │   ├── chain.go         # Chain state management
│   │   ├── block.go         # Block structure
│   │   └── loop.go          # Block generation loop
│   ├── state/               # State storage
│   │   ├── state.go         # Global state interface
│   │   └── storage.go       # Thread-safe key-value store
│   ├── ftso/                # FTSO oracle module
│   │   ├── ftso.go          # FTSO price management
│   │   └── handler.go       # HTTP handlers
│   ├── rpc/                 # RPC server
│   │   ├── rpc.go           # Server setup
│   │   └── routes.go        # Route handlers
│   └── utils/               # Utilities
│       └── logger.go        # Logging helpers
├── Dockerfile
├── go.mod
└── README.md
```

## Design Notes

- **In-Memory Storage**: State is stored in memory (no persistence). This can be extended with LevelDB or similar.
- **Thread-Safe**: All state operations use mutexes for concurrent access safety.
- **Simple Architecture**: Minimal dependencies, easy to understand and modify.
- **Extensible**: Code structure allows for easy addition of features like persistence, more RPC endpoints, or additional oracle types.

## License

This is a minimal MVP for development and testing purposes.

