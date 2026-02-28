# Arbitrage Scenario Specification

## Overview

Thread-based event detection system. When a scenario triggers, a **LogThread** is created. Subsequent events for the same coin append as sub-entries. Threads can be collapsed/expanded in the UI (shows 5-6 entries when expanded, scrollable).

---

## Implementation Status

| Scenario | ScenarioType | Status |
|----------|-------------|--------|
| 1: Gap Threshold Crossing | `GapThreshold` | **Implemented** |
| 2: Domestic Gap Alert | `DomesticGap` | **Implemented** |
| 3: Futures Basis | `FutBasis` | **Implemented** |
| 4: D/W + Gap Combined | — | **Not Implemented** (planned) |
| 5: Rapid Change Detection | — | **Not Implemented** (disabled in original spec) |
| 6: Orderbook Slippage Scenarios | — | **Not Implemented** as scenario threads (data IS calculated and displayed in UI) |

---

## Scenario 1: Gap Threshold Crossing (Implemented)

**Purpose**: Detect when a coin's Kimchi premium significantly exceeds the USDT premium baseline across **all three overseas exchanges** (Binance, Bybit, OKX).

**Trigger**:
- Coin's Kimchi premium (Upbit or Bithumb vs any overseas exchange) >= configured threshold `n%` (default: 5%)
- Uses depth-based kimp from `OrderbookInfo.real_kimp_up/bt` HashMap, with fallback to mid-price kimp
- Checked independently for each overseas exchange

**Per-exchange thread keys**:
| Key | Meaning |
|-----|---------|
| `UP-BN` | Upbit vs Binance |
| `UP-BB` | Upbit vs Bybit |
| `UP-OK` | Upbit vs OKX |
| `BT-BN` | Bithumb vs Binance |
| `BT-BB` | Bithumb vs Bybit |
| `BT-OK` | Bithumb vs OKX |

**Thread behavior**:
- Threshold crossed → main thread created (per exchange key)
- Subsequent change >= `gap_sub_thread_change` (default: 3.0%p) → sub-entry added
- Drops below threshold → thread deactivated (closed)

**Tracked per (coin, key)**: `was_above_gap: HashMap<String, bool>` keyed by `"SYMBOL:KEY"` (e.g. `"BTC:UP-BN"`, `"ETH:BT-OK"`)

**API signature**: `on_price_update(symbol, kimp_by_exchange: &[(&str, f64, f64)], ...)` where each tuple is `(exchange_code, up_kimp, bt_kimp)`

**Wallet filtering**: Key is parsed as `DOMESTIC-OVERSEAS` — domestic part (UP/BT) determines deposit check, overseas suffix (BN/BB/OK) determines withdraw check

**Log example**:
```
▶ [KIMP] XRP UP-BN kimp crossed ▲5.0% → 5.3%
  └ UP-BN kimp 8.4% (+3.1%p)
  └ UP-BN kimp ▼4.5% (below 5.0%) — closed
```

---

## Scenario 2: Domestic Gap Alert (Implemented)

**Purpose**: Detect Upbit-Bithumb price difference (Bithumb buy → Upbit sell opportunity).

**Trigger**:
- `domestic_gap = (upbit_price - bithumb_price) / bithumb_price × 100`
- Absolute value >= configured threshold (default: 1.5%)
- Detects both positive gap (UP > BT) and negative gap (BT > UP)

**Thread behavior**:
- Threshold crossed → main thread created
- Gap changes significantly → sub-entry added
- Drops below threshold → thread deactivated

**Tracked per coin**: `was_above_domestic: HashMap<String, (bool, bool)>` tracking `(above_positive, below_negative)`

**Log example**:
```
▶ [DOM-GAP] XRP UP-BT gap +2.5% (UP ₩3,200 > BT ₩3,120)
  └ Gap widened +3.1% (UP ₩3,250 > BT ₩3,150)
  └ Gap narrowed +1.2% — closed
```

---

## Scenario 3: Futures Basis (Implemented)

**Purpose**: Detect when spot-futures basis crosses a threshold, indicating potential convergence trading opportunity.

**Trigger**:
- `fut_basis = (futures_price - spot_price) / spot_price × 100`
- Absolute value >= configured threshold (default: 0.5%)
- Detects both contango (positive) and backwardation (negative)

**Thread behavior**:
- Threshold crossed → main thread created (with `key` distinguishing positive/negative)
- Basis changes significantly → sub-entry added
- Drops below threshold → thread deactivated

**Tracked per coin**: `was_above_fut_basis: HashMap<String, (bool, bool)>` tracking `(above_positive, below_negative)`

