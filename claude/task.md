# KimchiCEX Arbitrage Bot — Task Tracking

## Phase Overview

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 0 | Gap Monitor GUI | **Completed** |
| Phase 0.5 | Cross-Exchange Transfer Module | **Completed** |
| Phase 0.7 | Performance Optimization (100ms refresh) | **Completed** |
| Phase 0.8 | UI Visual Redesign | **Completed** |
| Phase 0.9 | Transfer UX Improvements | **Completed** |
| Phase 1.0 | Bybit/OKX Extension (Scenario + D/W) | **Completed** |
| Phase 1.1 | D/W Status Panel + Scenario Click Fix | **Completed** |
| Phase 1.2 | UI/UX Polish (Transfer, Futures, Precision) | **Completed** |
| Phase 1 | Execution Module (buy + hedge) | Not Started |
| Phase 2 | Closer Module (sell + close hedge) | Not Started |
| Phase 3 | Infrastructure (alerting, monitoring) | Not Started |
| Phase 4 | Testing & Hardening | Not Started |

---

## Phase 0: Gap Monitor GUI (Completed)

### Core Features
- [x] WebSocket price monitoring for all 7 exchanges via `TickerWsManager`
- [x] Kimchi premium calculation (per-coin, relative to USDT premium baseline)
- [x] Native macOS GUI with egui/eframe
- [x] Auto-discover all common coins across exchanges
- [x] Sortable columns with None values at bottom
- [x] Binance KRW conversion (using Upbit USDT/KRW market)
- [x] USD/KRW forex rate display
- [x] Color-coded premium display
- [x] Deposit/Withdrawal status display (REST polling)
- [x] API keys via .env file (7 exchanges)
- [x] Proxy support (HTTP/SOCKS5, per-exchange)

### UI Features
- [x] Search/filter by coin name
- [x] Favorites (pin coins to top)
- [x] Alert settings (popup when premium >= threshold)
- [x] Filter options (D/W available, All exchanges only)
- [x] Settings window (threshold slider, filter toggles)
- [x] Quick filter toggles in toolbar

### Scenario Detection
- [x] `ScenarioDetector` with thread-based logging
- [x] Scenario 1: GapThreshold (Kimchi premium crossing)
- [x] Scenario 2: DomesticGap (UP-BT domestic gap)
- [x] Scenario 3: FutBasis (spot-futures basis)

### Orderbook & Slippage
- [x] Orderbook collection via WebSocket
- [x] Slippage calculation (buy avg / sell avg for given order size)
- [x] Spread calculation (bid-ask spread %)
- [x] Real gap calculation (slippage-adjusted effective gap)
- [x] UI display of slippage/spread/real gap data

### Database
- [x] SQLite for D/W event logging (`LogEntry`)
- [x] Session start/end tracking
- [x] Load historical logs on startup

### Planned (Not Started)
- [ ] Profit calculator (with fees)
- [ ] Volume display
- [ ] One-click to exchange trading page
- [ ] CSV export
- [ ] Network selection for D/W

---

## Phase 0.5: Cross-Exchange Transfer Module (Completed)

- [x] Transfer state machine (6 steps: idle → withdrawal requested → processing → confirmed → deposit processing → deposit confirmed → completed)
- [x] Transfer API client for all 7 exchanges
- [x] Cross-exchange network name mapping (21 networks across 7 exchanges)
- [x] Transfer executor (async state machine with polling)
- [x] Transfer window UI (separate egui viewport)
  - [x] Left panel: Transfer Settings + Deposit Info + Log
  - [x] Right panel: 6-step Transfer Progress
  - [x] Editable deposit address and tag/memo fields
  - [x] Confirmation dialog before execution
  - [x] TX hash display with copy button and browser link
  - [x] 30-minute timeout warning
- [x] API key support with passphrase (Bitget, OKX)
- [x] Auth patterns: HMAC-SHA256 (Binance, Bybit), JWT (Upbit), HMAC-SHA512 (Bithumb, Gate), HMAC-SHA256→base64 (Bitget, OKX)

---

