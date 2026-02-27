# Kimchi Premium Monitor UI — AI-Native Specification

## 0. What This Document Is
- This is an implementation contract, not a narrative guide.
- It is optimized for both humans and AI coding agents.
- Every section answers: what exists, what changes are allowed, and how to verify correctness.
- **Runtime**: Python Textual TUI (migrated from Rust egui). Canonical source is `tui_textual/src/kimchi_tui/`.

## 1. Product Intent Contract
- Primary mission: detect and act on cross-exchange arbitrage signals with minimal latency.
- Design principle: signal clarity over decoration.
- Operational mode: always-on monitoring with high information density.
- Core user actions:
  - scan premium/futures signals
  - inspect one coin deeply
  - execute transfer or market actions quickly

## 2. UI Topology Contract

| Region | Textual Widget ID | Responsibility |
|---|---|---|
| Header | `#kimchi-header` (`KimchiHeader`) | Tether premium, rates, coin count, freshness, alerts, theme, clock |
| Tab bar | `#main-tabs` (`TabbedContent`) | 5 tabs: Monitor, Transfer, Scenarios, D/W Status, Logs |
| Monitor (tab) | `#monitor-screen` | Coin table/cards, detail panel, scenario panel, D/W summary, recent logs |
| Transfer (tab) | `#transfer-screen` | Quick transfer form, progress panel, transfer history |
| Scenarios (tab) | `#scenarios-screen` | Threshold config, filterable scenario thread list |
| D/W Status (tab) | `DWStatusScreen` | Full D/W matrix (DataTable) per coin × exchange |
| Logs (tab) | `LogsScreen` | Full system log viewer (RichLog) |
| Status bar | `#kimchi-statusbar` (`KimchiStatusBar`) | Context-sensitive keybinding help per active tab |

### 2.1 App Layout Hierarchy

```
KimchiApp
├── KimchiHeader
├── TabbedContent (#main-tabs)
│   ├── TabPane "Monitor" (#monitor-tab)
│   │   └── MonitorScreen
│   │       ├── Static (#rate-bar)
│   │       ├── FilterBar (#filter-bar)
│   │       └── Horizontal (#monitor-main)
│   │           ├── Vertical (#monitor-left)
│   │           │   ├── CoinTable (#coin-table)
│   │           │   ├── CardView (#card-view)           [toggle: v]
│   │           │   └── CoinDetailPanel (#coin-detail-panel) [toggle: e]
│   │           └── Vertical (#monitor-right)
│   │               ├── ScenarioPanel (#monitor-scenario)
│   │               ├── Vertical (#monitor-dw)
│   │               └── Vertical (#monitor-logs)
│   ├── TabPane "Transfer" (#transfer-tab)
│   │   └── TransferScreen
│   ├── TabPane "Scenarios" (#scenarios-tab)
│   │   └── ScenariosScreen
│   ├── TabPane "D/W Status" (#dw-tab)
│   │   └── DWStatusScreen
│   └── TabPane "Logs" (#logs-tab)
│       └── LogsScreen
└── KimchiStatusBar
```

## 3. Component Contracts

### 3.1 Header Contract (`widgets/header_bar.py`)
- Inputs: `snapshot.usdt_krw`, `snapshot.usd_krw_forex`, `snapshot.coin_states` count, `snapshot.last_ticker_age_ms`, `snapshot.scenario_threads`, `dark_mode`
- Outputs: Single-line Rich markup with: Tether premium%, USDT rate, USD rate, coin count, freshness indicator, alert indicator, theme indicator, clock
- Display: `freshness_indicator()` (Live/Stale/Connecting), `alert_indicator()` (active scenario count)
- Hard rules:
  - Tether premium = `(usdt_krw / usd_krw_forex - 1) × 100`, colored SUCCESS (>=0) or ERROR (<0)
  - Theme toggle must apply full visual token switch via `set_theme(dark=bool)`, not partial overrides

