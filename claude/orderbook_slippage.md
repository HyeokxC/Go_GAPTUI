# Orderbook & Slippage Specification

## Problem

Kimchi premium calculation based on **last traded price** doesn't reflect actual execution cost. When orderbooks are thin, market-order slippage can be significant — a surface gap of 10% might only yield 3.9% in practice.

**Example**: XYZ coin
- Last price: ₩100 (Bithumb), ₩110 (Upbit) → surface gap 10%
- Bithumb asks: 100×50, 102×30, 105×20 (thin)
- Upbit bids: 110×50, 108×30, 105×20 (thin)
- 2M KRW market buy avg: ₩103, market sell avg: ₩107
- **Real gap**: (107-103)/103 = 3.9% (vs 10% surface)

---

## Implementation Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Orderbook collection + spread display | **Implemented** (via WebSocket) |
| Phase 2 | Slippage calculation + real gap | **Implemented** |
| Phase 3 | Scenario thread integration (5-A/5-B/5-C) | **Not Implemented** |

**Note**: Phases 1 & 2 were implemented using **WebSocket** orderbook streams (not REST polling as originally planned), which provides lower latency. The data is collected in `exchanges/orderbook.rs`, slippage is calculated in `background.rs::realtime_processor`, and results are displayed in the coin card UI. However, Phase 3 (integrating slippage alerts as `ScenarioDetector` threads) has not been implemented yet.

---

## Solution Overview

### 1. Orderbook Data Collection (Implemented)

Orderbook data is collected via **WebSocket** streams through `exchanges/orderbook.rs`.

**Data model:**
```rust
struct OrderbookEntry {
    price: f64,
    quantity: f64,
}

struct Orderbook {
    symbol: String,
    exchange: Exchange,
    asks: Vec<OrderbookEntry>,  // Sell orders (low → high)
    bids: Vec<OrderbookEntry>,  // Buy orders (high → low)
    timestamp: DateTime<Utc>,
}
```

### 2. Slippage Calculation (Implemented)

For a given order size (default: 2M KRW), calculates the actual average execution price.

**Buy slippage** (consuming asks):
```
fn calc_buy_avg(asks, budget_krw) -> Option<BuyResult> {
    Walk through asks, consuming liquidity until budget exhausted.
    If asks exhausted before budget → None (insufficient liquidity).
    avg_price = budget / total_qty
    slippage = (avg_price - best_ask) / best_ask × 100
}
```

**Sell slippage** (consuming bids):
```
fn calc_sell_avg(bids, sell_qty) -> Option<SellResult> {
    Walk through bids, consuming liquidity until quantity filled.
    If bids exhausted before filled → None (insufficient liquidity).
    avg_price = total_krw / sell_qty
    slippage = (best_bid - avg_price) / best_bid × 100
}
```

### 3. Real Gap Calculation (Implemented)

```
Buy exchange (Bithumb): 2M KRW market buy → avg_buy_price, buy_qty
Sell exchange (Upbit):  buy_qty market sell → avg_sell_price, sell_total

real_gap = (avg_sell_price - avg_buy_price) / avg_buy_price × 100

net_profit = sell_total × (1 - sell_fee) - budget × (1 + buy_fee) - transfer_fee
net_profit_pct = net_profit / budget × 100
```

### 4. UI Display (Implemented)

Slippage, spread, and real gap data are shown in the coin card UI:
- Slippage percentage for buy/sell sides
- Bid-ask spread percentage per exchange
- Real gap (slippage-adjusted) when orderbook data is available

---

## Planned: Scenario Thread Integration (Not Implemented)

### 5-A: Thin Orderbook Warning
**Trigger:**
- 2M KRW buy slippage >= 0.5%
- OR total ask volume < order amount (insufficient liquidity)

### 5-B: Real Gap Alert
**Trigger:**
- Surface premium >= threshold (default 2%)
- Real gap >= threshold (default 1.5%)
- Difference between surface and real < 1%p (opportunity is genuine)

### 5-C: Fake Gap Warning
**Trigger:**
- Surface premium >= 3% (looks attractive)
- Real gap < 1% (profit is minimal after slippage)

---

## Spread Calculation

Per-exchange bid-ask spread:
```
spread = (best_ask - best_bid) / best_bid × 100
```

| Spread | Assessment |
|--------|------------|
| < 0.1% | Healthy (major coins like BTC, XRP) |
| > 0.5% | Risky (small-cap altcoins) |
| > 1.0% | Dangerous (not recommended for trading) |

---

## Configuration

```toml
[orderbook]
enabled = true
order_size_krw = 2_000_000       # Reference order size (2M KRW)

# Scenario 5 thresholds (for future integration)
slippage_warning_pct = 0.5       # Slippage warning threshold
real_gap_min_pct = 1.5           # Minimum real gap for alert
fake_gap_surface_min = 3.0       # Fake gap: minimum surface gap
fake_gap_real_max = 1.0          # Fake gap: maximum real gap
spread_warning_pct = 0.5         # Spread warning threshold
```

---

## Data Flow (Current)

```
Orderbook WebSocket (exchanges/orderbook.rs)
  ├─ Upbit orderbook stream
  ├─ Bithumb orderbook stream
  └─ Binance orderbook stream
       │
       ▼
  background.rs::realtime_processor
  ├─ calc_buy_avg(BT asks, 2M KRW)
  ├─ calc_sell_avg(UP bids, buy_qty)
  └─ real_gap, slippage_pct, spread_pct
       │
       └─► UI: coin card displays slippage/spread/real gap
           (Scenario thread integration → future Phase 3)
```
