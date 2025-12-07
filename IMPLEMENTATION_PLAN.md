# LFTS Implementation Plan

## Bug Fix: Status Command

**Problem**: `lfts status` shows "Chain is not initialized" when run in a separate process because it tries to access a global chain instance that only exists in the `lfts start` process.

**Solution**: Make `status` command query via RPC (HTTP GET to `/status` endpoint) instead of accessing global state directly. Docker won't fix this - it's a process isolation issue.

**Files to modify**:
- `cmd/lfts/commands.go` - Update `runStatus()` to use HTTP client

---

## Feature 1: FDC (Flare Data Connector) Tester/Simulator

### What is FDC?
**Question needed**: Is FDC a separate system from FTSO? Based on naming, I'm assuming:
- FDC = Flare Data Connector - allows external data sources to feed arbitrary data (not just prices) into Flare
- FTSO = Flare Time Series Oracle - specifically for price/time-series data
- Both are oracles but serve different purposes

**Assumptions** (please confirm):
- FDC handles arbitrary data feeds (not just prices)
- FDC has similar injection/query patterns as FTSO
- FDC data is keyed differently (maybe `fdc:<feed_name>` instead of `ftso:<asset>`)

### Implementation Plan:

#### 1.1 Create FDC Package Structure
```
internal/
├── fdc/
│   ├── fdc.go          # FDC data feed management
│   ├── handler.go      # HTTP handlers
│   └── feed.go         # Feed data structures
```

#### 1.2 FDC Data Structure
```go
type FDCFeed struct {
    FeedName  string                 `json:"feedName"`
    Data      map[string]interface{} `json:"data"`  // Arbitrary JSON data
    Timestamp int64                   `json:"timestamp"`
    BlockNum  uint64                 `json:"blockNum"`
}
```

#### 1.3 FDC Storage
- Store feeds under key: `fdc:<feed_name>`
- Support multiple feed types (prices, weather, sports scores, etc.)
- Maintain history similar to FTSO

#### 1.4 CLI Commands
- `lfts inject fdc <feed_name> <json_data>` - Inject FDC data
- `lfts query fdc <feed_name>` - Query FDC feed
- `lfts list fdc` - List all FDC feeds

#### 1.5 RPC Endpoints
- `GET /fdc/feed?name=<feed_name>` - Get latest feed data
- `POST /fdc/inject?name=<feed_name>` - Inject feed data (with JSON body)
- `GET /fdc/list` - List all available feeds

---

## Feature 2: Price History

### Current Limitation
Only the latest price is stored. Injecting a new price overwrites the old one.

### Implementation Plan:

#### 2.1 Modify FTSO Storage Structure
Change from single price storage to time-series storage:
- Current: `ftso:BTC` → single price
- New: `ftso:BTC:history` → array of prices with timestamps
- Keep: `ftso:BTC:latest` → latest price (for fast access)

#### 2.2 Price History Data Structure
```go
type FTSOPriceHistory struct {
    Asset     string       `json:"asset"`
    Prices    []PricePoint `json:"prices"`
    Latest    *FTSOPrice   `json:"latest"`
}

type PricePoint struct {
    Price     float64 `json:"price"`
    Timestamp int64   `json:"timestamp"`
    BlockNum  uint64  `json:"blockNum"`
}
```

#### 2.3 New RPC Endpoints
- `GET /ftso/price?asset=BTC&timestamp=<ts>` - Get price at specific timestamp
- `GET /ftso/history?asset=BTC&from=<ts>&to=<ts>` - Get price history range
- `GET /ftso/history?asset=BTC&limit=100` - Get last N prices

#### 2.4 Update CLI
- `lfts status` - Show price history summary (last 5 prices per asset)
- `lfts history ftso BTC` - Show full price history for asset

#### 2.5 Storage Strategy
- Store history in append-only format
- Limit history size (configurable, default: 1000 entries per asset)
- Option to persist to disk for larger history

---

## Feature 3: Auto-Update Simulation Flag

### Purpose
Simulate automatic price updates like real FTSO (updates every ~1.8 seconds).

### Implementation Plan:

#### 3.1 Add Flag to Start Command
```bash
lfts start --auto-update-ftso
lfts start --auto-update-ftso --update-interval 1800  # milliseconds
```

#### 3.2 Price Update Patterns
Implement different simulation patterns:
- **Random Walk**: Prices change by small random amounts
- **Sine Wave**: Oscillating prices (for testing)
- **Crash Scenario**: Sudden price drops
- **Spike Scenario**: Sudden price increases
- **Stable**: Minimal changes

