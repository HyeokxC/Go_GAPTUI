# Crypto Arbitrage Strategy — Complete Guide

> Reference document for buy/sell bot implementation.
> Covers all scenarios for Kimchi Premium (KP) gap trading between Korean and international exchanges.
>
> **Monitor Coverage**: The current Gap Monitor detects scenarios via `ScenarioDetector`:
> - **GapThreshold** → Maps to strategies 1-A, 1-B, 2-A (high premium triggers)
> - **DomesticGap** → Maps to strategy 1-C (UP-BT price difference)
> - **FutBasis** → Maps to strategies 3-A, 3-B, 3-C, 3-D (spot-futures basis for convergence trades)
> - **FutSpread** *(planned)* → Maps to strategies 4-A, 4-B (inter-exchange futures spread)
> - Orderbook slippage data supports pre-flight slippage/liquidity checks (displayed in UI)

---

## Terminology: Spot-Futures Basis

### Definitions

```
Contango (현선):  Spot < Futures  →  Spot is cheap, Futures is expensive
Backwardation (역현선):  Spot > Futures  →  Spot is expensive, Futures is cheap
```

### Action Matrix

```
Principle: Always BUY the cheap side, SELL the expensive side.

Contango (현선)    Spot < Futures  →  Buy Spot  + Short Futures
Backwardation (역현선)  Spot > Futures  →  Sell Spot + Long Futures
```

### Combined with Kimchi Premium

```
                  │ Positive KP            │ Reverse KP
                  │ (Domestic > Intl)      │ (Domestic < Intl)
──────────────────┼────────────────────────┼────────────────────────
Contango          │ Sell domestic spot     │ Buy domestic spot
(Spot < Futures)  │ + Short intl futures   │ + Short intl futures
                  │ (DOUBLE FAVORABLE)     │
──────────────────┼────────────────────────┼────────────────────────
Backwardation     │ Sell domestic spot     │ Buy domestic spot
(Spot > Futures)  │ + Long intl futures    │ + Long intl futures
                  │                        │ (DOUBLE FAVORABLE)
```

**DOUBLE FAVORABLE** = KP direction + basis direction align on the same side.
These are the highest-conviction entry points.

### When You Don't Hold Spot

Korean exchanges do not support spot short selling.
If you need to "sell spot" but have no holdings, your options change:

```
                  │ Hold Spot: YES              │ Hold Spot: NO
──────────────────┼─────────────────────────────┼─────────────────────────────
Contango          │ Sell spot + Short futures    │ Short futures only (naked)
(Spot < Futures)  │ → Market neutral, locked gap │ → Directional, but selling
                  │ → Hedge: Qty Match           │   the expensive side
                  │                              │ → Risk: futures can rise more
                  │                              │ → Cushion: funding (shorts
                  │                              │   may pay in contango)
──────────────────┼─────────────────────────────┼─────────────────────────────
Backwardation     │ Sell spot + Long futures     │ Long futures only (naked)
(Spot > Futures)  │ → Market neutral, locked gap │ → Directional, but buying
                  │ → Hedge: Qty Match           │   at 10% discount to spot
                  │                              │ → Risk: price drops further
                  │                              │ → Cushion: 10% discount
                  │                              │   + funding (longs receive
                  │                              │   in backwardation)
```

#### No-Spot Strategy Details

**Contango + No Spot → Naked Short Futures**

```
Entry:  Short futures at ₩1,100 (spot is ₩1,000)
Target: Futures converge down to ₩1,000 → +9.1% profit
Funding: In contango, shorts often pay funding (cost)
Risk:   Unhedged — if price pumps, short loses

Grade:
  BEST:  Gap converges + price drops    → gap profit + directional profit
  SOSO:  Gap converges + price rises    → gap profit offset by directional loss
  BAD:   Gap widens + price pumps       → double loss

Verdict: Less attractive — funding works against you
```

**Backwardation + No Spot → Naked Long Futures (RECOMMENDED)**

