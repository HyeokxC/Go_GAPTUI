# KimchiCEX Arbitrage Bot — System Specification

## 1. System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    KIMCHI PREMIUM MONITOR                       │
├─────────────────────────────────────────────────────────────────┤
│  [1] Gap Monitor  →  [2] Executor  →  [3] Transfer  →  [4] Closer │
│      (Done)           (Planned)        (Done)           (Planned)  │
└─────────────────────────────────────────────────────────────────┘
```

**Language**: Rust (native macOS GUI with egui/eframe)

**Supported Exchanges** (7):
- Korean: Upbit, Bithumb
- Global: Binance, Bybit, Bitget, OKX, Gate.io

---

## 2. Architecture

### Data Flow

```
                        ┌──────────────────────────────────────┐
                        │         WebSocket Streams            │
                        │  (TickerWsManager — all 7 exchanges) │
                        └──────────────┬───────────────────────┘
                                       │
                        ┌──────────────▼───────────────────────┐
                        │      Shared Arc<Mutex<...>> Stores   │
                        │  coin_states, wallet_status, logs,   │
                        │  orderbook_info, scenario_threads    │
                        └──────────────┬───────────────────────┘
                                       │ 200ms
                        ┌──────────────▼───────────────────────┐
                        │     Snapshot Publisher (background)   │
                        │  Reads all stores → builds immutable │
                        │  Arc<AppSnapshot> → swaps into       │
                        │  SharedSnapshot                      │
                        └──────────────┬───────────────────────┘
                                       │ Arc::clone (O(1))
                        ┌──────────────▼───────────────────────┐
                        │     egui UI (100ms repaint)          │
                        │  get_snapshot() once per frame       │
                        │  Virtual scrolling, cached sorts,    │
                        │  pre-formatted strings               │
                        └──────────────────────────────────────┘
```

### Snapshot Swap Pattern

The key architectural pattern is the snapshot swap, which decouples background data writes from UI reads:

1. **Background writers** update shared `Arc<Mutex<HashMap<...>>>` stores independently
2. **Snapshot publisher** (200ms interval) locks each store once, clones into per-field `Arc<>` wrappers, builds `AppSnapshot`, and swaps it into `SharedSnapshot`
3. **UI thread** calls `get_snapshot()` once per frame — a single `Arc::clone()` with zero contention

```rust
// shared.rs
pub struct AppSnapshot {
    pub coin_states: Arc<HashMap<String, CoinState>>,
    pub wallet_status: Arc<HashMap<String, CoinWalletStatus>>,
    pub korean_names: Arc<HashMap<String, String>>,
    pub logs: Arc<Vec<LogEntry>>,
    pub orderbook_info: Arc<HashMap<String, OrderbookInfo>>,
    pub scenario_threads: Arc<Vec<LogThread>>,
    pub usdt_krw: Option<f64>,
    pub usd_krw_forex: Option<f64>,
}