**Note**: This scenario was added during implementation but was not in the original specification. It enables detection of convergence trading opportunities described in `Gap_Strategy.md` (strategies 3-A, 3-B, 3-C).

**Log example**:
```
▶ [FUT%] BTC basis +1.2% (contango)
  └ Basis widened +1.8%
  └ Basis narrowed +0.3% — closed
```

---

## Planned Scenarios (Not Yet Implemented)

### Scenario 4: D/W + Gap Combined Signal

**Purpose**: Combine deposit/withdrawal status changes with Kimchi premium for composite signals.

**4-A: D/W Opens + High Premium (Golden Opportunity)**
- Trigger: Exchange D/W changes from BLOCKED → OPEN while coin's premium >= threshold
- Indicates a real arbitrage opportunity has opened up

**4-B: D/W Blocked → Internal Pump Watch**
- Trigger: D/W changes from OPEN → BLOCKED
- Track subsequent premium changes (price pumps while D/W blocked are likely fake — coins can't leave the exchange)

**Current state**: D/W wallet status changes ARE logged as `LogEntry` events in the D/W Event Log panel, but are NOT integrated as scenario detection threads with combined gap signals.

### Scenario 5: Rapid Change Detection

**Purpose**: Detect rapid Kimchi premium spikes using rolling window rate-of-change.

**Method**: Multi-timeframe detection (10s/30s/60s/5min windows) with cooldown for duplicate suppression.

| Window | Threshold | Meaning |
|--------|-----------|---------|
| 10s | ±1.5%p | Flash spike |
| 30s | ±2.0%p | Rapid change |
| 60s | ±3.0%p | Fast trend shift |
| 5min | ±5.0%p | Sustained large move |

**Current state**: **Disabled** in original spec (`spike_enabled = false`). Not implemented in `ScenarioDetector`.

### Scenario 6: Orderbook Slippage Scenarios

**Purpose**: Warn when surface gap doesn't reflect reality due to thin orderbooks.

**6-A: Thin Orderbook Warning** — Slippage >= 0.5% on 2M KRW order
**6-B: Real Gap Alert** — Surface gap >= 2% AND real gap >= 1.5%
**6-C: Fake Gap Warning** — Surface gap >= 3% BUT real gap < 1%

**Current state**: Orderbook data IS collected via WebSocket (`exchanges/orderbook.rs`), slippage/spread IS calculated in `background.rs::realtime_processor`, and values ARE displayed in the coin card UI. However, these are NOT integrated as scenario detection threads in `ScenarioDetector`. See `orderbook_slippage.md` for details.

---

## Configuration

```toml
[scenario]
# Scenario 1: Gap Threshold
gap_threshold_percent = 5.0          # Trigger when premium >= n%
gap_sub_thread_change = 3.0          # Sub-entry threshold (%p change)

# Scenario 2: Domestic Gap
domestic_gap_threshold = 1.5         # UP-BT gap % (absolute, detects both directions)

# Scenario 3: Futures Basis
fut_basis_threshold = 0.5            # Spot-futures basis threshold %
```

---

## Data Model (Actual Implementation)

```rust
#[derive(Clone, Copy, PartialEq, Eq, Debug, Hash)]
pub enum ScenarioType {
    GapThreshold,   // Scenario 1: Kimchi premium crossing
    DomesticGap,    // Scenario 2: UP-BT domestic gap
    FutBasis,       // Scenario 3: Futures basis crossing
}

#[derive(Clone)]
pub struct LogThread {
    pub id: u64,
    pub symbol: String,
    pub scenario: ScenarioType,
    pub key: String,               // Sub-key: "UP-BN", "BT-OK", "pos"/"neg", etc.
    pub main_message: String,
    pub main_timestamp: DateTime<Local>,
    pub sub_entries: Vec<ThreadEntry>,
    pub is_active: bool,
    pub closed_at: Option<DateTime<Local>>,
    pub initial_value: f64,        // Value at thread creation
    pub last_logged_value: f64,    // Last value that triggered sub-entry
}

#[derive(Clone)]
pub struct ThreadEntry {
    pub timestamp: DateTime<Local>,
    pub message: String,
}
```

---

## Thread UI Behavior

- Main entry only shown by default (collapsed)
- Click to expand → shows up to 5-6 sub-entries
- Scrollable when sub-entries exceed 6
- Active threads: 3px left accent bar in scenario color
- Inactive (closed) threads: dimmed styling
- Pre-sorted in background: active first, newest first