```
Entry:  Long futures at ₩900 (spot is ₩1,000)
Target: Futures converge up to ₩1,000 → +11.1% profit
Funding: In backwardation, longs receive funding (income)
Risk:   Unhedged — if price drops, long loses
Cushion: Bought at 10% discount + receiving funding

Grade:
  BEST:  Gap converges + price rises    → gap profit + directional + funding
  SOSO:  Gap converges + price drops    → gap profit offsets drop, funding helps
  BAD:   Gap widens + price dumps       → double loss (but 10% buffer absorbs)

Verdict: More attractive — funding works FOR you + significant discount cushion

Example (10% backwardation):
  Price drops  5%:  Still +5% vs spot buyers (discount absorbs)
  Price drops 10%:  Break-even (discount fully absorbed)
  Price drops 15%:  -5% loss (but spot buyers would be -15%)
  Price flat:       +10% gap convergence + funding income
```

#### Phased Approach: Build Into Hedged Position

```
Phase 1 (no spot): Enter naked long futures at discount
  → Accept directional risk with 10% cushion + funding

Phase 2 (acquire spot): When opportunity arises, buy spot
  → Convert to hedged position: sell spot + keep long futures
  → Now market-neutral with locked gap

Phase 3 (convergence): Gap closes
  → Close both sides → deterministic profit
```

---

## 1. Domestic Inter-Exchange Arbitrage

> Exploiting price differences between Korean exchanges (e.g., Bithumb ₩1,294 / Upbit ₩1,354)

---

### 1-A. Unhedged Transfer Arbitrage

```
Buy on Bithumb → Transfer to Upbit → Sell on Upbit
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Fast transfer + gap holds + sufficient liquidity | Full gap as profit (minus fees) |
| **SOSO** | Gap partially narrows during transfer | Reduced but positive profit |
| **BAD** | Gap reverses during transfer OR withdrawal suspended | Loss or funds stuck |

```
Example (gap 4.6%):

  BEST:  Transfer 5min, gap holds     → +4.6% - fees ≈ +4.0%
  SOSO:  Transfer 20min, gap → 2%     → +2.0% - fees ≈ +1.4%
  BAD:   Transfer 1hr, gap reverses   → -1.0% - fees ≈ -1.6%
```

**Hedge: None** — Fully exposed to price movement. Speed is everything.

---

### 1-B. Hedged Transfer Arbitrage (via International Futures)

```
Buy on Bithumb + Short on Binance Futures
→ Transfer to Upbit → Sell on Upbit + Close Binance short
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Fast transfer + Upbit price holds | Gap profit secured + minimal hedge cost |
| **SOSO** | Upbit-Bithumb gap narrows during transfer (overall price hedged) | Reduced but no major loss |
| **BAD** | Withdrawal suspended → forced to hold short → funding fees accumulate | Funding fees exceed profit |

```
Example (Bithumb-Upbit gap 4.6%, Bithumb-Binance gap 4.1%):

  BEST:  Fast transfer, Upbit price holds → +4.6% - fees ≈ +4.0%
  SOSO:  Upbit drops during transfer, Binance short covers price → +2~3%
  BAD:   Withdrawal blocked, hold short 5 days, funding 1.5% → +4.0% - 1.5% = +2.5%
         (Even BAD case is positive — that's the advantage of hedging)
```

**Hedge: Amount Matching**

```
Reason: The goal is to hedge "price movement" during transfer.
        Matching notional value ensures P&L offsets regardless of direction.

  Bithumb buy:   ₩10,000,000 (7,728 tokens × ₩1,294)
  Binance short: ₩10,000,000 (7,424 tokens × ₩1,347)

  Price movement → perfectly offset
  Gap profit = quantity difference (304 tokens) × closing price
```

---

### 1-C. Simultaneous Execution (No Transfer)