### 3.2 Filter Bar Contract (`widgets/filter_bar.py`)
- State: `query` (str), `dw_only` (bool), `exchange_filter` (ALL/UP/BT/BN/BB/OK), `sort_column` (Symbol/Upbit/Bithumb/Binance/Bybit/Kimp), `sort_desc` (bool)
- Outputs: `FilterBar.Changed` message on any state mutation
- Controls: Search input, D/W toggle button, Exchange rotation button, Sort column button, Sort direction button
- Hard rules:
  - Search matches against `symbol.upper()` and `korean_name.upper()`
  - Exchange rotation: ALL → UP → BT → BN → BB → OK → ALL
  - Sort column rotation: Symbol → Upbit → Bithumb → Binance → Bybit → Kimp → Symbol

### 3.3 Coin Table Contract (`widgets/coin_table.py`)

**Columns** (10 total):

| # | Column | Data Source | Formatter | Sortable |
|---|--------|-------------|-----------|----------|
| 1 | `#` | Enumerated index | `str(idx)` | No |
| 2 | `Symbol` | `coin.symbol` | `_fmt_symbol()` — bold accent if in favorites | Yes |
| 3 | `Korean` | `snapshot.korean_names[symbol]` | Direct string | No |
| 4 | `Upbit` | `coin.upbit_price` | `fmt_krw_price()` | Yes |
| 5 | `Bithumb` | `coin.bithumb_price` | `fmt_krw_price()` | Yes |
| 6 | `Binance` | `coin.binance_krw` | `fmt_krw_price()` | Yes |
| 7 | `Bybit` | `coin.bybit_krw` | `fmt_krw_price()` | Yes |
| 8 | `Okx` | `coin.okx_krw` | `fmt_krw_price()` | No |
| 9 | `Kimp%` | `_primary_kimchi(coin)` — upbit_kimchi fallback bithumb_kimchi | `fmt_pct_rich()` | Yes (default) |
| 10 | `DomGap%` | `coin.domestic_gap` | `fmt_domgap_rich()` | No |

- Default sort: `Kimp` descending
- Filtering: query (symbol/name), dw_only (blocked coins only), exchange_filter
- Row cursor type, zebra stripes enabled
- Cache: `_row_cache` dict for cell-level diff updates (only changed cells re-rendered)

### 3.4 Card View Contract (`widgets/card_view.py`)
- Alternative display mode to CoinTable (toggled by `v` key)
- Uses same filter/sort pipeline as CoinTable via `CoinTable.select_coins()`
- Per-coin card layout (5 lines):
  1. `★ SYMBOL  Korean_Name  +X.XX%` (star if favorite, kimp via `fmt_kimp_markup()`)
  2. `BT ₩price (USD)  D/W  UP ₩price (USD)  D/W` (via `fmt_krw_price()`, `fmt_usd_equiv()`, `dw_dots()`)
  3. `slip BT:X% UP:X%  |  real_gap:X%` (via `fmt_signed_pct()`)
  4. Exchange badges: `[BNF ±X%] [BB ±X%] [OK ±X%]` (futures_basis, bybit/okx kimchi)
  5. Binance USD price: `$X.XXXX` (via `fmt_usd_price()`)
- Cards separated by `─` divider line

### 3.5 Expanded Coin Detail Contract (`widgets/detail_panel.py`)
- Inputs: selected coin symbol, snapshot
- Outputs: Multi-line Rich markup with prices, kimchi premiums (6 pairs), orderbook slippage, D/W status
- Display sections:
  - Prices: Upbit, Bithumb, Binance, Bybit, OKX (via `fmt_krw_price()`)
  - Kimp: UP-BN, UP-BB, UP-OK, BT-BN, BT-BB, BT-OK (via `fmt_kimp_markup()`)
  - Orderbook: BT buy slippage, UP sell slippage, real gap (via `fmt_signed_pct()`)
  - D/W: per-exchange checkmark indicators (via `dw_checkmarks()`)
- Hard rules:
  - Only shown when `expanded_coin` is set and card_view is off
  - Toggled by `e` key (deliberate action, not row click)