## Phase 0.7: Performance Optimization (Completed)

All 8 improvements implemented. See `performance_improvements.md` for details.

| # | Improvement | Impact |
|---|------------|--------|
| 1 | Virtual scrolling | Render cost 90%↓ |
| 2 | Snapshot swap pattern | Lock time 95%↓, allocations 95%↓ |
| 3 | parking_lot transition | Lock overhead 3-5x↓ |
| 4 | Dirty flag sort caching | Sort frequency 90%↓ |
| 5 | Orderbook parallel read | Processor latency ~2x↓ |
| 6 | Repaint 100ms | 2Hz → 10Hz |
| 7 | Scenario pre-sorted snapshot | Clone+sort elimination |
| 8 | String formatting cache | ~200 alloc/frame eliminated |

---

## Phase 0.8: UI Visual Redesign (Completed)

All 6 improvements implemented. See `ui_improvements.md` for details.

| # | Improvement |
|---|------------|
| 1 | Custom dark theme (GitHub-dark palette `#0D1117`) |
| 2 | Header panel redesign (Tether badge, icon toggles) |
| 3 | Coin card redesign (pill badges, micro-tags, dot D/W) |
| 4 | Scenario panel improvement (accent bars, dimmed inactive) |
| 5 | Typography hierarchy (15/13/11px) |
| 6 | Spacing & layout consistency |

---

## Phase 0.9: Transfer UX Improvements (Completed)

Addressed user-facing issues in the cross-exchange transfer module.

### Withdrawal Whitelist Error Handling
- [x] Parse Upbit `withdraw_address_not_registered` / `withdraw_address_not_matched` errors into actionable English messages with the rejected address included
- [x] Parse Bithumb equivalent whitelist errors with the same pattern
- [x] Previously: raw JSON error body shown to user (e.g., `{"error":{"message":"등록된 출금 주소가 아닙니다.","name":"withdraw_address_not_registered"}}`)

### Transfer Confirmation Dialog Enhancement
- [x] Show destination address (monospace) in confirmation step before executing
- [x] Show tag/memo if present
- [x] Display orange warning for Korean exchange sources (Upbit, Bithumb): "requires this address to be registered in your withdrawal whitelist"

### Empty Deposit Address Validation
- [x] Detect when `fetch_deposit_address` returns an empty string (API success but no address generated)
- [x] Show error: "Deposit address not generated on [exchange]. Please generate it on the exchange website first."
- [x] Previously: empty address silently stored, Execute button disabled with no explanation

### Files Changed
- `src/exchanges/transfer_api.rs` — Upbit/Bithumb withdrawal error parsing
- `src/transfer/executor.rs` — Empty address validation after fetch
- `src/transfer/window.rs` — Address preview + whitelist warning in confirmation dialog

---

## Phase 1.0: Bybit/OKX Extension (Completed)

Extended scenario detection and wallet D/W status to cover all three overseas exchanges (Binance, Bybit, OKX).

### Task A: Scenario Detection — Per-Exchange Keys
- [x] `ScenarioDetector.on_price_update` accepts `kimp_by_exchange: &[(&str, f64, f64)]` instead of scalar Binance-only values
- [x] `check_gap_threshold` refactored to per-key tracking (`HashMap<String, bool>` keyed by `"SYMBOL:KEY"`)
- [x] Thread keys: `UP-BN`, `UP-BB`, `UP-OK`, `BT-BN`, `BT-BB`, `BT-OK`
- [x] `background.rs` iterates all 3 overseas exchanges with depth-based kimp fallback to mid-price kimp
- [x] `app.rs` wallet filtering parses per-exchange keys (domestic deposit + overseas withdraw)

### Task B: Wallet D/W Status — Bybit & OKX
- [x] `CoinWalletStatus` extended with `bybit` and `okx` fields
- [x] `WalletStatusFetcher` extended with `bybit_client`/`okx_client` + fetch methods
- [x] Bybit: `GET /v5/asset/coin/query-info` (HMAC-SHA256, multi-chain per coin)
- [x] OKX: `GET /api/v5/asset/currencies` (HMAC-SHA256→base64, passphrase, multi-chain)
- [x] Background polling loop extended with change detection (`BB`/`OK` codes)
- [x] `get_exchange_wallet_status` extended with `Bybit`/`Okx` arms