```
Buy on Bithumb + Sell on Upbit (simultaneously)
→ Requires pre-positioned assets on both exchanges
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Assets pre-positioned on both sides + large gap | Instant profit, near-zero risk |
| **SOSO** | Small gap, barely covers fees | Minimal profit |
| **BAD** | One side fills, other doesn't → directional exposure | Unhedged position |

```
Example (gap 4.6%):

  BEST:  Both fill simultaneously → +4.6% - fees ≈ +4.0% (near zero risk)
  SOSO:  Gap 1.5% → +1.5% - fees ≈ +0.9%
  BAD:   Upbit fills, Bithumb doesn't → left with short position only
```

**Hedge: Quantity Matching**

```
Reason: No transfer involved — "inventory management" is the priority.
        Same quantity on both sides ensures balanced positions.

  Bithumb buy:  7,728 tokens
  Upbit sell:   7,728 tokens (identical)

  → KRW difference = profit
  → Rebalance token inventory later via offline transfer
```

---

---

## 2. Domestic-International Exchange Arbitrage

> Physically moving coins between Korean and international exchanges to realize price differences.

---

### 2-A. International → Domestic (Capturing Kimchi Premium)

```
Buy on international exchange → Transfer to Korea → Sell on Korean exchange
(When Korean price is higher)
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | KP 5%+ sustained + fast transfer + USDT held without premium | Full KP as profit |
| **SOSO** | KP narrows during transfer | Reduced profit |
| **BAD** | USDT premium > coin premium (buying USDT at premium) | Net loss |

```
Key Trap: "How you acquire USDT"

  USDT premium 2%, coin KP 5% → effective profit 3%
  USDT premium 3%, coin KP 2% → effective profit -1%

  Already hold USD/USDT offshore → BEST
  Buying USDT in Korea to send → lose the USDT premium
```

**Hedge: Amount Matching (International Futures Short)**

```
Buy international spot + Short international futures (price hedge)
→ After arrival in Korea, sell spot + close futures

  International spot: $1,000 worth bought
  International futures: $1,000 worth shorted
  → Price movement offset during transfer
```

---

### 2-B. Domestic → International (Capturing Reverse Premium)

```
Buy on Korean exchange → Transfer to international → Sell on international exchange
(When Korean price is lower)
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Reverse premium 5%+ sustained + fast transfer + can utilize USDT offshore | Full reverse premium as profit |
| **SOSO** | Reverse premium narrows during transfer | Reduced profit |
| **BAD** | Difficult to repatriate KRW after selling offshore (forex regulations) | Funds stuck offshore |

```
Key Issue: Profit Repatriation

  Proceeds received as USDT offshore
  → Sending USDT back to Korea? Lose the USDT premium
  → Only makes sense if you plan to keep operating offshore
```

**Hedge: Amount Matching (International Futures Long)**

```
Buy domestic spot + Long international futures → price hedge during transfer
After arrival offshore → sell spot + close futures
```

---

### 2-C. Stablecoin Arbitrage (Direct USDT Premium)

```
When USDT has reverse premium:
Buy USDT in Korea (cheap) → Transfer offshore → Use offshore

When USDT has premium:
Acquire USDT offshore → Transfer to Korea → Sell in Korea (expensive)
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | USDT premium 3%+ + already hold USDT offshore | 3% - fees |
| **SOSO** | USDT premium 1.5% | Sub-1% after fees |
| **BAD** | USDT premium disappears during transfer | Break-even or loss |

**Hedge: Not Required** — Stablecoin, no price volatility ($1 peg)

---

---

## 3. Gap Trading (Pure Premium Play) — No Transfer

> No coin movement. Use domestic spot + international futures to bet on KP/reverse premium convergence.

---

### 3-A. Kimchi Premium Convergence (Korean Price is Higher)

```
Sell domestic spot + Long international futures
→ Close both when KP narrows
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | KP 5% → 0% convergence + low funding rate | +5% - fees - funding |
| **SOSO** | KP 5% → 2% partial convergence | +3% - costs |
| **BAD** | KP 5% → 10% widening + funding rate spikes | -5% - funding (double loss) |

```
Prerequisite: Must already hold the coin to sell on Korean exchange.
              Cannot enter this strategy without existing holdings.
