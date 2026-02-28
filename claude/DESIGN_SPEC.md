# Kimchi Terminal — Visual Design Specification

## Executive Summary

This document provides a comprehensive redesign specification for the Kimchi Premium Monitor egui desktop application. The current app is **functionally solid** but visually generic. This spec transforms it into a **professional-grade trading terminal** with the visual density and intentionality of Bloomberg Terminal, the color clarity of TradingView, and the modern polish of Binance Pro — all adapted for the egui framework.

**Framework**: egui 0.31 (eframe) — Immediate mode GUI, Rust  
**Current state**: GitHub dark theme, flat cards, no charts/sparklines, no visual motion  
**Target state**: Bloomberg-meets-TradingView aesthetic with signal-to-noise optimization  

---

## Table of Contents

1. [Design Philosophy](#1-design-philosophy)
2. [Color Palette — "Terminal Noir"](#2-color-palette--terminal-noir)
3. [Typography & Number Formatting](#3-typography--number-formatting)
4. [Layout Architecture](#4-layout-architecture)
5. [Component Redesigns — Main Monitor](#5-component-redesigns--main-monitor)
6. [Component Redesigns — Coin Detail Inline](#6-component-redesigns--coin-detail-inline)
7. [Component Redesigns — Scenario Panel](#7-component-redesigns--scenario-panel)
8. [Component Redesigns — D/W Status](#8-component-redesigns--dw-status)
9. [Component Redesigns — Transfer Window](#9-component-redesigns--transfer-window)
10. [Motion & Animation](#10-motion--animation)
11. [egui Widgets & Crates to Leverage](#11-egui-widgets--crates-to-leverage)
12. [ASCII Mockups](#12-ascii-mockups)
13. [Priority-Ordered Implementation Plan](#13-priority-ordered-implementation-plan)
14. [Reference Projects](#14-reference-projects)

---

## 1. Design Philosophy

### Aesthetic Direction: "Signal Station"

Not a startup dashboard. Not a crypto bro app. This is a **professional monitoring station** — the kind of interface an operator sits in front of for 12 hours. Every pixel serves the mission: **find arbitrage opportunities faster than anyone else**.

**Core Principles:**
- **Data is the decoration.** Numbers, color-coded values, and status indicators ARE the visual design. No ornamental borders, no decorative gradients.
- **Scannable in under 1 second.** The most important metric (premium %) must be identifiable in peripheral vision.
- **Calm under pressure.** Muted backgrounds, controlled brightness. Color screams only when something demands attention.
- **Information density over whitespace.** More data visible = faster decisions. But density ≠ clutter — clear hierarchy separates them.

**Anti-patterns to avoid:**
- Generic card layouts with excessive rounded corners (current state)
- Equal visual weight on all information (everything bold = nothing bold)
- Decorative borders that consume space without adding meaning
- Pure white text on near-black background (eye strain after 2 hours)

---

## 2. Color Palette — "Terminal Noir"

### Why Change from GitHub Dark

The current GitHub dark palette (#0D1117 base) is **designed for reading code**, not monitoring financial data. It optimizes for syntax highlighting variety, not for the rapid green/red scanning that trading requires. The proposed "Terminal Noir" palette shifts toward:

- **Deeper, warmer blacks** (less blue-gray, more pure dark — reduces blue light)
- **Higher contrast on signal colors** (green/red pop more against the background)
- **A distinct accent color** (amber/gold) that separates this from every other dark theme
- **Graduated intensity system** for premium values (heat map effect)

### Base Palette

```
                    Current         Proposed         Rationale
                    ───────         ────────         ─────────
BG_DEEP            #0D1117         #0A0E14          Deeper, less blue cast. 
                                                     Reduces eye strain in 8h+ sessions.

BG_PANEL           #161B22         #111820          Warmer undertone. Slightly more 
                                                     separation from BG_DEEP.

BG_CARD            #1C2128         #161D26          Cards float above panels with 
                                                     just enough contrast.

BG_CARD_ALT        #161B22         #131A22          Alternating row background. 
                                                     Subtle stripe, not distracting.

BG_HOVER           #21262D         #1C2430          Mouse hover / keyboard focus.
                                                     Clear feedback without flash.

BG_SELECTED        #1F6FEB33       #D4940020        Selected row uses amber tint,
                                                     not blue — unique brand identity.

BORDER             #30363D         #1E2A36          Borders are QUIETER. They guide 
                                                     layout, not compete with data.
                                                     Reduced from 48→30 brightness.

BORDER_FOCUS       (none)          #D49400           Active panel/focus gets amber 
                                                     border — Bloomberg-style focus.
```

### Text Hierarchy

```
FG_PRIMARY         #E6EDF3         #D4DBE5          Pulled back from near-white.
                                                     93% → 84% brightness. Easier on
                                                     eyes, still fully readable.

FG_SECONDARY       #8B949E         #7A8694          Labels, timestamps, exchange names.
                                                     Clearly subordinate to primary.

FG_MUTED           #484F58         #3D4654          Placeholder text, disabled states,
                                                     divider text like "|".

FG_BRIGHT          (none)          #F0F4F8          Reserved for THE most important 
                                                     number on screen (e.g., tether
                                                     premium in header). Use sparingly.
```

### Signal Colors (The Money Palette)

These are the colors that convey meaning. They are the reason the app exists.

```
GREEN_STRONG       #3FB950         #10B981          TradingView emerald. 65% saturation.
                                                     Tested for 8h+ sessions. Less neon
                                                     than GitHub green.

GREEN_MUTED        (none)          #10B98150        50% opacity version for backgrounds
                                                     (e.g., flash-on-update overlay).

GREEN_BG           (none)          #10B98118        Background tint for positive rows.

RED_STRONG         #F85149         #EF4444          Standard financial red. Slightly 
                                                     desaturated from GitHub red.

RED_MUTED          (none)          #EF444450        50% opacity for backgrounds.

RED_BG             (none)          #EF444418        Background tint for negative rows.

AMBER              #D29922         #D49400          Warning + accent + brand color.
                                                     Warmer, more gold than yellow.
                                                     This is "the Kimchi color."

AMBER_MUTED        (none)          #D4940050        For pill backgrounds.

BLUE               #58A6FF         #3B82F6          Information, links, DOM-GAP scenario.
                                                     Slightly more saturated.

PURPLE             #BC8CFF         #A78BFA          Futures basis. Shifted toward 
                                                     violet for better distinction
                                                     from blue.
```

### Premium Heat Map (The Core Innovation)

This is the most important color system. Premium % is the #1 signal. It deserves a **graduated color ramp** that communicates magnitude at a glance:

```
PREMIUM RAMP (for pill/badge backgrounds):

  Negative (<0%):
    < -5%        #3B82F6  (strong blue — reverse kimchi, noteworthy)
    -5% to -2%   #3B82F680 (muted blue)
    -2% to 0%    #3D4654  (gray — non-event)

  Low positive (0% to 2%):
    0% to 1%     #3D4654  (gray — normal market conditions)
    1% to 2%     #4B6043  (gray-green hint — mild)

  Moderate (2% to 4%):
    2% to 3%     #10B981  (emerald — opportunity emerging)
    3% to 4%     #D49400  (amber — strong opportunity)

  High (4%+):
    4% to 5%     #E8760C  (orange — high alert)
    5% to 7%     #EF4444  (red — extreme)
    7%+          #DC2626  (deep red, pulsing — action required)
```

**Implementation:**
```rust
fn premium_pill_color(pct: f64) -> Color32 {
    if pct < -5.0      { Color32::from_rgb(59, 130, 246) }     // strong blue
    else if pct < -2.0 { Color32::from_rgba_unmultiplied(59, 130, 246, 128) }
    else if pct < 0.0  { Color32::from_rgb(61, 70, 84) }       // gray
    else if pct < 1.0  { Color32::from_rgb(61, 70, 84) }       // gray
    else if pct < 2.0  { Color32::from_rgb(75, 96, 67) }       // gray-green
    else if pct < 3.0  { Color32::from_rgb(16, 185, 129) }     // emerald
    else if pct < 4.0  { Color32::from_rgb(212, 148, 0) }      // amber
    else if pct < 5.0  { Color32::from_rgb(232, 118, 12) }     // orange
    else if pct < 7.0  { Color32::from_rgb(239, 68, 68) }      // red
    else               { Color32::from_rgb(220, 38, 38) }      // deep red
}
```

### Exchange Identity Colors

Each exchange gets a unique color. These are used for exchange pill/tag backgrounds and should be **muted enough** to not compete with signal colors:

```
UPBIT              #3B82F6  (blue — matching their brand)
BITHUMB            #A78BFA  (violet — distinct from Upbit blue)
BINANCE            #D49400  (gold/amber — matching their brand yellow)
BYBIT              #EF4444  (red — close to their orange brand)
OKX                #10B981  (green — distinct from others)
BITGET             #06B6D4  (cyan — fresh, distinct)
GATE               #8B5CF6  (purple — distinct from Bithumb violet)
```

---

## 3. Typography & Number Formatting

### Font Strategy

**Current:** AppleSDGothicNeo as primary (Korean support), system proportional for everything.

**Proposed:** Keep Korean font fallback but add a dedicated **monospace font for ALL numbers**.

```rust
pub fn setup_fonts(ctx: &egui::Context) {
    let mut fonts = egui::FontDefinitions::default();

    // 1. Load JetBrains Mono for numbers (bundle in assets/)
    //    OR use system monospace as baseline
    // JetBrains Mono has excellent number legibility:
    //  - Distinct 0/O, 1/l/I
    //  - Consistent digit width (critical for scanning columns)
    //  - Excellent at small sizes (10-12px)
    
    // 2. Korean font as fallback for Proportional
    // (keep existing AppleSDGothicNeo loading)
    
    // 3. Register a custom "Numbers" family if desired
    // Or simply use Monospace for all RichText that displays prices/percentages
}
```

**Key rule: Every number in the app must use monospace.** This ensures:
- Decimal points align vertically across rows
- Price changes don't cause text to jump left/right
- Columns of numbers are scannable as a unit

### Number Formatting Rules

```
PRICES (KRW):
  ≥ 1,000,000   →  "₩1,234,567"    (comma separated, no decimals)
  ≥ 1,000       →  "₩1,234"        (comma separated, no decimals)
  ≥ 1           →  "₩1.23"         (2 decimals)
  < 1            →  "₩0.1234"       (4 decimals)

PRICES (USD):
  ≥ 10,000      →  "$12,345"       (comma, no decimals)
  ≥ 100         →  "$123.45"       (2 decimals)
  ≥ 1           →  "$1.2345"       (4 decimals)
  < 1            →  "$0.001234"     (6 decimals)

PERCENTAGES:
  Always show sign:  "+2.34%"  or  "-1.56%"
  Always 2 decimals: "+0.00%"  not  "0%"
  
  Current format_kimchi() is correct. Keep it.

LARGE NUMBERS (header/summary only):
  Budget slider:     "200만"  →  "200만" (keep, this is idiomatic Korean)
  Coin count:        "285 coins" (keep)

EXCHANGE RATES:
  USDT/KRW:  "₩1,384.50"  (always 2 decimals for forex)
  USD/KRW:   "₩1,380.25"  (always 2 decimals)
```

### Spacing & Alignment

```
Number columns:     Right-aligned (prices, percentages, amounts)
Text columns:       Left-aligned (symbols, names, exchange labels)
Pill/badge:         Center-aligned text within the pill

Minimum column widths:
  Symbol:           56px  (accommodates "SHIB1000")
  Korean name:      72px  (accommodates 4 Korean characters)
  Price (KRW):      96px  (accommodates "₩1,234,567,890")
  Price (USD):      80px  (accommodates "($0.001234)")
  Premium %:        72px  (accommodates pill with "+12.34%")
  Futures basis:    72px  (same as premium)
```

---

## 4. Layout Architecture

### Current Layout (Analyzed)

```
┌─────────────────────────────────────────────────────────────────────┐
│ HEADER: title + tether premium pill + USD/USDT rates + coin count  │
│ TOOLBAR: search + budget slider + exchange A/B + dark/light + xfer │
├──────────────────────────────────────┬──────────────────────────────┤
│                                      │ Scenarios (right panel)     │
│  MAIN: Coin card list                │  - Filter pills             │
│   - Each card: symbol, kr_name,      │  - Threshold drag values    │
│     prices, premium pill,            │  - Thread list (scrollable) │
│     futures basis pill               │                             │
│   - Click expands inline detail      │──────────────────────────── │
│   - Virtual scrolling (>30 coins)    │ D/W Status                  │
│                                      │  - Blocked coin list        │
│                                      │  - Per-exchange status      │
├──────────────────────────────────────┴──────────────────────────────┤
│ INLINE TRANSFER (bottom panel, when open)                          │
└─────────────────────────────────────────────────────────────────────┘
```

### Proposed Layout Improvements

**The layout structure is fundamentally sound.** The main improvements are within components, not in moving panels around. Here's what changes:

1. **Header bar gets a visual upgrade** — ticker tape feel, more prominent tether premium
2. **Column headers become sticky** — always visible above the scrolling coin list
3. **Coin cards become denser rows** — less card-like, more table-row-like
4. **Right panel gets collapsible sections** — scenarios and D/W can expand/collapse independently
5. **Status bar added at bottom** — keybindings, connection status, last update time

---

## 5. Component Redesigns — Main Monitor

### A. Header Bar Redesign

**Current:** Flat horizontal bar with small text.  
**Proposed:** Two-tier header with visual prominence hierarchy.

```
┌─────────────────────────────────────────────────────────────────────┐
│ KIMCHI TERMINAL                                                     │
│                                                                     │
│  ┌─────────────────┐  USDT ₩1,384.50    USD ₩1,380.25   285 coins │
│  │ Tether  +3.42%  │                                               │
│  └─────────────────┘                                               │
│                                                                     │
│  🔍 [Search____________]  Budget [====200만====]                    │
│  A: [Upbit ▼]  ⇄  B: [Binance ▼]     ☀/🌙   [Transfer]          │
└─────────────────────────────────────────────────────────────────────┘
```

**Key changes:**
- Tether premium pill is **2x larger** than current — it's the single most important number on the entire screen
- Use `FG_BRIGHT` (#F0F4F8) for the tether premium number, **bold, 20px**
- Pill background uses the premium heat map color (so if tether premium is +3.42%, the pill is amber)
- Exchange rate values use monospace, `FG_PRIMARY`
- Title "KIMCHI TERMINAL" in `FG_MUTED` small-caps — it's branding, not information
- Separate "info row" (rates, coin count) from "control row" (search, selectors)

**Implementation detail — Tether premium pill:**
```rust
// Make the tether premium the visual anchor of the entire app
let pill_bg = premium_pill_color(tether_prem);
let pill_size = 20.0; // vs current 17.0
egui::Frame::new()
    .fill(pill_bg)
    .inner_margin(egui::Margin::symmetric(16, 6))  // more padding
    .corner_radius(8.0)  // larger radius for larger pill
    .show(ui, |ui| {
        ui.label(
            egui::RichText::new(format!("Tether {}{:.2}%", 
                if tp >= 0.0 { "+" } else { "" }, tp))
                .size(pill_size)
                .strong()
                .color(Color32::from_rgb(0x0A, 0x0E, 0x14))  // dark text on colored bg
        );
    });
```

### B. Coin List Redesign

**Current:** Card-based layout with alternating backgrounds, ~78px row height.  
**Proposed:** Dense table rows with information hierarchy, ~52px row height.

The current card approach wastes vertical space. With 285 coins, the user sees ~11 cards at a time. By tightening to table rows, we can show ~17 rows — a 55% improvement in visible data.

**Row structure (2-line compact):**
```
┌──────────────────────────────────────────────────────────────────┐
│ BTC 비트코인   UP ₩97,234,567  |  BN ₩96,500,000   [+0.76%] [BNF +0.12%] │
│                ($73,456.78)       ($72,901.23)     ●● slip 0.3% │
└──────────────────────────────────────────────────────────────────┘
```

**Line 1 (primary):**
- Symbol (bold, 14px, FG_PRIMARY)
- Korean name (12px, FG_SECONDARY)  
- Exchange A pill + KRW price (monospace, 12px)
- Separator "|"
- Exchange B pill + KRW price (monospace, 12px)
- Premium % pill (bold, 12px, heat-mapped background)
- Futures basis pill (11px, if available)

**Line 2 (secondary, muted):**
- USD prices in parentheses (11px, FG_SECONDARY, monospace)
- D/W status dots (●● green/red)
- Slippage info (11px, FG_MUTED)

**Key differences from current:**
1. **Two lines, not three** — slippage moves to line 2 instead of its own line
2. **Prices are right-aligned in their columns** — numbers stack vertically
3. **Premium pill uses heat map gradient** — not just green/amber/red, but the full 10-step ramp
4. **D/W dots are inline** on line 2, not requiring a hover or separate panel check
5. **Row height: 52px** (down from ~78px) = 55% more visible coins

**Alternating row pattern:**
```rust
let bg = if i % 2 == 0 { BG_CARD } else { BG_CARD_ALT };
// Difference between these two should be BARELY perceptible
// Current: #1C2128 vs #161B22 (delta=6) — good
// Proposed: #161D26 vs #131A22 (delta=3) — even subtler
```

### C. Column Headers (New — Currently Missing)

**Add sticky column headers** above the scrolling coin list:

```
┌──────┬────────┬─────────────────┬─────────────────┬─────────┬────────┐
│ Coin │ Korean │ Upbit (A)    ▼ │ Binance (B)  ▼ │ Gap% ▼  │ FUT%   │
└──────┴────────┴─────────────────┴─────────────────┴─────────┴────────┘
```

- Use `FG_SECONDARY` for header text, `BG_PANEL` background
- Sort indicator arrows use `FG_MUTED` (current is correct)
- Headers should be rendered **outside the ScrollArea** so they stay visible
- Add a subtle 1px `BORDER` line below headers

**Implementation:**
```rust
// Render column headers BEFORE the ScrollArea
ui.horizontal(|ui| {
    // ... header buttons with sort indicators
});
// Thin separator
let rect = ui.max_rect();
ui.painter().line_segment(
    [rect.left_bottom(), rect.right_bottom()],
    egui::Stroke::new(1.0, BORDER),
);
// THEN the scroll area with coin rows
egui::ScrollArea::vertical()
    .auto_shrink([false, false])
    .show(ui, |ui| { /* coin rows */ });
```

This is actually already done in the codebase! The headers render before the scroll area. But visually they need more distinction:
- Add a subtle background fill to the header row
- Increase bottom margin between headers and first row
- Make the active sort column header slightly brighter

---

## 6. Component Redesigns — Coin Detail Inline

**Current:** Expands below coin card with grid of exchanges, prices, buy/sell buttons.  
**Proposed:** Cleaner layout with visual separation and progress tracking.

```
┌─────────────────────────────────────────────────────────────────────┐
│ ▼ BTC Market Order                                     [↻ Refresh] │
│─────────────────────────────────────────────────────────────────────│
│                                                                     │
│  Exchange    Price              Amount    Actions     Balances      │
│  ┌────┐                                                            │
│  │ UP │ ₩97,234,567            [____$__]  [Buy] [Sell]  0.1234 BTC │
│  └────┘                                                 $342.10    │
│  ┌────┐                                                            │
│  │ BN │ ₩96,500,000 ($73,456)  [____$__]  [Buy] [Sell]  0.0500 BTC │
│  └────┘                                                 $1,234.56  │
│  ...                                                               │
│                                                                     │
│  Quick: [25%] [50%] [75%] [100%]    ☑ Auto-sell on arrival         │
│                                                                     │
│  Transfer Progress ─────────────────────────────────────────────── │
│  Job #1: BN→UP 0.05 BTC via TRC20                                 │
│  [████████████░░░░░░░░] Step 4/6: Confirming deposit...            │
│─────────────────────────────────────────────────────────────────────│
│  ⚠ Confirm: BUY $500.00 USDT of BTC on Binance?   [Confirm] [✕]  │
└─────────────────────────────────────────────────────────────────────┘
```

**Key improvements:**

1. **Transfer progress gets a real progress bar** (not just text status):
```rust
// 6-step progress bar
let steps = ["Balance", "Withdraw", "Tx Sent", "Confirming", "Deposited", "Complete"];
let current_step = 3; // 0-indexed
let progress = (current_step + 1) as f32 / steps.len() as f32;

// Draw background track
let track_rect = Rect::from_min_size(pos, vec2(total_width, 6.0));
painter.rect_filled(track_rect, 3.0, BG_HOVER);

// Draw filled portion with gradient
let fill_rect = Rect::from_min_size(pos, vec2(total_width * progress, 6.0));
painter.rect_filled(fill_rect, 3.0, GREEN_STRONG);

// Draw step markers
for i in 0..steps.len() {
    let x = pos.x + (i as f32 / (steps.len() - 1) as f32) * total_width;
    let dot_color = if i <= current_step { GREEN_STRONG } else { FG_MUTED };
    painter.circle_filled(pos2(x, pos.y + 3.0), 4.0, dot_color);
}
```

2. **Confirmation dialog uses stronger visual warning:**
```rust
// Amber background with border for confirmation
egui::Frame::new()
    .fill(Color32::from_rgba_unmultiplied(212, 148, 0, 20))  // AMBER_BG
    .stroke(egui::Stroke::new(1.0, AMBER))
    .corner_radius(4.0)
    .inner_margin(8.0)
    .show(ui, |ui| { /* confirmation content */ });
```

3. **Buy/Sell buttons get more visual distinction:**
```rust
// Buy: filled green
let buy_btn = egui::Button::new(
    egui::RichText::new("Buy").size(11.0).strong().color(BG_DEEP))
    .fill(GREEN_STRONG)
    .corner_radius(3.0);

// Sell: filled red  
let sell_btn = egui::Button::new(
    egui::RichText::new("Sell").size(11.0).strong().color(BG_DEEP))
    .fill(RED_STRONG)
    .corner_radius(3.0);
```

---

## 7. Component Redesigns — Scenario Panel

**Current:** Right sidebar with thread cards showing scenario type pills, value pills, expand/collapse.  
**Proposed:** Tighter layout with visual urgency signals.

### Active Thread Design

```
┌─ SCENARIOS ──────────────────── 3 active / 12 total ─┐
│                                                       │
│ [KIMP] 5.0%▼  [DOM-GAP] 1.5%▼  [FUT%] 0.50%▼       │
│─────────────────────────────────────────────────────── │
│ ▼3  14:23:05  KIMP  BTC: 비트코인  [+5.23%]          │
│     UP-BN kimp crossed ▲5.0% → 5.23%                 │
│     ├ 14:22:01  kimp 4.8% (initial)                   │
│     ├ 14:22:30  kimp 5.1% (+0.3%p)                    │
│     └ 14:23:05  kimp 5.23% (+0.43%p)                  │
│                                                       │
│ ▶1  14:20:15  DOM   ETH: 이더리움  [+1.82%]          │
│     UP-BT gap +1.82% (UP 2,100,000 > BT 2,062,000)   │
│                                                       │
│ ●   14:18:32  FUT%  SOL: 솔라나   [+0.65%]           │  ← no sub-entries
│     basis +0.65% (spot > futures)                      │
│                                                       │
│ ─── Closed ───────────────────────────────────────── │
│     14:15:00  KIMP  XRP  [-0.2%]  (closed)            │
└───────────────────────────────────────────────────────┘
```

**Key improvements:**

1. **Left border color bar** (already exists, keep it — it's the best pattern)
2. **Closed threads get visually suppressed:**
   - Text becomes `FG_MUTED`
   - Background slightly darker than card bg
   - Grouped under a "Closed" separator
3. **Active thread premium pill pulses** when value is extreme (>5%):
```rust
if thread.is_active && thread.last_logged_value.abs() > 5.0 {
    let t = (ui.ctx().input(|i| i.time) * 2.0).sin() as f32 * 0.5 + 0.5;
    let alpha = (180.0 + 75.0 * t) as u8;
    pill_bg = pill_bg.with_alpha(alpha);
}
```
4. **Click-to-jump** works great already. Add a subtle hover effect on the symbol button.

---

## 8. Component Redesigns — D/W Status

**Current:** Simple list of blocked coins with exchange-level deposit/withdraw indicators.  
**Proposed:** Compact grid view that's scannable at a glance.

### Grid Layout for D/W Status

```
┌─ D/W STATUS ─── 7 blocked ──────────────────────────┐
│                                                       │
│ Coin    UP    BT    BN    BB    OK                    │
│ ────────────────────────────────────────────────────  │
│ BTC     ●●    ●●    ●○    ●●    ●●                   │
│ ETH     ●●    ○●    ●●    ●●    ●○                   │
│ SOL     ●○    ●●    ●●    ○○    ●●                   │
│ XRP     ●●    ●●    ●●    ●●    ●●  ← all ok, why   │
│ ...                                     shown? Filter!│
│                                                       │
│ Legend: ● ok  ○ blocked  (D=deposit, W=withdraw)      │
│         First dot = D, Second dot = W                 │
└───────────────────────────────────────────────────────┘
```

**Key improvements:**

1. **Grid format instead of list** — shows all exchanges for a coin in one row
2. **Only show blocked coins** (current behavior, keep it)
3. **Two-dot notation**: `●○` means deposit OK, withdraw blocked
4. **Color coding**: Green dot = ok, Red dot = blocked
5. **Clicking a coin filters the main list** (already works)

**Implementation:**
```rust
// Compact D/W grid
egui::Grid::new("dw_grid")
    .num_columns(6) // coin + 5 exchanges
    .spacing([6.0, 2.0])
    .show(ui, |ui| {
        // Header
        ui.label(RichText::new("Coin").size(10.0).color(FG_SECONDARY));
        for ex_label in ["UP", "BT", "BN", "BB", "OK"] {
            ui.label(RichText::new(ex_label).size(10.0).color(FG_SECONDARY));
        }
        ui.end_row();
        
        // Rows
        for (symbol, cws) in &blocked_coins {
            if ui.small_button(symbol).clicked() { /* filter */ }
            for getter in exchange_getters {
                if let Some(es) = getter(cws) {
                    let d_color = if es.deposit { GREEN_STRONG } else { RED_STRONG };
                    let w_color = if es.withdraw { GREEN_STRONG } else { RED_STRONG };
                    ui.horizontal(|ui| {
                        ui.label(RichText::new("●").size(8.0).color(d_color));
                        ui.label(RichText::new("●").size(8.0).color(w_color));
                    });
                } else {
                    ui.label(RichText::new("--").size(8.0).color(FG_MUTED));
                }
            }
            ui.end_row();
        }
    });
```

---

## 9. Component Redesigns — Transfer Window

The transfer window is already well-structured. Key improvements:

### Progress Visualization

Replace text-only step status with **visual step indicator:**

```
Transfer: BN → UP  |  0.5 BTC via TRC20  |  Fee: 1.0 USDT

  ① ──── ② ──── ③ ──── ④ ──── ⑤ ──── ⑥
  Bal    W/D    TxSent  Conf   Dep    Done
  ✓      ✓      ✓      ●      ○      ○

  Step 4: Confirming on-chain... (2/12 confirmations)
  Tx: 0x1234...5678  [View on Explorer ↗]
```

- Completed steps: Green circle with checkmark
- Current step: Pulsing amber circle
- Pending steps: Gray outline circle
- Failed step: Red circle with X

### Network Selection Enhancement

Current chain buttons are good. Add:
- **Fee comparison**: Show fee in USD equivalent next to each network
- **Speed estimate**: "~5 min" next to network name
- **Recommended badge**: Auto-highlight the cheapest network

---

## 10. Motion & Animation

### Flash-on-Update (Priority 1)

When a coin's premium changes, flash the row:

```rust
// In your CoinState or a parallel HashMap<String, Instant>:
struct FlashState {
    last_premium_change: Instant,
    last_premium_value: f64,
    direction: FlashDirection, // Up or Down
}

enum FlashDirection { Up, Down }

// During render:
let elapsed = flash.last_premium_change.elapsed().as_secs_f32();
let flash_duration = 0.6; // 600ms

if elapsed < flash_duration {
    let t = elapsed / flash_duration;
    let alpha = ((1.0 - t) * 40.0) as u8; // fade from 40 to 0 alpha
    
    let flash_color = match flash.direction {
        FlashDirection::Up => Color32::from_rgba_unmultiplied(16, 185, 129, alpha),
        FlashDirection::Down => Color32::from_rgba_unmultiplied(239, 68, 68, alpha),
    };
    
    // Paint flash overlay on the row background
    painter.rect_filled(row_rect, 0.0, flash_color);
    
    // Request repaint to animate
    ctx.request_repaint();
}
```

### Smooth Value Interpolation (Priority 2)

For the tether premium pill, interpolate the displayed value:

```rust
struct AnimatedPremium {
    displayed: f64,
    target: f64,
}

impl AnimatedPremium {
    fn update(&mut self, dt: f32) {
        let speed = 8.0; // higher = snappier
        self.displayed += (self.target - self.displayed) * speed as f64 * dt as f64;
    }
}
```

### Spinner for Loading States (Priority 3)

Use egui's built-in spinner for balance loading:

```rust
if bal.loading {
    ui.spinner(); // egui built-in, matches theme automatically
} else {
    ui.label(format!("{:.4}", coin_avail));
}
```

---

## 11. egui Widgets & Crates to Leverage

### Already Available in egui (No Extra Dependencies)

| Widget | Use Case | Current Usage | Recommendation |
|--------|----------|---------------|----------------|
| `Spinner` | Loading indicators | Not used | Use for balance loading |
| `ProgressBar` | Transfer steps | Not used | Use for 6-step state machine |
| `Separator` | Section dividers | `ui.separator()` | Replace some with custom painted lines |
| `CollapsingHeader` | Collapsible sections | Not used | Use for D/W Status, Scenario sections |
| `Grid` | Tabular data | Used in detail view | Use for D/W status grid |
| `ScrollArea::stick_to_bottom()` | Auto-scroll | Used in scenarios | Keep |
| Custom painting (`Painter`) | Sparklines, indicators | Not used | **High priority** for inline sparklines |

### Recommended New Dependencies

```toml
# In Cargo.toml [dependencies]:

# 1. CHART/PLOT — Add inline price sparklines
egui_plot = "0.31"  # Matches your egui version
# Use: Tiny inline Line plots showing 24h price trend per coin
# Implementation: Plot::new("spark_BTC").height(20.0).width(60.0)
#   .show_axes([false, false]).show_grid(false)
#   ... minimal sparkline

# 2. VIRTUAL SCROLL (if egui's built-in isn't sufficient)
# egui_virtual_list = "0.5"  # Optional — egui ScrollArea + manual 
#                             # virtualization (already implemented) may suffice

# 3. NOTIFICATIONS / TOASTS
# egui-notify = "0.17"  # Pop-up notifications for transfer completion,
#                        # scenario alerts, errors
```

### Custom Widgets to Build

1. **Sparkline Widget** — Tiny 60x20px inline chart showing 24h price trend:
```rust
fn render_sparkline(ui: &mut egui::Ui, data: &[f64], width: f32, height: f32) {
    let (response, painter) = ui.allocate_painter(
        egui::vec2(width, height), egui::Sense::hover());
    
    if data.len() < 2 { return; }
    let min = data.iter().cloned().fold(f64::MAX, f64::min);
    let max = data.iter().cloned().fold(f64::MIN, f64::max);
    let range = (max - min).max(0.0001);
    
    let rect = response.rect;
    let points: Vec<egui::Pos2> = data.iter().enumerate().map(|(i, v)| {
        let x = rect.left() + (i as f32 / (data.len() - 1) as f32) * rect.width();
        let y = rect.bottom() - ((v - min) as f32 / range as f32) * rect.height();
        egui::pos2(x, y)
    }).collect();
    
    // Determine trend color
    let color = if data.last() >= data.first() { GREEN_STRONG } else { RED_STRONG };
    
    // Draw line
    for window in points.windows(2) {
        painter.line_segment([window[0], window[1]], 
            egui::Stroke::new(1.5, color));
    }
}
```

2. **Heat Pill Widget** — Colored badge with premium value:
```rust
fn heat_pill(ui: &mut egui::Ui, value: f64) {
    let bg = premium_pill_color(value);
    let text = format!("{}{:.2}%", if value >= 0.0 { "+" } else { "" }, value);
    // Dark text on light pills, light text on dark pills
    let text_color = if value.abs() < 1.0 { FG_SECONDARY } else { BG_DEEP };
    
    egui::Frame::new()
        .fill(bg)
        .corner_radius(4.0)
        .inner_margin(egui::Margin::symmetric(8, 2))
        .show(ui, |ui| {
            ui.label(egui::RichText::new(text)
                .size(12.0).strong().monospace().color(text_color));
        });
}
```

3. **Step Progress Widget** — For transfer state machine:
```rust
fn step_progress(ui: &mut egui::Ui, current: usize, total: usize, labels: &[&str]) {
    let (response, painter) = ui.allocate_painter(
        egui::vec2(ui.available_width(), 30.0), egui::Sense::hover());
    let rect = response.rect;
    let step_width = rect.width() / (total as f32 - 1.0);
    
    // Draw connecting lines
    for i in 0..total-1 {
        let x1 = rect.left() + i as f32 * step_width;
        let x2 = x1 + step_width;
        let y = rect.top() + 8.0;
        let color = if i < current { GREEN_STRONG } else { FG_MUTED };
        painter.line_segment([pos2(x1, y), pos2(x2, y)], 
            egui::Stroke::new(2.0, color));
    }
    
    // Draw step dots and labels
    for i in 0..total {
        let x = rect.left() + i as f32 * step_width;
        let y = rect.top() + 8.0;
        
        let (fill, radius) = if i < current {
            (GREEN_STRONG, 5.0) // completed
        } else if i == current {
            (AMBER, 6.0)       // active (slightly larger)
        } else {
            (FG_MUTED, 4.0)    // pending
        };
        
        painter.circle_filled(pos2(x, y), radius, fill);
        
        // Label below
        if i < labels.len() {
            painter.text(pos2(x, y + 12.0), egui::Align2::CENTER_TOP,
                labels[i], egui::FontId::new(9.0, egui::FontFamily::Proportional),
                FG_SECONDARY);
        }
    }
}
```

---

## 12. ASCII Mockups

### Main Monitor View (Full Window)

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ KIMCHI TERMINAL                                                                     ┃
┃  ╔═══════════════════╗   USDT ₩1,384.50   USD ₩1,380.25              285 coins  ●  ┃
┃  ║ Tether  +3.42%    ║                                                LIVE         ┃
┃  ╚═══════════════════╝                                                              ┃
┃  🔍 [Search___________]  Budget [═══════200만═══════]                               ┃
┃  A: [Upbit  ▼]  ⇄  B: [Binance ▼]                          [☀ Light]  [Transfer]  ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃  Coin    Korean  UP Price        BN Price       Gap%▼  FUT%  ┃  SCENARIOS  3/12     ┃
┃ ─────────────────────────────────────────────────────────── ┃                       ┃
┃ ▸BTC 비트코인  UP ₩97,234,567  BN ₩96,500,000  +0.76%  +0.12 ┃ [KIMP▼5.0] [DOM▼1.5] ┃
┃               ($73,456)        ($72,901)       ●● 0.3% ┃ [FUT%▼0.50]           ┃
┃ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┃──────────────────────┃
┃  ETH 이더리움  UP ₩2,134,567   BN ₩2,100,000   +1.65%  -0.05 ┃ █14:23 KIMP BTC +5.2 ┃
┃               ($1,567)         ($1,542)        ●● 0.1% ┃   UP-BN crossed ▲5.0% ┃
┃ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┃   ├ 14:22 kimp 4.8%  ┃
┃  SOL 솔라나   UP ₩234,567     BN ₩228,000      +2.88%  +0.65 ┃   └ 14:23 kimp 5.23% ┃
┃               ($172.34)        ($167.50)       ●○ 0.8% ┃                       ┃
┃ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┃ █14:20 DOM  ETH +1.8 ┃
┃  XRP 리플     UP ₩1,234       BN ₩1,200        +2.83%   --  ┃   BT-UP gap +1.82%  ┃
┃               ($0.9012)        ($0.8801)       ●● 0.2% ┃                       ┃
┃ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┃ ●14:18 FUT% SOL +0.6 ┃
┃  DOGE 도지    UP ₩234         BN ₩229           +2.18%   --  ┃   basis widened +0.65┃
┃               ($0.1712)        ($0.1676)       ●● 0.5% ┃                       ┃
┃ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┃─── Closed ──────────┃
┃  ADA 에이다   UP ₩567         BN ₩560           +1.25%   --  ┃   14:15 KIMP XRP -0.2┃
┃               ($0.4156)        ($0.4103)       ●● 0.1% ┃                       ┃
┃ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┃══════════════════════┃
┃  ...  (scrollable, 285 total)                           ┃  D/W STATUS  7 blocked┃
┃                                                         ┃  Coin  UP BT BN BB OK ┃
┃                                                         ┃  BTC   ●● ●● ●○ ●● ●● ┃
┃                                                         ┃  SOL   ●○ ●● ●● ○○ ●● ┃
┃                                                         ┃  ...                   ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ ◉ Connected  |  Last update: 0.1s ago  |  ↑↓ Scroll  |  ⌘F Search  |  ⌘T Transfer ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Coin Detail Expanded View

```
┃ ▸BTC 비트코인  UP ₩97,234,567  BN ₩96,500,000  [+0.76%]  [BNF +0.12%]            ┃
┃               ($73,456)        ($72,901)        ●● 0.3%                           ┃
┃ ┌─────────────────────────────────────────────────────────────────────────────────┐ ┃
┃ │ BTC Market Order                                                [↻ Refresh]    │ ┃
┃ │                                                                                │ ┃
┃ │  Exchange    Price                  Amount       Actions       Balances         │ ┃
┃ │  UP  ₩97,234,567                   [$________]  [Buy] [Sell]  0.1234  $342.10  │ ┃
┃ │  BT  ₩96,800,000                   [$________]  [Buy] [Sell]  0.0000  $0.00    │ ┃
┃ │  BN  ₩96,500,000 ($73,456.78)      [$________]  [Buy] [Sell]  0.0500  $1234.56 │ ┃
┃ │  BB  ₩96,400,000 ($73,380.12)      [$________]  [Buy] [Sell]  0.0000  $567.89  │ ┃
┃ │  OK  ₩96,450,000 ($73,418.00)      [$________]  [Buy] [Sell]  0.0000  $0.00    │ ┃
┃ │                                                                                │ ┃
┃ │  Quick: [25%] [50%] [75%] [100%]              ☑ Auto-sell on arrival           │ ┃
┃ │                                                                                │ ┃
┃ │  Transfer Progress                                                             │ ┃
┃ │  BN→UP 0.05 BTC TRC20                                                         │ ┃
┃ │  ●────●────●────◉────○────○                                                    │ ┃
┃ │  Bal   W/D  TxSent Conf  Dep  Done    Step 4/6: 2/12 confirmations             │ ┃
┃ └─────────────────────────────────────────────────────────────────────────────────┘ ┃
```

---

## 13. Priority-Ordered Implementation Plan

### Phase 1: Quick Wins (1-2 days, high impact)

| # | Change | Impact | Effort | File |
|---|--------|--------|--------|------|
| 1 | **Update color palette** to Terminal Noir | High — immediate visual upgrade | Low | `app.rs` ThemeColors |
| 2 | **Enlarge tether premium pill** (20px, heat-mapped) | High — visual anchor established | Low | `app.rs` header render |
| 3 | **Premium pills use 10-step heat ramp** | High — scannable at a glance | Low | `app.rs` get_kimchi_color |
| 4 | **Monospace for all numbers** | Medium — alignment, professionalism | Low | `app.rs` all RichText with prices |
| 5 | **Reduce row height** to ~52px (tighter card layout) | High — 55% more visible data | Medium | `app.rs` row_height, margins |
| 6 | **Quieter borders** (#1E2A36 vs current #30363D) | Medium — less visual noise | Low | `app.rs` ThemeColors |
| 7 | **Add status bar** at bottom (connection, last update, keys) | Low-Medium — professional feel | Low | `app.rs` new TopBottomPanel::bottom |

### Phase 2: Structural Improvements (3-5 days, medium impact)

| # | Change | Impact | Effort | File |
|---|--------|--------|--------|------|
| 8 | **D/W grid layout** (replace list with table grid) | Medium — much faster scanning | Medium | `app.rs` D/W section |
| 9 | **Transfer progress bar** (visual step indicator) | Medium — professional transfer UX | Medium | `transfer/window.rs` |
| 10 | **Flash-on-update animation** for price changes | High — real-time feel | Medium | `app.rs` + new flash state |
| 11 | **Collapsible sections** for Scenarios / D/W | Medium — space management | Low | `app.rs` right panel |
| 12 | **Closed scenarios visually suppressed** | Low — cleaner scenario view | Low | `app.rs` scenario render |
| 13 | **Sticky column headers** (visual distinction) | Medium — better table UX | Low | `app.rs` header render |

### Phase 3: Advanced Polish (1-2 weeks, high polish)

| # | Change | Impact | Effort | File |
|---|--------|--------|--------|------|
| 14 | **Add egui_plot** for sparklines per coin | High — biggest visual differentiator | High | New dependency, data collection |
| 15 | **Price history ring buffer** (collect 24h data per coin) | Prerequisite for sparklines | High | `models/`, `monitor/` |
| 16 | **Animated tether premium** (smooth number interpolation) | Medium — premium feel | Low | `app.rs` |
| 17 | **Toast notifications** (egui-notify) for scenario alerts | Medium — awareness without distraction | Medium | New dependency |
| 18 | **Keyboard shortcuts** (⌘F search, ⌘T transfer, j/k scroll) | Medium — power user experience | Medium | `app.rs` input handling |
| 19 | **Custom font bundle** (JetBrains Mono for numbers) | Medium — number legibility | Medium | `app.rs` setup_fonts |
| 20 | **Sound alerts** for extreme premium (>7%) | Low — optional power feature | Medium | New module |

### Phase 4: Future Vision (exploration)

| # | Change | Impact | Effort |
|---|--------|--------|--------|
| 21 | **Mini orderbook depth visualization** in coin detail | Very High | Very High |
| 22 | **Arbitrage P&L tracker** with historical chart | Very High | Very High |
| 23 | **Multi-monitor support** (undock panels to separate viewports) | High | High |
| 24 | **Configurable column layout** (user chooses which columns to show) | Medium | Medium |

---

## 14. Reference Projects

### Must-Study Repositories

| Project | URL | What to Learn |
|---------|-----|---------------|
| **tickrs** | github.com/tarkah/tickrs | Inline sparklines, financial data formatting, real-time charts |
| **bottom** | github.com/ClementTsang/bottom | Data-dense layouts, theme system, 60fps rendering with 100s of data points |
| **rerun** | github.com/rerun-io/rerun | Best egui theming in production — `design_tokens.rs` is a masterclass |
| **egui_plot** | github.com/emilk/egui_plot | Line plots, heatmaps, gradient fills — all applicable to sparklines |
| **hello_egui** | github.com/lucasmerlin/hello_egui | Virtual scrolling, flexbox, drag-drop — advanced egui patterns |
| **Hedge UI** | hedgeui.com | React trading UI — study their layout grid, dark theme, component catalog |

### Color Palette References

| Platform | Background | Text | Green | Red | Accent |
|----------|-----------|------|-------|-----|--------|
| Bloomberg | #000000 | #00FF00 | #00FF00 | #FF0000 | #0068FF |
| TradingView | #131722 | #D1D5DB | #26A69A | #F6465D | #2962FF |
| Binance | #0F0F0F | #FFFFFF | #0ECB81 | #F6465D | #F0B90B |
| Coinbase Pro | #1A1A2E | #E0E0E0 | #00B300 | #FF3300 | #0052FF |
| **Kimchi (proposed)** | **#0A0E14** | **#D4DBE5** | **#10B981** | **#EF4444** | **#D49400** |

### Dark Mode Chart Best Practices (from research)

- Background: #1E1E1E range (not pure black, not gray)  
- Text: #E0E0E0 (not pure white — reduces glare)
- Grid lines: 10-15% opacity of text color
- Colors: De-saturated to 65% (prevents bloom/afterimage)
- Contrast ratio: Minimum 4.5:1 for readability (WCAG AA)

---

## Summary

The Kimchi Terminal is a powerful tool with an excellent functional foundation. The visual improvements proposed here follow a clear philosophy: **the data IS the design**. By implementing the Terminal Noir palette, the premium heat map, flash-on-update animations, and inline sparklines, the app transforms from "functional monitoring tool" to "professional trading terminal" — the kind of interface that makes users feel they have an edge.

Phase 1 (color palette + premium heat map + row density) can be completed in 1-2 days and will deliver the most dramatic visual improvement. Phase 2 (animations + progress bars + grid layouts) adds polish. Phase 3 (sparklines + notifications) establishes the "wow factor."

Start with Phase 1, item #1: `ThemeColors::dark()` in `app.rs`, line 92.