### Files Changed
- `src/scenario.rs` — per-exchange key tracking, refactored `check_gap_threshold`
- `src/background.rs` — multi-exchange kimp feed, Bybit/OKX wallet polling
- `src/app.rs` — wallet filtering for per-exchange keys, `get_exchange_wallet_status`
- `src/models/price.rs` — `CoinWalletStatus` fields
- `src/exchanges/wallet_status.rs` — Bybit/OKX fetch methods

---

## Phase 1.1: D/W Status Panel + Scenario Click Fix (Completed)

Replaced the chronological D/W event log with a real-time blocked status view, and fixed the scenario click handler for per-exchange keys.

### Scenario Click Handler Fix
- [x] Parse per-exchange GapThreshold keys (`UP-BN`, `BT-BB`, `BT-OK`, etc.) into correct `exchange_a`/`exchange_b` pair
- [x] Previously hardcoded `Exchange::Binance` for all GapThreshold clicks

### ExchangeWalletStatus — Per-Chain Blocked Info
- [x] Added `deposit_blocked_chains: Vec<String>` and `withdraw_blocked_chains: Vec<String>` fields
- [x] Multi-chain exchanges (Binance, Bybit, OKX): collect specifically blocked chain names from `Vec<WalletStatus>` responses
- [x] Single-chain exchanges (Upbit): use `net_type` as chain name when blocked
- [x] Bithumb (no chain info): use `"default"` as chain name when blocked

### D/W Status Panel Redesign
- [x] Replaced chronological `LogEntry` event log with current blocked status view
- [x] Groups blocked coins alphabetically, shows per-exchange deposit/withdraw status
- [x] Displays blocked chain names (e.g., `D:blocked [ETH, BSC]`) for multi-chain exchanges
- [x] Clickable coin symbols (filters main table like before)
- [x] Header shows blocked coin count with red/green color indicator

### Files Changed
- `src/models/price.rs` — `ExchangeWalletStatus` extended with per-chain fields
- `src/background.rs` — All 5 exchange wallet polling blocks updated to populate chain info
- `src/app.rs` — D/W panel replaced (lines 989-1077), scenario click handler fixed (lines 1088-1113)

---

## Phase 1.2: UI/UX Polish — Transfer, Futures, Precision (Completed)

Six improvements across transfer module, futures display, and price formatting.

### Asset Move Viewport Close Fix
- [x] `TransferWindow.is_open` changed from `bool` to `Arc<AtomicBool>`
- [x] Viewport callback detects `close_requested()` and sets `is_open = false`
- [x] `app.rs` reads/writes via `Ordering::Relaxed` atomic ops

### Transfer Module — Exchange Buttons 2 Lines
- [x] `exchange_button_group` (ui_helpers.rs) split into two horizontal rows: domestic (line 1) + overseas (line 2)

### USD Price 6 Decimal Places
- [x] All USD price format strings changed from `{:.4}` to `{:.6}` (4 occurrences in app.rs)
- [x] KRW formatting (`format_price`) unchanged

### Dual Futures Exchange Display (BN + BB)
- [x] `CoinFormatted` extended with `fut_label_bn`, `fut_label_bb`, `fut_price_label_bn`, `fut_price_label_bb`
- [x] Per-exchange basis calculation: `(Binance_Spot - Futures) / Futures * 100` for both BN and BB
- [x] UI renders both BN and BB futures pills stacked when both available

### Quick Transfer Panel Enlarged
- [x] Bottom panel default height: 140 → 400, max height: 200 → 800

### Unified Transfer Form (Inline = Viewport)
- [x] Extracted `render_transfer_form()` from `render_panels`'s CentralPanel body
- [x] `render_inline_transfer` now calls `render_transfer_form("inline", ...)` — identical form layout
- [x] `render_panels` calls `render_transfer_form("viewport", ...)`
- [x] ID salt prevents egui ID conflicts when both panels are open simultaneously