### 3.6 Scenario Panel Contract (`screens/scenarios.py`)
- Inputs: `snapshot.scenario_threads`, `snapshot.scenario_config`
- Outputs: filtered/expanded scenario rendering, threshold updates via IPC
- Thread display fields: dot (●/○), scenario badge, symbol, key, message, timestamp, state
- Sub-entries: last 6 entries shown when expanded (reversed chronological)
- Threshold editing: 3 fields (K=KIMP, D=DomGap, F=FutBasis) with Tab/→ cycling, Enter to apply, Esc to cancel
- Filters: All → GapThreshold → DomesticGap → FutBasis → All
- Hard rules:
  - Threshold edits send `SetScenarioThreshold` IPC command
  - Threads sorted by `main_timestamp` descending (newest first)

### 3.7 D/W Status Contract (`screens/dw_status.py`)

**Columns** (6 total):

| Column | Data Source | Formatter |
|--------|-------------|-----------|
| `Symbol` | sorted wallet_status keys | Direct string |
| `UP` | `status.upbit` | `dw_cell_with_chains()` |
| `BT` | `status.bithumb` | `dw_cell_with_chains()` |
| `BN` | `status.binance` | `dw_cell_with_chains()` |
| `BB` | `status.bybit` | `dw_cell_with_chains()` |
| `OK` | `status.okx` | `dw_cell_with_chains()` |

- Shows ALL coins (not just blocked), with D/W indicators per exchange
- Title dynamically shows blocked count: "D/W Blocked Status (Blocked: X / Y)"
- D/W indicator format: `D/W` (both OK), `d/W` (deposit blocked), `D/w` (withdraw blocked), `d/w` (both blocked)
- Blocked chains shown as badges (max 2 per exchange)

### 3.8 Transfer Contract (`screens/transfer.py`)
- Layout: Left panel (form + log), Right panel (progress + history)
- Exchanges: 7 total — Upbit, Bithumb (Domestic) | Binance, Bybit, Bitget, Okx, Gate (Global)
- Form fields: Coin, From exchange, To exchange, Network (dynamic buttons), Address, Tag/Memo, Amount (with 25/50/75/100% buttons)
- Toggles: Auto-buy before transfer, Auto-sell on arrival, Wallet mode (personal wallet vs exchange)
- Actions: Market Buy, Market Sell, Execute Transfer
- IPC commands: SetTransferFrom/To, FetchNetworks, FetchBalance, FetchDepositAddress, SelectNetwork, SetAmountRatio, ExecuteTransfer, SetAutoBuy/Sell, SubmitMarketOrder

**Transfer History Table** (text-rendered, 6 columns):

| Column | Width | Data Source | Formatter |
|--------|-------|-------------|-----------|
| `ID` | 4 | `job.id` | Left-aligned int |
| `Coin` | 8 | `job.coin` | Left-aligned string |
| `Route` | 24 | `from->to` | Truncated with "..." |
| `Amount` | 14 | `job.amount` | `fmt_amount()` |
| `Status` | 10 | executing/error state | `job_status_badge()` |
| `Time` | 10 | `job.started_at_secs` | `fmt_elapsed()` |

- Scrollable via j/k with offset pagination, shows "Rows X-Y / Total"
- Active transfer progress: step chain with `step_marker()` indicators

### 3.9 Logs Contract (`screens/logs.py`)
- Display: Last 500 log entries via RichLog widget
- Fields per line: timestamp (secondary color), `log_level_badge()`, symbol (bold, 8-char), message
- Incremental append: only new lines written if prefix matches cached lines
- Scrollable via j/k, jump to bottom via G

## 4. Design Token Contract

### 4.1 Semantic Token Source
- Canonical source: `colors.py` — `Dark` / `Light` classes accessed via `_ThemeProxy` singleton `C`
- Theme switching: `set_theme(dark=bool)` swaps `C._theme` between `Dark` and `Light`
- All widget code reads `C.<TOKEN>` — never raw hex colors

**Complete Token Inventory:**

