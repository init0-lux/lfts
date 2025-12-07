# Local Flare Testnet Sandbox (LFTS) - FTSO MVP

A lightweight, standalone Flare-like local blockchain simulator with FTSO (Flare Time Series Oracle) and FDC (Flare Data Connector) mock simulation, built in Go.

## Purpose

LFTS provides a minimal local testnet environment for testing FTSO price feeds, FDC data feeds, and smart contract interactions without requiring a full Flare network setup. It's designed to be hackathon-friendly, simple to understand, and easy to extend.

## Features

- **Minimal Chain Engine**: Single-node blockchain simulator with configurable block generation
- **FTSO Mock Oracle**: Simulate price feeds for various assets with full price history (up to 1000 entries)
- **FDC Mock Connector**: Simulate arbitrary JSON data feeds (weather, sports, custom data, etc.)
- **Price History**: Maintains historical price data with timestamp and block number tracking
- **Auto-Update Simulation**: Automatic price updates with configurable patterns (random, sine, crash, spike, stable)
- **Smart Contract Testing**: JSON-RPC endpoint for testing contract calls to FTSO and FDC
- **HTTP RPC API**: Simple REST endpoints for querying chain state, FTSO prices, and FDC feeds
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

# Start with automatic FTSO price updates
./lfts start --auto-update-ftso

# Start with auto-updates using specific pattern and assets
./lfts start --auto-update-ftso --update-pattern random --update-assets BTC,ETH --volatility 2.0

# Start with auto-updates every 1.8 seconds (matching real FTSO)
./lfts start --auto-update-ftso --update-interval 1800
```

**Auto-Update Options:**
- `--auto-update-ftso`: Enable automatic price updates
- `--update-interval <ms>`: Update interval in milliseconds (default: 1800ms)
- `--update-pattern <pattern>`: Update pattern - `random`, `sine`, `crash`, `spike`, or `stable` (default: random)
- `--update-assets <assets>`: Comma-separated list of assets to update (default: BTC,ETH)
- `--volatility <percent>`: Price volatility percentage (default: 1.0%)

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

**Note**: Each price injection is stored in history. Previous prices are preserved (up to 1000 entries per asset).

### Inject FDC Feeds

```bash
# Inject weather data
./lfts inject fdc weather '{"temp":25,"humidity":60,"condition":"sunny"}'

# Inject sports data
./lfts inject fdc sports '{"game":"Lakers vs Warriors","score":"98-95","quarter":4}'

# Inject custom JSON data
./lfts inject fdc custom '{"key":"value","number":42,"array":[1,2,3]}'
```

### Query Data

```bash
# Query FTSO price
./lfts query ftso BTC  # (via RPC, or use curl)

# Query FDC feed
./lfts query fdc weather

# List all FDC feeds
./lfts list fdc
```

### View Price History

```bash
# View price history for an asset (last 10 entries)
./lfts history ftso BTC

# View full chain status
./lfts status
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
  "timestamp": 1710000000,
  "blockNum": 42
}
```

**Get price at specific timestamp:**
```bash
curl "http://localhost:9650/ftso/price?asset=BTC&timestamp=1710000000"
```

### GET /ftso/history?asset=BTC

Returns price history for the specified asset.

**Query options:**
- `limit=N`: Get last N prices
- `from=<timestamp>&to=<timestamp>`: Get prices in time range
- No parameters: Get full history

**Response:**
```json
{
  "asset": "BTC",
  "latest": {
    "asset": "BTC",
    "price": 65000.0,
    "timestamp": 1710000000,
    "blockNum": 42
  },
  "history": [
    {
      "price": 65000.0,
      "timestamp": 1710000000,
      "blockNum": 42
    },
    {
      "price": 64900.0,
      "timestamp": 1709999000,
      "blockNum": 41
    }
  ]
}
```

**Examples:**
```bash
# Get last 10 prices
curl "http://localhost:9650/ftso/history?asset=BTC&limit=10"

# Get prices in time range
curl "http://localhost:9650/ftso/history?asset=BTC&from=1709990000&to=1710000000"
```

### GET /ftso/prices

Returns all current FTSO prices.

**Response:**
```json
{
  "prices": {
    "BTC": {
      "asset": "BTC",
      "price": 65000.0,
      "timestamp": 1710000000
    },
    "ETH": {
      "asset": "ETH",
      "price": 3500.0,
      "timestamp": 1710000000
    }
  }
}
```

### POST /ftso/inject?asset=BTC&price=65000

Injects a new FTSO price (can also be done via CLI).

### GET /fdc/feed?name=weather

Returns the latest FDC feed data.

**Response:**
```json
{
  "feedName": "weather",
  "data": {
    "temp": 25,
    "humidity": 60,
    "condition": "sunny"
  },
  "timestamp": 1710000000,
  "blockNum": 42
}
```

### POST /fdc/inject?name=weather

Injects new FDC feed data. Send JSON in request body.

**Example:**
```bash
curl -X POST "http://localhost:9650/fdc/inject?name=weather" \
  -H "Content-Type: application/json" \
  -d '{"temp":25,"humidity":60}'