```

**Hedge Matching: Quantity Matching**

```
Reason: No transfer. "Gap convergence" itself is the profit source.
        Matching quantity locks in deterministic profit at convergence.

  Domestic sell: 7,728 tokens × ₩1,354
  International long: 7,728 tokens × ₩1,294 equivalent

  Profit at convergence = 7,728 × (₩1,354 - ₩1,294) = ₩463,680
  → Same regardless of where the price ends up
```

---

### 3-B. Reverse Premium Convergence (Korean Price is Lower)

```
Buy domestic spot + Short international futures
→ Close both when reverse premium narrows (reverts to positive KP)
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Reverse -4% → KP +2% reversion + receive funding | +6% + funding income |
| **SOSO** | Reverse -4% → 0% convergence | +4% - costs |
| **BAD** | Reverse -4% → -10% widening + short liquidation risk | -6% + additional liquidation loss |

```
Why reverse premium is more favorable than positive KP:
  1. Only need KRW to enter (no existing coin holdings required)
  2. Historical reversion to positive KP is highly probable
  3. Short positions may receive funding rate payments
```

**Hedge Matching: Quantity Matching**

```
Reason: Same as 3-A. Gap convergence is the profit source.

  Domestic buy: 7,728 tokens × ₩1,294
  International short: 7,728 tokens × ₩1,347

  Profit at convergence = 7,728 × (₩1,347 - ₩1,294) = ₩409,584
  → Price-independent, deterministic
```

---

### 3-C. Gap + Funding Rate Double Play (Advanced)

```
During reverse premium period:
  Buy domestic spot + Short international futures

  Profit source 1: Reverse premium → KP convergence (gap profit)
  Profit source 2: Short position funding rate income (during market overheating)
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Reverse premium converges + positive funding (shorts receive) | Gap profit + funding income |
| **SOSO** | Reverse premium converges but pay funding | Gap profit - funding |
| **BAD** | Reverse premium widens + pay funding | Double loss |

---

### 3-D. Reverse Cash-and-Carry via Lending (대출 기반 역캐시앤캐리)

```
Borrow spot (CEX margin or DeFi lending) → Sell spot + Long futures
→ Close when backwardation converges: Buy spot (repay loan) + Close futures long
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Deep backwardation (5%+) + low borrow rate (<20% APR) + receive funding | Basis convergence + funding income − low interest = max profit |
| **SOSO** | Partial convergence, moderate borrow rate | Convergence profit partially offset by interest |
| **BAD** | Backwardation deepens + borrow rate spikes + liquidation | Interest bleeds + basis loss + potential forced close |

```
Key difference from 3-A (sell existing spot):
  3-A requires existing coin holdings → opportunity cost (implicit)
  3-D borrows coin → interest cost (explicit)
  3-D enables entry WITHOUT existing holdings

Profit equation:
  P = (Spot − Futures) × Qty                    ... basis convergence
    + Funding rate income (longs receive in backwardation)
    − Borrow interest (variable, can spike)
    − Trading fees (spot sell + futures long + close both)

Entry condition:
  Basis convergence profit + Funding income > Borrow interest + Fees
```

**Venue Options:**

```
CEX Margin (Binance Cross Margin):
  ✅ Borrow + sell + futures long on same platform
  ✅ Lower operational complexity
  ❌ Variable interest rate (exchange-controlled)
  ❌ Separate margin vs futures accounts → no auto-netting on liquidation

DeFi Lending (Aave, Compound):
  ✅ Transparent, market-driven rates
  ✅ No KYC, permissionless
  ❌ Overcollateralized (150%+) → low capital efficiency
  ❌ Cross-platform risk (DeFi borrow + CEX futures)
  ❌ Gas fees, smart contract risk
```

**Hedge Matching: Quantity Matching**