| Category | Token | Dark Value | Light Value | Usage |
|----------|-------|------------|-------------|-------|
| **Surface** | `PANEL_BG` | `#0d1117` | `#ffffff` | Panel backgrounds |
| | `BG_DEEP` | `#161b22` | `#f6f8fa` | Deep background |
| | `CARD_BG` | `#1c2128` | `#ffffff` | Card background |
| | `CARD_BG_ALT` | `#161b22` | `#f6f8fa` | Alternate card background |
| | `WIDGET_BG` | `#1c2128` | `#eaeef2` | Widget background |
| | `BORDER` | `#30363d` | `#d0d7de` | Border color |
| | `STATUSBAR_BG` | `#0f3460` | `#0969da` | Status bar background |
| | `HOVER_BG` | `#30363d` | `#d8dee4` | Hover state background |
| **Text** | `TEXT_PRIMARY` | `#e6edf3` | `#24292f` | Primary text |
| | `TEXT_SECONDARY` | `#8b949e` | `#57606a` | Secondary text |
| | `TEXT_MUTED` | `dim` | `dim` | Muted/disabled text |
| **Action** | `ACCENT` | `#58a6ff` | `#0969da` | Accent color (links, highlights) |
| | `ACCENT_ACTIVE` | `#58a6ff` | `#0969da` | Active accent |
| | `SELECTION_BG` | `#30363d` | `#eaeef2` | Selection background |
| **Status** | `SUCCESS` | `#3fb950` | `#1a7f37` | Success states |
| | `ERROR` | `#f85149` | `#cf222e` | Error states |
| | `WARNING` | `#ff8c00` | `#bc4c00` | Warning states |
| | `INFO` | `#58a6ff` | `#0969da` | Informational states |
| | `PROGRESS` | `#58a6ff` | `#0969da` | Progress indicator |
| **Kimchi Premium** | `KIMP_5PLUS` | `#f85149` | `#cf222e` | >=5% (bold red) |
| | `KIMP_3TO5` | `#ff8c00` | `#bc4c00` | >=3% (yellow/orange) |
| | `KIMP_1TO3` | `#3fb950` | `#1a7f37` | >=1% (green) |
| | `KIMP_NEUTRAL` | `#8b949e` | `#57606a` | -1% ~ +1% (gray) |
| | `KIMP_NEG1TO3` | `#58a6ff` | `#0969da` | -1% ~ -3% (blue) |
| | `KIMP_NEG3TO5` | `#388bfd` | `#0550ae` | -3% ~ -5% (deep blue) |
| | `KIMP_NEG5PLUS` | `#1f6feb` | `#033d8b` | <=-5% (bold blue) |
| **Domestic Gap** | `DOMGAP_HIGH` | `#ff8c00` | `#bc4c00` | abs >=3% |
| | `DOMGAP_MID` | `#d2991a` | `#9a6700` | abs >=1% |
| | `DOMGAP_LOW` | `#8b949e` | `#57606a` | abs <1% |
| **Futures Basis** | `FUTBASIS_HIGH` | `#f85149` | `#cf222e` | abs >=1% |
| | `FUTBASIS_MID` | `#ff8c00` | `#bc4c00` | abs >=0.3% |
| | `FUTBASIS_LOW` | `#3fb950` | `#1a7f37` | else |
| **Badges** | `BADGE_POSITIVE_BG` | `#238636` | `#1a7f37` | Positive value badge bg |
| | `BADGE_NEGATIVE_BG` | `#da3633` | `#cf222e` | Negative value badge bg |
| | `BADGE_BT` | `#30363d` | `#d0d7de` | Bithumb badge bg |
| | `BADGE_UP` | `#58a6ff` | `#0969da` | Upbit badge bg |
| **Scenario Badges** | `BADGE_KIMP` | `#58a6ff` | `#0969da` | KIMP scenario badge |
| | `BADGE_DOMGAP` | `#ff8c00` | `#bc4c00` | DOM-GAP scenario badge |
| | `BADGE_FUTBASIS` | `#f85149` | `#cf222e` | FUT% scenario badge |
| **D/W Status** | `DW_OK` | `green` | `#1a7f37` | Deposit/withdraw enabled |
| | `DW_BLOCKED` | `red` | `#cf222e` | Deposit/withdraw blocked |
| | `DW_CHAIN_BADGE_BG` | `#da3633` | `#cf222e` | Blocked chain badge bg |