### Files Changed
- `src/app.rs` — USD precision, futures dual display, panel height, AtomicBool read/write
- `src/transfer/window.rs` — `is_open` to Arc<AtomicBool>, close detection, form extraction
- `src/transfer/ui_helpers.rs` — 2-row exchange button layout

---

## Phase 1.3: Market Order Integration (Completed)

Added market buy/sell functionality via two complementary approaches.

### Option A: Pre/Post-Transfer Market Orders
- [x] `auto_buy_before_transfer` field added to `TransferState`, `TransferJob`, `FormCommandSnapshot`
- [x] "Auto-buy before transfer" checkbox in transfer form (next to existing "Auto-sell on arrival")
- [x] Executor performs market buy on `from_exchange` before initiating withdrawal (with 1s delay for settlement)
- [x] On pre-buy failure: job fails immediately (no withdrawal attempted)
- [x] Existing `auto_sell_on_arrival` post-transfer sell unchanged

### Option C: Coin Detail Panel with Per-Exchange Buy/Sell
- [x] Click any coin row in gap monitor to expand detail panel
- [x] Detail panel shows all 5 exchanges with current prices (KRW + USD where available)
- [x] Per-exchange Buy/Sell buttons with amount input field
- [x] Confirmation dialog with side/amount/exchange display before execution
- [x] Orders routed through existing `TransferCommand::SubmitMarketOrder` pipeline
- [x] Expanded row highlighted with subtle color shift

### Cmd+W Viewport Close
- [x] Asset Move viewport responds to `Cmd+W` keyboard shortcut (sends `ViewportCommand::Close`)

### Quick Transfer — Inline Job Progress
- [x] `render_transfer_form` gains `show_log: bool` parameter; viewport=true, inline=false
- [x] `render_inline_transfer` rewritten: left side = form (no log), right side = 280px Transfer Progress panel
- [x] Job snapshot + dismiss/cancel pattern reused from viewport `render_panels`

### Files Changed
- `src/transfer/state.rs` — `auto_buy_before_transfer`, `market_buy_result` fields
- `src/transfer/executor.rs` — Pre-buy logic before withdrawal
- `src/transfer/window.rs` — `show_log` param, Cmd+W close, inline layout rewrite
- `src/app.rs` — `PendingMarketOrder`, `render_coin_detail_inline`, expanded coin state

---

## Phase 1: Execution Module (Not Started)

- [ ] Pre-trade validation checks (balance, D/W status, network, slippage)
- [ ] Async parallel order execution (domestic buy + offshore short)
- [ ] Quantity matching validation
- [ ] Rollback logic for partial fills
- [ ] Position state management
- [ ] Integration with scenario detection (auto-trigger from high-confidence signals)

---

## Phase 2: Closer Module (Not Started)

- [ ] Deposit confirmation listener
- [ ] Parallel close execution (domestic sell + offshore close short)
- [ ] P&L calculation
- [ ] Position finalization and logging

---

## Phase 3: Infrastructure (Not Started)

### Alerting
- [ ] Telegram bot integration
- [ ] Alert types (trade executed, error, timeout, high-premium signal)
- [ ] Daily summary reports

### Additional Scenarios
- [ ] Scenario 4: D/W + Gap Combined Signal (D/W open + high premium)
- [ ] Scenario 5: Rapid Change Detection (rolling window rate-of-change)
- [ ] Scenario 6: Orderbook Slippage as scenario threads (thin/real/fake gap detection)

---

## Phase 4: Testing & Hardening (Not Started)

- [ ] Paper trading mode
- [ ] Stress testing (connection drops, API errors)
- [ ] Circuit breaker implementation (daily trade limits, loss limits)
- [ ] Position timeout handling

---

## Notes

- The project started with a Python/Redis/PostgreSQL plan but was fully implemented in Rust with egui and SQLite
- All price data uses WebSocket (no REST polling for tickers)
- Start with small amounts for live testing
- Always run with position limits during initial deployment