```
Reason: Same as 3-A/3-B. Gap convergence is the profit source.

  Borrow & sell: 7,728 tokens × ₩1,354 (spot)
  Futures long:  7,728 tokens × ₩1,294 (futures)

  At convergence: Buy back spot at converged price, repay loan
  Profit = 7,728 × (₩1,354 − ₩1,294) − interest − fees
```

**Risk Management:**

```
Must monitor:
  1. Borrow utilization rate → interest rate proxy
  2. Collateral ratio → liquidation distance
  3. Funding rate direction → income or cost
  4. Basis trend → convergence or divergence

Stop-loss triggers:
  − Borrow rate exceeds annualized basis convergence rate
  − Collateral ratio < 130% (if DeFi)
  − Basis widens beyond 2× entry basis
  − Cumulative interest exceeds 50% of expected profit
```

---

---

## 4. Inter-exchange Futures Spread (거래소간 선물 스프레드)

> No coin movement. Long futures on one exchange + Short futures on another.
> Captures price discrepancies or funding rate differentials between venues.
>
> **Traditional finance term: Inter-market Spread (거래소간 스프레드)**
> Also known as: Exchange Spread, Cross-venue Arbitrage, Location Spread

---

### Why Spreads Exist Between Exchanges

```
Same underlying (e.g., BTC), different exchange → different price.

Causes:
  1. Liquidity depth differences (Binance >> Bybit)
  2. User base composition (retail vs institutional)
  3. Funding rate calculation differences (each exchange independent)
  4. Margin/leverage rules (affect positioning capacity)
  5. Regional demand pressure (geographic user base)

Key insight: In TradFi, inter-market spreads exist for similar reasons
  (e.g., Brent vs WTI oil, COMEX vs TOCOM gold, CME vs Eurex futures)
```

---

### 4-A. Perpetual Futures Spread (Funding Rate Arbitrage)

```
Short on high-funding exchange + Long on low-funding exchange
→ Capture funding rate differential continuously
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Sustained funding differential (>0.03% per 8h) + prices stay aligned | Continuous funding income, near-zero directional risk |
| **SOSO** | Funding differential narrows over time | Reduced but positive income |
| **BAD** | Price divergence between exchanges + funding flips | Unrealized loss on spread + funding reversal |

```
Example (BTC perpetual):
  Binance funding: +0.10% / 8h (longs pay shorts)
  Bybit funding:   +0.02% / 8h (longs pay shorts)

  Action: Short Binance (receive 0.10%) + Long Bybit (pay 0.02%)
  Net income: 0.08% per 8h = 0.24% per day ≈ 87% APR

  Risk: If Binance price drops faster than Bybit
        → temporary unrealized loss (prices re-converge)
        → manageable if position size is conservative

  Capital: Margin required on BOTH exchanges
           → capital efficiency ≈ 50% (funds split across venues)
```

**Hedge Matching: Quantity Matching**

```
Reason: Pure spread trade — profit comes from funding differential,
        not price movement. Identical quantity ensures delta neutrality.

  Binance short: 1.000 BTC
  Bybit long:    1.000 BTC

  Net exposure: 0 BTC (delta neutral)
  Income: funding rate differential × notional × time
```

---

### 4-B. Dated Futures Spread (Cross-exchange Basis)

```
When same-expiry futures trade at different prices across exchanges:
Long cheap exchange + Short expensive exchange
→ Convergence at expiry guarantees profit (if same settlement index)
```

| Grade | Condition | Result |
|-------|-----------|--------|
| **BEST** | Large price gap + both converge to same settlement price | Deterministic profit at settlement |
| **SOSO** | Small gap, barely covers fees | Minimal profit |
| **BAD** | Different settlement indices or exchange-specific risk event | Settlement price divergence |

```
Example (BTC quarterly futures, same expiry):
  Binance: $101,000
  Bybit:   $100,500

  Action: Long Bybit ($100,500) + Short Binance ($101,000)
  At expiry: Both settle to ~same index price
  Profit: $500 per BTC − fees

  This is the purest form of inter-market spread arbitrage.
  Risk is near-zero IF settlement indices are identical.
