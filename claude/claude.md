# Kimchi Premium Arbitrage — Project Summary

## Overview

A Korean CEX (Centralized Exchange) arbitrage monitoring bot built in **Rust**. Exploits price gaps (Kimchi Premium) between Korean exchanges (Bithumb, Upbit) and global exchanges (Binance, Bybit, Bitget, OKX, Gate.io). Native macOS GUI built with egui/eframe.

---

## Key Premises

1. **Bithumb has low external withdrawal limits** (~4,500 average)
2. **Direction**: When Upbit is more expensive than Bithumb, buy on Bithumb and transfer to Upbit
3. **Withdrawal limit is small** relative to the gap, so keep running
4. **Hedging strategy**: Use quantity-based hedging on offshore exchange (Binance, Bybit, Bitget, OKX, or Gate)

---

## Strategy Analysis

### Quantity Hedging vs Amount Hedging

| Approach | Description | Risk Exposure |
|----------|-------------|---------------|
| **Quantity Hedging** | Short same number of coins offshore | Neutral to coin price movement |
| Amount Hedging | Short same KRW value offshore | Exposed due to quantity mismatch |

**Conclusion**: Quantity hedging is preferred because it neutralizes coin price volatility.

---

## Profit Simulation (3% Gap)

### Conditions
- Bithumb: 122 KRW, Upbit: 126 KRW (+3%), Offshore: 121 KRW
- Investment: 5,000,000 KRW, All fees: 0.05%

### Results

| Step | Calculation | Result |
|------|-------------|--------|
| Bithumb Buy | 5,000,000 ÷ 122 | 40,984 coins |
| Transfer (0.05% fee) | 40,984 × 0.9995 | 40,963 coins |
| Upbit Sell | 40,963 × 126 × 0.9995 | 5,158,772 KRW |
| **Domestic Profit** | | **+156,272 KRW** |

With quantity hedge: ~151,000 KRW net profit (3% return). Breakeven buffer: offshore can rise up to +3.3%.

---

## Key Insights

1. **3% gap threshold is safe** — provides enough buffer against Kimchi premium fluctuations
2. **Quantity hedging neutralizes price risk** — only exposed to Kimchi premium changes
3. **Price co-movement** — when offshore rises, domestic usually rises too, so hedge works
4. **1.64% gap is risky** — only ~1.9% buffer, too thin

---

## Technical Decisions

### WebSocket for All Exchanges

| Aspect | WebSocket | REST API |
|--------|-----------|----------|
| Latency | ~10ms | 50-200ms |
| Complexity | Higher | Lower |
| Rate Limits | Minimal | Per-minute caps |

**Decision**: All 7 exchanges use WebSocket for real-time price monitoring via the unified `TickerWsManager`. REST is only used for wallet/transfer APIs.

### Snapshot Swap Architecture

Background tasks write to shared `Arc<Mutex<...>>` data stores. A snapshot publisher (200ms interval) reads all data once, builds an immutable `Arc<AppSnapshot>`, and swaps it into `SharedSnapshot = Arc<RwLock<Arc<AppSnapshot>>>`. The UI reads a single `Arc::clone()` per frame — no mutex contention.

### parking_lot for Synchronization

All `std::sync::Mutex` replaced with `parking_lot::Mutex`/`RwLock` for 2-5x lower uncontended lock overhead (user-space spinlock, no syscall, no poisoning).

---

## Current Implementation Status

### Completed
1. **Phase 0 — Gap Monitor GUI**: Real-time price monitoring across 7 exchanges via WebSocket, Kimchi premium calculation, native macOS GUI with egui, sortable/filterable coin list, D/W status display, favorites, alerts, proxy support
2. **Phase 0.5 — Cross-Exchange Transfer**: Transfer state machine (6 steps), all 7 exchange APIs, 21-network chain mapping, separate viewport UI with progress tracking
3. **Phase 0.7 — Performance Optimization**: 8 improvements achieving 100ms refresh (virtual scrolling, snapshot swap, parking_lot, dirty flag caching, parallel reads, string formatting cache)
4. **Phase 0.8 — UI Redesign**: Custom dark theme (GitHub-dark palette), header/card/scenario panel redesign, typography hierarchy, spacing consistency
5. **Scenario Detection**: 3 active scenario types — GapThreshold (Kimchi premium crossing), DomesticGap (UP-BT gap), FutBasis (spot-futures basis)
6. **Orderbook Slippage**: WebSocket orderbook collection, slippage/spread calculation, real gap display in UI
7. **Phase 0.9 — Transfer UX Improvements**: Whitelist error parsing (Upbit/Bithumb), address preview in confirmation dialog, empty deposit address validation
8. **Phase 1.0 — Bybit/OKX Extension**: Scenario detection extended to all 3 overseas exchanges (per-exchange thread keys UP-BN/BB/OK, BT-BN/BB/OK), wallet D/W status for Bybit and OKX
9. **Phase 1.1 — D/W Status Panel + Scenario Click Fix**: Replaced chronological D/W event log with real-time blocked status view (per-coin, per-exchange, per-chain). Fixed scenario click handler to parse per-exchange keys (UP-BN, BT-BB, etc.)
10. **Phase 1.2 — UI/UX Polish**: Asset Move viewport close fix (AtomicBool), transfer exchange buttons split into 2 rows, USD prices to 6 decimals, dual BN/BB futures display, enlarged quick transfer panel (400px default), unified inline/viewport transfer form
11. **Phase 1.3 — Market Order Integration**: Pre/post-transfer auto buy/sell (Option A), expandable coin detail panel with per-exchange Buy/Sell + confirmation dialog (Option C), Cmd+W viewport close, inline transfer with job progress panel

### Next Steps (Planned)
1. Implement execution module (buy + hedge)
2. Implement closing module (sell + close hedge)
3. Additional scenario types (D/W combined signals, rapid change detection)
4. Alerting infrastructure (Telegram)

---

## Reference Documents

| Document | Description |
|----------|-------------|
| `spec.md` | System specification, module details, project structure |
| `task.md` | Phase-by-phase task tracking |
| `scenario.md` | Scenario detection specification |
| `orderbook_slippage.md` | Orderbook slippage calculation details |
| `Gap_Strategy.md` | Complete arbitrage strategy reference |
| `performance_improvements.md` | 8 performance optimizations (all complete) |
| `ui_improvements.md` | 7 UI visual improvements (all complete) |