### 4.2 Signal Color Rules

```text
kimchi premium (heat-map convention: red = hot/expensive):
  >= 5.0   -> bold KIMP_5PLUS   (red)
  >= 3.0   -> bold KIMP_3TO5    (yellow/orange)
  >= 1.0   -> KIMP_1TO3         (green)
  > -1.0   -> KIMP_NEUTRAL      (gray)
  > -3.0   -> KIMP_NEG1TO3      (blue)
  > -5.0   -> bold KIMP_NEG3TO5 (deep blue)
  <= -5.0  -> bold KIMP_NEG5PLUS(bold blue)

domestic gap (absolute value):
  abs >= 3.0  -> bold DOMGAP_HIGH  (orange)
  abs >= 1.0  -> DOMGAP_MID       (amber)
  abs <  1.0  -> DOMGAP_LOW       (gray)

futures basis (absolute value):
  abs >= 1.0  -> FUTBASIS_HIGH    (red)
  abs >= 0.3  -> FUTBASIS_MID     (orange)
  else        -> FUTBASIS_LOW     (green)
```

### 4.3 Token Compliance Rule
- New UI code must not introduce raw one-off colors when a semantic token exists.
- All color references must go through `C.<TOKEN>` from `colors.py`.
- Badge text uses `bold white on {C.<TOKEN>}` Rich markup pattern.

## 5. State Schema Contract

### 5.1 Snapshot State (Python dataclasses in `models.py`)

```python
@dataclass
class Snapshot:
    coin_states: dict[str, CoinState]          # symbol -> coin data
    wallet_status: dict[str, CoinWalletStatus]  # symbol -> per-exchange D/W
    korean_names: dict[str, str]                # symbol -> Korean name
    logs: list[LogEntry]                        # system log entries
    orderbook_info: dict[str, OrderbookInfo]    # symbol -> orderbook data
    scenario_threads: list[LogThread]           # scenario thread list
    usdt_krw: Optional[float]                   # USDT/KRW rate
    usd_krw_forex: Optional[float]              # USD/KRW forex rate
    last_ticker_age_ms: Optional[int]           # ticker freshness (ms)
    transfer: TransferState                     # transfer form state
    scenario_config: ScenarioConfig             # threshold config
    transfer_jobs: list[TransferJob]            # transfer job history
```

### 5.2 CoinState Fields (19 fields)

```python
@dataclass
class CoinState:
    symbol: str
    upbit_price, bithumb_price: Optional[float]          # KRW prices
    binance_price, bybit_price, okx_price: Optional[float]  # USD prices
    binance_krw, bybit_krw, okx_krw: Optional[float]     # USD→KRW converted
    upbit_kimchi, bithumb_kimchi: Optional[float]          # UP/BT vs BN kimchi%
    bybit_kimchi_up, bybit_kimchi_bt: Optional[float]      # BB vs UP/BT kimchi%
    okx_kimchi_up, okx_kimchi_bt: Optional[float]          # OK vs UP/BT kimchi%
    domestic_gap: Optional[float]                           # UP-BT gap%
    binance_futures_price, bybit_futures_price: Optional[float]
    futures_basis: Optional[float]                          # spot-futures basis%
    timestamp: Optional[str]
```

### 5.3 UI Local State (per-screen)

**MonitorScreen:**
- `filter_state` (FilterBar.State): query, dw_only, exchange_filter, sort_column, sort_desc
- `favorites: set[str]` — starred coin symbols
- `expanded_coin: str | None` — currently expanded coin
- `card_view: bool` — table vs card display mode
- `exchange_a, exchange_b: str` — comparison pair (default BT, UP)