pub type SharedSnapshot = Arc<RwLock<Arc<AppSnapshot>>>;
```

---

## 3. Module Specifications

### 3.1 Gap Monitor Module (Implemented)

**Purpose**: Real-time price monitoring and gap detection across all 7 exchanges.

**Data Sources** — All via WebSocket through unified `TickerWsManager`:
```
┌──────────────────────────────────────┐
│          PRICE MONITOR               │
├──────────────────────────────────────┤
│  Bithumb WebSocket   ──┐             │
│  Upbit WebSocket     ──┤             │
│  Binance WebSocket   ──┼─→ Gap Calc  │
│  Bybit WebSocket     ──┤             │
│  Bitget WebSocket    ──┤             │
│  OKX WebSocket       ──┤             │
│  Gate WebSocket      ──┘             │
│                                      │
│  + Orderbook WebSocket (slippage)    │
│  + Wallet Status REST (D/W polling)  │
│                                      │
│  Gap ≥ threshold → Scenario Thread   │
└──────────────────────────────────────┘
```

**Features**:
- WebSocket connection management with auto-reconnect
- USDT/KRW conversion using Upbit USDT market
- USD/KRW forex rate for reference
- Orderbook-based slippage calculation (WebSocket)
- Scenario detection (3 active types: GapThreshold, DomesticGap, FutBasis)
- D/W wallet status polling (Bithumb, Upbit, Binance, Bybit, OKX) with per-chain blocked info
- D/W blocked status panel (real-time view of currently blocked coins/exchanges/chains)
- Auto-discover common symbols across exchanges
- SQLite logging for D/W events and session tracking

### 3.2 Executor Module (Planned)

**Purpose**: Simultaneous execution of spot buy and futures short.

**Execution Logic**:
- Parallel execution: Buy on domestic exchange + Short on offshore exchange
- Quantity matching validation
- Rollback logic for partial fills

**Failure Handling**:
| Scenario | Action |
|----------|--------|
| Buy succeeds, Short fails | Immediately sell on domestic |
| Buy fails, Short succeeds | Immediately close short |
| Slippage exceeds threshold | Abort and close all |
| Both fail | Log and retry once |

### 3.3 Transfer Module (Implemented)

**Purpose**: Cross-exchange coin transfer with real-time progress tracking.

**Supported transfers**: Any exchange → Any exchange (7 exchanges)

**Architecture**:
```
┌─────────────────────────────────┐
│        TRANSFER MODULE          │
├─────────────────────────────────┤
│  1. Fetch common networks       │
│  2. Fetch balance + deposit addr│
│  3. Submit withdrawal           │
│  4. Poll withdrawal status      │
│  5. Poll deposit status         │
│  6. Transfer complete           │
└─────────────────────────────────┘
```

**Features**:
- Cross-exchange network name mapping (21 networks)
- 6-step progress tracker with real-time updates
- Editable deposit address/tag fields
- Balance check + minimum withdrawal validation
- Confirmation dialog before execution
- TX hash display with copy/link
- 30-minute timeout warning
- Separate egui viewport window

**Network mapping examples**:
| Network | Binance | Upbit | Bithumb | Bybit | Bitget | OKX | Gate |
|---------|---------|-------|---------|-------|--------|-----|------|
| Ethereum | ETH | ETH | ETH | ETH | ETH | ETH-ERC20 | ETH |
| BSC | BSC | BSC | BNB | BSC | BSC | BSC-BEP20 | BSC_BEP20 |
| Tron | TRX | TRX | TRX | TRX | TRC20 | TRON-TRC20 | TRC20 |
| Solana | SOL | SOL | SOL | SOL | SOL | SOL-Solana | SOL |

**Auth patterns per exchange**:
| Exchange | Method | Signature | Extra |
|----------|--------|-----------|-------|
| Binance | HMAC-SHA256 | query string → hex | Header: X-MBX-APIKEY |
| Upbit | JWT (HMAC-SHA256) | query_hash (SHA512) in payload | Bearer token |
| Bithumb | HMAC-SHA512 | endpoint\0params\0timestamp → base64 | Api-Key, Api-Sign, Api-Nonce (UUID) |
| Bybit | HMAC-SHA256 | timestamp+apikey+recvWindow+payload → hex | X-BAPI-* headers |
| Bitget | HMAC-SHA256 → base64 | timestamp+method+path+body | ACCESS-* headers, passphrase required |
| OKX | HMAC-SHA256 → base64 | ISO_timestamp+method+path+body | OK-ACCESS-* headers, passphrase required |
| Gate | HMAC-SHA512 | method\npath\nquery\nsha512(body)\ntimestamp | KEY, SIGN, Timestamp headers |

### 3.4 Closer Module (Planned)

**Purpose**: Close positions when transfer completes (sell on destination + close hedge).

---

## 4. Configuration

```toml
# config.toml
[exchange]
request_timeout_secs = 10

[dashboard]
refresh_ms = 100              # UI repaint interval
stale_display_secs = 30