```

---

### Traditional Finance Terminology Reference

```
┌───────────────────────────────┬────────────────────────────────────┐
│ TradFi Term                   │ Crypto Application                 │
├───────────────────────────────┼────────────────────────────────────┤
│ Inter-market Spread           │ 거래소간 선물 가격차 거래           │
│ (거래소간 스프레드)            │ e.g., Binance vs Bybit futures     │
├───────────────────────────────┼────────────────────────────────────┤
│ Calendar Spread               │ 같은 거래소, 다른 만기 선물         │
│ (캘린더 스프레드)              │ e.g., BTC 0328 vs BTC 0627         │
├───────────────────────────────┼────────────────────────────────────┤
│ Cash-and-Carry                │ 현물 매수 + 선물 매도 (컨탱고)      │
│ (캐시앤캐리)                  │                                    │
├───────────────────────────────┼────────────────────────────────────┤
│ Reverse Cash-and-Carry        │ 현물 매도(대차) + 선물 매수          │
│ (역캐시앤캐리)                │ (백워데이션)                        │
├───────────────────────────────┼────────────────────────────────────┤
│ Convergence Trade             │ 괴리 수렴에 베팅하는 모든 전략       │
│ (수렴 거래)                   │                                    │
├───────────────────────────────┼────────────────────────────────────┤
│ Statistical Arbitrage         │ 통계적 평균회귀 기반 차익거래        │
│ (통계적 차익거래)              │                                    │
└───────────────────────────────┴────────────────────────────────────┘

Crypto-specific terms (no direct TradFi equivalent):
  − Funding Rate Arbitrage: 무기한 선물 펀딩비 차익거래
  − DEX-CEX Arbitrage: 탈중앙-중앙 거래소 간 차익거래
  − Kimchi Premium Trade: 김치 프리미엄 차익거래
```

---

---

## Hedge Method Summary

### When to Use Quantity vs Amount Matching

```
Decision rule: "Is there a transfer?"

  Transfer involved (arbitrage)  → Amount Matching
  No transfer (gap trading)      → Quantity Matching
```

| Attribute | Amount Matching | Quantity Matching |
|-----------|----------------|-------------------|
| **Purpose** | Hedge price movement during transfer | Lock deterministic profit at convergence |
| **Used in** | 1-B, 2-A, 2-B (transfer arbitrage) | 1-C, 3-A, 3-B (gap trading) |
| **Advantage** | Perfect price movement offset | Profit is fixed (price-independent) |
| **Disadvantage** | Profit varies with closing price | Slight directional exposure while holding |
| **Calculation** | Match KRW notional value on both sides | Match token quantity on both sides |

```
Amount Matching Example:
  Bithumb:  7,728 tokens × ₩1,294 = ₩10,000,000
  Binance:  7,424 tokens × ₩1,347 = ₩10,000,000
  → Same notional, different quantity

Quantity Matching Example:
  Bithumb:  7,728 tokens × ₩1,294 = ₩10,000,032
  Binance:  7,728 tokens × ₩1,347 = ₩10,409,616
  → Same quantity, different notional
```

---

## Strategy Selection Flowchart

```
Can coins be transferred?
│
├── YES → Withdrawal/deposit enabled
│   │
│   ├── Domestic↔Domestic gap?  → 1-B. Domestic arb + intl hedge (Amount Match)
│   ├── Domestic↔International gap? → 2-A/2-B. International arb (Amount Match)
│   └── USDT gap only?          → 2-C. Stablecoin arb (No hedge needed)
│
├── NO → Withdrawal blocked or choosing not to transfer
│   │
│   ├── Already holding the coin?
│   │   ├── YES + positive KP → 3-A. Sell domestic + Long intl (Qty Match)
│   │   └── NO  + reverse KP  → 3-B. Buy domestic + Short intl (Qty Match)
│   │
│   ├── Funding rate favorable? → 3-C. Gap + funding double play (Qty Match)
│   │
│   └── Backwardation + no spot holdings?
│       └── Borrow rate < basis? → 3-D. Rev. Cash-and-Carry via lending (Qty Match)
│
├── Futures price differs across exchanges?
│   │
│   ├── Perpetual + funding rate gap?
│   │   └── Sustained differential → 4-A. Perp Futures Spread (Qty Match)
│   │
│   └── Dated futures + price gap?
│       └── Same expiry, different price → 4-B. Dated Futures Spread (Qty Match)
│
└── Assets pre-positioned on both exchanges?
    └── YES → 1-C. Simultaneous execution (Qty Match, safest)