```

### GET /fdc/history?name=weather&limit=10

Returns FDC feed history (similar to FTSO history).

### GET /fdc/list

Returns all available FDC feeds.

**Response:**
```json
{
  "feeds": {
    "weather": {
      "feedName": "weather",
      "data": {...},
      "timestamp": 1710000000
    }
  }
}
```

### POST /rpc

JSON-RPC endpoint for smart contract calls.

**Supported methods:**
- `eth_call`: Execute contract calls
- `eth_blockNumber`: Get current block number
- `eth_getBlockByNumber`: Get block information

**Example - Call FTSO contract:**
```bash
curl -X POST http://localhost:9650/rpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "eth_call",
    "params": [{
      "to": "0x0000000000000000000000000000000000000001",
      "data": "0x893d20e80000000000000000000000000000000000000000000000000000000000000001"
    }, "latest"]
  }'
```

**Mock Contract Addresses:**
- FTSO Contract: `0x0000000000000000000000000000000000000001`
- FDC Contract: `0x0000000000000000000000000000000000000002`

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

# Inject FDC feed
docker exec lfts-sandbox ./lfts inject fdc weather '{"temp":25}'

# Check status
docker exec lfts-sandbox ./lfts status

# View history
docker exec lfts-sandbox ./lfts history ftso BTC
```

### Stop the Container

```bash
docker stop lfts-sandbox
docker rm lfts-sandbox
```

## Example Workflow

1. **Start the chain with auto-updates:**
   ```bash
   ./lfts start --auto-update-ftso --update-pattern random --update-assets BTC,ETH
   ```

2. **In another terminal, inject some prices manually:**
   ```bash
   ./lfts inject ftso BTC 65000
   ./lfts inject ftso ETH 3500
   ./lfts inject ftso XRP 0.5
   ```

3. **Inject FDC feeds:**
   ```bash
   ./lfts inject fdc weather '{"temp":25,"humidity":60}'
   ./lfts inject fdc sports '{"game":"Lakers vs Warriors","score":"98-95"}'
   ```

4. **Query prices and feeds via RPC:**
   ```bash
   curl http://localhost:9650/ftso/price?asset=BTC
   curl http://localhost:9650/ftso/history?asset=BTC&limit=10
   curl http://localhost:9650/fdc/feed?name=weather
   ```

5. **View history:**
   ```bash
   ./lfts history ftso BTC
   ./lfts status
   ```

6. **Test smart contract calls:**
   ```bash
   curl -X POST http://localhost:9650/rpc \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"eth_call","params":[{"to":"0x0000000000000000000000000000000000000001","data":"0x893d20e80000000000000000000000000000000000000000000000000000000000000001"},"latest"]}'
   ```

## CLI Commands Reference

### Chain Management
- `lfts start [flags]` - Start the chain node
- `lfts stop` - Stop the running chain
- `lfts status` - Show chain status and prices

### FTSO Commands
- `lfts inject ftso <asset> <price>` - Inject a price
- `lfts history ftso <asset>` - Show price history

### FDC Commands
- `lfts inject fdc <feed_name> <json_data>` - Inject FDC feed data
- `lfts query fdc <feed_name>` - Query FDC feed
- `lfts list fdc` - List all FDC feeds

### Start Command Flags
- `--block-time <ms>` - Block generation interval (default: 1000ms)
- `--port <port>` - RPC server port (default: 9650)
- `--auto-update-ftso` - Enable automatic price updates
- `--update-interval <ms>` - Auto-update interval (default: 1800ms)
- `--update-pattern <pattern>` - Update pattern: random, sine, crash, spike, stable
- `--update-assets <assets>` - Comma-separated assets to update
- `--volatility <percent>` - Price volatility percentage (default: 1.0%)


## Design Notes

- **In-Memory Storage**: State is stored in memory (no persistence). This can be extended with LevelDB or similar.
- **Price History**: Maintains up to 1000 historical entries per asset/feed for testing time-series queries.
- **Thread-Safe**: All state operations use mutexes for concurrent access safety.
- **Simple Architecture**: Minimal dependencies, easy to understand and modify.
- **Extensible**: Code structure allows for easy addition of features like persistence, more RPC endpoints, or additional oracle types.
- **Call Simulation**: Smart contract testing uses call simulation (not full EVM) for fast, lightweight testing.

## Auto-Update Patterns

- **random**: Random walk - prices change by random percentage (configurable volatility)
- **sine**: Sine wave - oscillating prices for testing periodic patterns
- **crash**: Market crash scenario - gradual decline then recovery
- **spike**: Price spike scenario - sudden increase then gradual decline
- **stable**: Stable prices - minimal random changes (±0.05%)

## License

This is a minimal MVP for development and testing purposes.