[scenario]
gap_threshold_percent = 5.0
gap_sub_thread_change = 3.0
domestic_gap_threshold = 1.5
fut_basis_threshold = 0.5
```

**API Keys**: Loaded from `.env` file (see `.env.example`)

---

## 5. API Requirements

### Per-Exchange Transfer APIs

| Method | Description |
|--------|-------------|
| `fetch_networks(coin)` | Get available networks with D/W status |
| `fetch_balance(coin)` | Get available balance |
| `fetch_deposit_address(coin, network)` | Get deposit address + tag |
| `submit_withdrawal(coin, network, amount, address, memo)` | Submit withdrawal |
| `check_withdrawal(coin, id)` | Poll withdrawal status |
| `check_deposit(coin, txid)` | Poll deposit status |

### Exchange API Endpoints

| Exchange | Networks | Balance | Deposit Addr | Withdraw | W-Status | D-Status |
|----------|----------|---------|-------------|----------|----------|----------|
| Binance | /sapi/v1/capital/config/getall | /api/v3/account | /sapi/v1/capital/deposit/address | /sapi/v1/capital/withdraw/apply | /sapi/v1/capital/withdraw/history | /sapi/v1/capital/deposit/hisrec |
| Upbit | /v1/status/wallet | /v1/accounts | /v1/deposits/coin_address | /v1/withdraws/coin | /v1/withdraw?uuid= | /v1/deposits |
| Bithumb | /info/wallet_address | /info/balance | /info/wallet_address | /trade/btc_withdrawal | /info/user_transactions | /info/user_transactions |
| Bybit | /v5/asset/coin/query-info | /v5/account/wallet-balance | /v5/asset/deposit/query-address | /v5/asset/withdraw/create | /v5/asset/withdraw/query-record | /v5/asset/deposit/query-record |
| Bitget | /api/v2/spot/public/coins | /api/v2/spot/account/assets | /api/v2/spot/wallet/deposit-address | /api/v2/spot/wallet/withdrawal | /api/v2/spot/wallet/withdrawal-records | /api/v2/spot/wallet/deposit-records |
| OKX | /api/v5/asset/currencies | /api/v5/asset/balances | /api/v5/asset/deposit-address | /api/v5/asset/withdrawal | /api/v5/asset/withdrawal-history | /api/v5/asset/deposit-history |
| Gate | /api/v4/spot/currencies/{coin} | /api/v4/spot/accounts | /api/v4/wallet/deposit_address | /api/v4/withdrawals | /api/v4/wallet/withdrawals/{id} | /api/v4/wallet/deposits |

---

## 6. Project Structure

```
kimchiCEX_arbitrage/
├── claude/                        # Documentation
│   ├── claude.md                  # Project summary
│   ├── spec.md                    # This file
│   ├── task.md                    # Phase task tracking
│   ├── scenario.md                # Scenario detection specs
│   ├── orderbook_slippage.md      # Orderbook slippage specs
│   ├── Gap_Strategy.md            # Strategy reference
│   ├── performance_improvements.md # Performance optimizations
│   └── ui_improvements.md         # UI redesign details
├── src/
│   ├── main.rs                    # Entry point, wiring, snapshot + watch setup
│   ├── app.rs                     # egui UI rendering (~1800 lines)
│   ├── shared.rs                  # AppSnapshot, SharedSnapshot, get_snapshot()
│   ├── background.rs              # Background tasks, snapshot publisher
│   ├── config.rs                  # AppConfig, ApiKeysConfig, ProxyConfig
│   ├── db.rs                      # SQLite: LogEntry, session logging
│   ├── http_client.rs             # HTTP client builder with proxy
│   ├── scenario.rs                # ScenarioDetector (3 types)
│   ├── models/
│   │   ├── mod.rs                 # Exchange enum (7 variants)
│   │   └── price.rs               # CoinState, CoinWalletStatus, OrderbookInfo
│   ├── exchanges/
│   │   ├── mod.rs
│   │   ├── ticker_ws.rs           # TickerWsManager — WebSocket streams (all 7)
│   │   ├── orderbook.rs           # Orderbook WebSocket + slippage calculation
│   │   ├── wallet_status.rs       # D/W status fetcher (REST polling — 5 exchanges)
│   │   ├── symbol_fetcher.rs      # Auto-discover common symbols
│   │   ├── transfer_api.rs        # Transfer APIs (all 7 exchanges)
│   │   └── ticker.rs              # Legacy REST ticker (unused)
│   ├── monitor/
│   │   └── kimchi_monitor.rs      # Kimchi premium calculation
│   └── transfer/
│       ├── mod.rs
│       ├── state.rs               # TransferState, TransferStep, commands
│       ├── executor.rs            # Async state machine
│       └── window.rs              # egui viewport UI
├── .env.example                   # Environment variable template
├── config.toml                    # App configuration
└── Cargo.toml                     # Dependencies
```

---

## 7. Risk Management

### Pre-Trade Checks
- Balance >= trade amount
- Withdrawal enabled on source exchange
- Deposit enabled on target exchange
- Common network available
- Amount >= minimum withdrawal
- Confirmation dialog before execution

### Circuit Breakers
| Condition | Action |
|-----------|--------|
| Transfer stuck > 30 min | Warning displayed |
| No progress > 60 min | Manual intervention required |

---

## 8. Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Rust |
| Async Runtime | tokio |
| GUI | egui/eframe 0.31 |
| HTTP | reqwest (rustls-tls) |
| WebSocket | tokio-tungstenite (native-tls) |
| Synchronization | parking_lot (Mutex, RwLock) |
| Auth | hmac, sha2, base64, uuid |
| Config | toml, dotenvy |
| Logging | tracing, tracing-subscriber |
| Database | rusqlite (bundled SQLite) |
| Precision | rust_decimal |
| Time | chrono |