#### 3.3 Configuration
```go
type AutoUpdateConfig struct {
    Enabled    bool
    Interval   time.Duration
    Pattern    string  // "random", "sine", "crash", "spike", "stable"
    Assets     []string
    BasePrices map[string]float64
    Volatility float64  // For random walk
}
```

#### 3.4 Background Goroutine
- Run in separate goroutine when flag is enabled
- Update prices at specified interval
- Log updates for debugging
- Respect chain stop signal

#### 3.5 CLI Commands
- `lfts start --auto-update-ftso --pattern random --assets BTC,ETH`
- `lfts scenario crash BTC` - Simulate market crash for asset

---

## Feature 4: Smart Contract Testing with FTSO and FDC

### Purpose
Allow developers to test Solidity contracts that interact with FTSO/FDC oracles.

### Implementation Plan:

#### 4.1 JSON-RPC Implementation
Implement standard Ethereum JSON-RPC methods:
- `eth_call` - Execute contract calls (read-only)
- `eth_sendTransaction` - Send transactions (for testing)
- `eth_getBlockByNumber` - Get block info
- `eth_blockNumber` - Get current block
- `eth_getCode` - Get contract code (for deployed contracts)

#### 4.2 Mock FTSO Contract Interface
Create a mock contract that mimics Flare's FTSO contract:
```solidity
interface IFTSO {
    function getCurrentPrice(address asset) external view returns (uint256 price, uint256 timestamp);
    function getPrice(address asset, uint256 epoch) external view returns (uint256 price);
    function getPriceAt(address asset, uint256 timestamp) external view returns (uint256 price);
}
```

#### 4.3 Mock FDC Contract Interface
Create a mock contract for FDC:
```solidity
interface IFDC {
    function getData(string memory feedName) external view returns (bytes memory data);
    function getDataAt(string memory feedName, uint256 timestamp) external view returns (bytes memory data);
}
```

#### 4.4 Contract State Management
- Store deployed contracts in state
- Map contract addresses to their code/ABI
- Execute contract calls against our FTSO/FDC state

#### 4.5 Contract Deployment Simulation
- `POST /rpc` - JSON-RPC endpoint
- `eth_sendTransaction` with contract creation code
- Store contract at generated address
- Return transaction receipt

#### 4.6 Contract Call Execution
- Parse `eth_call` requests
- Decode function calls (using ABI)
- Route FTSO/FDC calls to our internal state
- Return encoded results

#### 4.7 Testing Utilities
- Example test contracts (Solidity files)
- Scripts to deploy and test contracts
- Integration with Hardhat/Foundry (optional)

---

## Implementation Order & Dependencies

### Phase 1: Bug Fix (Immediate)
1. Fix status command bug
2. Test and verify

### Phase 2: Foundation (Price History)
1. Implement price history storage
2. Add history endpoints
3. Update CLI to show history
4. **Dependency**: Needed for auto-update to work properly

### Phase 3: FDC Module
1. Create FDC package structure
2. Implement FDC storage and retrieval
3. Add FDC CLI commands
4. Add FDC RPC endpoints
5. **Dependency**: Can be done in parallel with Phase 2

### Phase 4: Auto-Update Simulation
1. Add auto-update flag and config
2. Implement price update patterns
3. Add background update goroutine
4. Add scenario commands
5. **Dependency**: Requires Phase 2 (price history)

### Phase 5: Smart Contract Testing
1. Implement JSON-RPC endpoint
2. Create mock FTSO contract interface
3. Create mock FDC contract interface
4. Implement contract state management
5. Add contract deployment simulation
6. Add contract call execution
7. **Dependency**: Requires Phase 2 and Phase 3

---

## Questions for Clarification

1. **FDC Details**:
   - What types of data does FDC handle? (prices, weather, sports, etc.)
   - Is FDC structure similar to FTSO or completely different?
   - Should FDC also maintain history like FTSO?

2. **Price History**:
   - How much history should we keep? (1000 entries? Unlimited?)
   - Should history be persisted to disk or stay in-memory?
   - Do we need to support querying by block number?

3. **Auto-Update**:
   - Should auto-update simulate multiple data providers (like real FTSO)?
   - Do we need weighted median aggregation?
   - Should updates be tied to block production?

4. **Smart Contracts**:
   - Do we need full EVM execution or just contract call simulation?
   - Should we support contract deployment or just read-only calls?
   - Do we need to support events/logs?

5. **Priority**:
   - Which features are most important for your use case?
   - Should we implement all features or focus on specific ones first?

---

## Estimated Complexity

- **Bug Fix**: 30 minutes
- **Price History**: 2-3 hours
- **FDC Module**: 3-4 hours
- **Auto-Update**: 2-3 hours
- **Smart Contract Testing**: 4-6 hours (basic) to 8-12 hours (full EVM)

**Total**: ~12-20 hours for full implementation