**TransferScreen:**
- `_last_snapshot_transfer: TransferState` — cached form state
- `_history_scroll_offset: int` — pagination offset

**ScenariosScreen:**
- `_current_filter: str` — active filter (All/GapThreshold/DomesticGap/FutBasis)
- `_selected_index: int` — cursor position
- `_expanded_ids: set[str]` — expanded thread UIDs
- `_editing_thresholds: bool` — threshold edit mode active
- `_fields: list[str]` — 3 threshold input fields
- `_active_field: int` — which field has focus (0=K, 1=D, 2=F)

### 5.4 Transfer State

```python
@dataclass
class TransferState:
    selected_coin: str
    from_exchange, to_exchange: str
    available_networks: list[NetworkInfo]
    selected_network_idx: Optional[int]
    amount: str
    balance: Optional[float]
    deposit_address, deposit_tag: str
    to_is_personal_wallet: bool
    auto_buy_before_transfer: bool
    auto_sell_on_arrival: bool
    market_order_pending: bool
    market_order_result: Optional[str]
    logs: list[TransferLogEntry]
```

## 6. Keybinding Contract

### 6.1 Global Keybindings (all tabs)

| Key | Action | Source |
|-----|--------|--------|
| `l` | Toggle dark/light theme | `KimchiApp.BINDINGS` |
| `q` | Quit application | `KimchiApp.BINDINGS` |

### 6.2 Monitor Tab Keybindings

| Key | Action | Condition |
|-----|--------|-----------|
| `j` | Cursor down in coin table | Not in card view |
| `k` | Cursor up in coin table | Not in card view |
| `g` | Jump to top of table | Not in card view |
| `G` | Jump to bottom of table | Not in card view |
| `o` | Cycle sort column | — |
| `O` | Toggle sort direction (ASC/DESC) | — |
| `e` | Expand/collapse coin detail panel | Not in card view, coin selected |
| `x` | Swap exchange A ↔ B | — |
| `d` | Toggle D/W only filter | — |
| `f` | Toggle favorite on selected coin | Coin selected |
| `v` | Toggle table/card view | — |
| `/` | Focus search input | — |
| `Escape` | Clear search, close expanded, reset | — |
| `Enter` | No-op (prevents accidental expand) | — |

### 6.3 Transfer Tab Keybindings

| Key | Action |
|-----|--------|
| `s` | Cycle source (From) exchange |
| `d` | Cycle destination (To) exchange |
| `c` | Focus coin input |
| `a` | Focus amount input |
| `p` | Focus address input |
| `m` | Focus tag/memo input |
| `n` | Cycle network selection |
| `r` | Refresh balance and networks |
| `b` | Toggle auto-buy before transfer |
| `v` | Toggle auto-sell on arrival |
| `w` | Toggle wallet mode (exchange / personal) |
| `1` | Set amount to 25% of balance |
| `2` | Set amount to 50% of balance |
| `3` | Set amount to 75% of balance |
| `4` | Set amount to 100% of balance |
| `j` | Scroll transfer history down |
| `k` | Scroll transfer history up |
| `Enter` | Execute transfer |

### 6.4 Scenarios Tab Keybindings

| Key | Action | Mode |
|-----|--------|------|
| `j` | Move selection down | Normal |
| `k` | Move selection up | Normal |
| `Enter` | Expand/collapse selected thread | Normal |
| `f` | Cycle filter (All→Gap→Dom→Fut→All) | Normal |
| `t` | Enter threshold editing mode | Normal |
| `Tab` / `→` | Next threshold field (K→D→F) | Editing |
| `0-9` / `.` | Input threshold value | Editing |
| `Backspace` | Delete last character | Editing |
| `Enter` | Apply thresholds (sends IPC) | Editing |
| `Escape` | Cancel editing, revert values | Editing |

### 6.5 D/W Status Tab Keybindings

| Key | Action |
|-----|--------|
| `j` | Cursor down in D/W table |
| `k` | Cursor up in D/W table |

### 6.6 Logs Tab Keybindings