```

---

## Strategy Comparison Matrix

| Strategy | Profit Source | Risk | Hedge | Difficulty | Capital Efficiency |
|----------|-------------|------|-------|------------|-------------------|
| 1-A Unhedged Transfer | Inter-exchange gap | Price movement during transfer | None | Low | High |
| 1-B Hedged Transfer | Inter-exchange gap | Withdrawal suspension | Amount | Medium | Medium |
| 1-C Simultaneous Exec | Inter-exchange gap | Fill delay | Quantity | Low | Low (both sides) |
| 2-A Intl→Domestic | Kimchi premium | USDT premium | Amount | High | Medium |
| 2-B Domestic→Intl | Reverse premium | KRW repatriation | Amount | High | Medium |
| 2-C Stablecoin | USDT premium | Transfer time | None | Low | High |
| 3-A KP Convergence | KP narrowing | KP widening | Quantity | Medium | Low (both sides) |
| 3-B Reverse Convergence | Reverse KP reversion | Reverse KP widening | Quantity | Medium | Low |
| 3-C Gap + Funding | Gap + funding rate | Double adverse move | Quantity | High | Low |
| 3-D Rev. Cash-and-Carry | Backwardation convergence | Interest spike + liquidation | Quantity | High | Low (+ collateral) |
| 4-A Perp Futures Spread | Funding rate differential | Price divergence + funding flip | Quantity | Medium | Low (split across venues) |
| 4-B Dated Futures Spread | Cross-exchange basis | Settlement index divergence | Quantity | Low | Low (split across venues) |

---

## Bot Implementation Notes

### Key Parameters for Automation

```
Entry Signals:
  - gap_threshold:     Minimum gap % to trigger (e.g., 3.0%)
  - funding_rate_max:  Maximum acceptable funding rate for holding
  - liquidity_min:     Minimum orderbook depth at target price
  - slippage_max:      Maximum acceptable slippage (e.g., 0.1%)
  - spread_max:        Maximum bid-ask spread (e.g., 0.5%)

Exit Signals:
  - target_gap:        Target gap % to close (e.g., 0.5%)
  - stop_loss_gap:     Gap widening stop-loss (e.g., gap doubles)
  - time_stop:         Max holding period (e.g., 14 days)
  - funding_stop:      Cumulative funding cost threshold

Pre-flight Checks:
  - withdrawal_enabled: Verify withdrawal is open (for transfer strategies)
  - transfer_time_est:  Estimated network transfer time
  - margin_available:   Sufficient margin for futures position
  - network_fee:        Current blockchain network fee

Strategy 3-D Specific (Reverse Cash-and-Carry via Lending):
  - borrow_rate_max:   Maximum acceptable borrow APR (e.g., 20%)
  - collateral_ratio:  Minimum collateral ratio to maintain (e.g., 150%)
  - basis_min:         Minimum backwardation depth to enter (e.g., 5%)
  - interest_stop:     Cumulative interest exceeds X% of expected profit (e.g., 50%)

Strategy 4 Specific (Inter-exchange Futures Spread):
  - funding_diff_min:  Minimum funding rate differential (e.g., 0.03% per 8h)
  - price_div_max:     Maximum acceptable price divergence between venues (e.g., 0.5%)
  - fut_spread_min:    Minimum dated futures price spread (e.g., 0.3%)
  - rebalance_trigger: Quantity drift threshold for rebalancing (e.g., 2%)
```