| Key | Action |
|-----|--------|
| `j` | Scroll down |
| `k` | Scroll up |
| `G` | Jump to bottom |

### 6.7 Status Bar Help Text

The status bar dynamically shows available keybindings for the active tab. Defined in `widgets/status_bar.py` `TAB_HELP` dictionary.

## 7. Numeric Formatting Contract (`formatters.py`)

| Formatter | Input | Output | Rules |
|-----------|-------|--------|-------|
| `fmt_krw_price()` | Optional[float] | str | >=1000 → 0dp, >=1 → 2dp, <1 → 4dp, None → "-" |
| `fmt_usd_price()` | Optional[float] | str | >=1 → 4dp, <1 → 6dp, None → "-" |
| `fmt_usd_equiv()` | krw, usdt_krw | str | `krw/usdt_krw`, >=1 → 2dp, <1 → 6dp, prefixed `$` |
| `fmt_amount()` | float | str | 8dp then strip trailing zeros |
| `fmt_rate()` | Optional[float] | str | 0dp with commas, None → "--" |
| `fmt_rate_detailed()` | Optional[float] | str | 1dp with commas, None → "-" |
| `fmt_signed_pct()` | Optional[float] | str | 2dp with sign prefix, None → "-" |
| `fmt_pct_rich()` | Optional[float] | Text | Kimchi% with KIMP_* color tiers (see 4.2) |
| `fmt_domgap_rich()` | Optional[float] | Text | DomGap% with DOMGAP_* color tiers (see 4.2) |
| `fmt_kimp_markup()` | Optional[float] | str | Rich markup string version of `fmt_pct_rich()` |
| `fmt_domgap_markup()` | Optional[float] | str | Rich markup string version of `fmt_domgap_rich()` |
| `fmt_elapsed()` | int (secs) | str | >=3600 → `H:MM:SS`, else → `M:SS` |

## 8. Indicator Contract (`indicators.py`)

| Indicator | Usage | Symbols |
|-----------|-------|---------|
| `step_marker(status)` | Transfer step progress | ✓ (Completed), ⟳ (InProgress), ✗ (Failed), [ ] (Pending) |
| `job_status_badge(executing, error)` | Transfer job status | ⟳ RUNNING, ✗ FAILED, ✓ DONE |
| `dw_dots(deposit, withdraw)` | Compact D/W indicator | D●W● (OK), d●w● (blocked), D- W- (unknown) |
| `dw_checkmarks(deposit, withdraw)` | Checkmark D/W indicator | D✓W✓ (OK), D✗W✗ (blocked) |
| `dw_cell(deposit, withdraw)` | Table cell D/W | D/W, d/W, D/w, d/w, - |
| `dw_cell_with_chains(...)` | Table cell + blocked chains | D/W + chain badges (max 2) |
| `log_level_badge(level)` | Log entry badge | ERR (red), WRN (orange), INF (blue), DBG (gray) |
| `transfer_log_badge(is_error)` | Transfer log badge | ERR or INF |
| `scenario_badge(scenario)` | Scenario type badge | [KIMP], [DOM-GAP], [FUT%] (color-coded) |
| `scenario_state_text(is_active)` | Scenario state | ACTIVE (green), closed (muted) |
| `freshness_indicator(age_ms)` | Ticker freshness | ● Live X.Xs, ● Stale X.Xs, ● Connecting... |
| `alert_indicator(count)` | Active alert count | ⚠ N alerts, ✓ No alerts |

## 9. IPC Protocol Contract

- WebSocket endpoint: `ws://127.0.0.1:9876/ws` (Go backend → Python TUI)
- Snapshot broadcast: JSON, 200ms interval
- JSON format: snake_case fields matching Python `@dataclass` definitions in `models.py`
- Parse function: `parse_snapshot(data: dict) -> Snapshot`
- Commands (TUI → backend): JSON via `ipc_client.send_command(command, **kwargs)`

## 10. Performance and Reliability Contract
- IPC snapshot rate: 200ms (Go backend broadcast)
- Cell-level diff: CoinTable and DWStatusScreen only update changed cells via `_row_cache`
- Log append optimization: only write new lines when prefix matches cached lines
- Theme switch: instant — `set_theme()` swaps proxy backing class, all `C.<TOKEN>` reads resolve to new theme

## 11. AI Execution Protocol

### 11.1 Change Boundaries
- Preserve widget IDs (`#coin-table`, `#kimchi-header`, `#transfer-screen`, etc.) unless migration is explicit.
- Do not break scenario threshold editing workflow.
- Keep transfer step semantics and status progression stable.
- Do not break the tab-based navigation structure.

### 11.2 Definition of Done for UI Changes
- Behavior correctness:
  - scenario filtering, search, sorting, transfer execution all still functional
  - keybindings in all 5 tabs work correctly
- Visual correctness:
  - dark and light themes both render correctly
  - no raw hex colors — all via `C.<TOKEN>`
  - no token regression in major components
- Performance sanity:
  - no obvious render stutter with typical symbol counts
- Static verification:
  - `python3 -m compileall tui_textual/src/kimchi_tui` clean
  - LSP diagnostics (error level) clean on changed files

### 11.3 Regression Checklist
- [ ] Search still matches symbol and Korean name
- [ ] A/B exchange swap updates premium column correctly
- [ ] Scenario threshold edits refresh detector and UI thread list
- [ ] Coin expand (e key) shows detail panel; Enter does NOT expand
- [ ] Transfer command chain still works: network → balance → address → execute
- [ ] Transfer cancel/dismiss interactions still work
- [ ] Copy address/tag buttons still work
- [ ] Theme toggle does not leave mixed hardcoded colors
- [ ] Card view toggle (v) shows cards with correct formatting
- [ ] Favorites toggle (f) highlights symbol in accent color
- [ ] D/W filter (d) shows only coins with blocked paths
- [ ] All status bar help text matches actual keybindings per tab

---

## 12. Design Consistency Audit and Improvement Backlog

### 12.1 Consistency Checklist
- [x] Unify state color mapping across all regions: SUCCESS/PROGRESS/WARNING/ERROR
- [x] Standardize numeric format policy for price, amount, percent, and FX (`formatters.py`)
- [x] Standardize terminology: English-only (Domestic/Global/Wallet), zero Korean chars in code
- [ ] Standardize spacing primitives across components (TCSS files)
- [ ] Standardize hover/active/disabled contrast behavior (TCSS files)
- [x] Remove remaining Transfer-area hardcoded colors and map to semantic tokens
- [x] Kimchi premium color convention aligned: red = high (heat-map), blue = negative

### 12.2 Completed Backlog Items

| Item | Resolution | Files Changed |
|------|-----------|---------------|
| P1 Color token unification | All colors via `C.<TOKEN>`, zero raw hex in screens/widgets | `colors.py`, all screens |
| P1 Number formatting policy | 12 shared formatters in `formatters.py` | `formatters.py` |
| P1 Expansion interaction clarity | `e` key only (Enter is no-op) | `screens/monitor.py` |
| P2 Indicator consistency | 12 shared indicators in `indicators.py` | `indicators.py` |
| P2 Accessibility hardening | Icon+text patterns (✓/✗ + D/W, ● + Live/Stale) | `indicators.py` |
| P3 Transfer log readability | `log_level_badge()`: ERR/WRN/INF/DBG badges | `indicators.py` |
| P3 Localization consistency | All labels English (Domestic/Global/Wallet) | `screens/transfer.py` |

### 12.3 Remaining Backlog

| Priority | Item | Current Gap | Target |
|----------|------|-------------|--------|
| P2 | Header density | Single-line info bar can be crowded | Consider multi-line or collapsible header |
| P3 | TCSS spacing audit | Spacing defined per-widget | Centralize spacing tokens |
| P3 | DomGap% sortable | Not in SORT_COLUMNS | Add to CoinTable.SORT_COLUMNS |
| P3 | Okx sortable | Not in SORT_COLUMNS | Add to CoinTable.SORT_COLUMNS |
