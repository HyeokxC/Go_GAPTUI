package monitor

import (
	"sync"
	"time"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
	"github.com/shopspring/decimal"
)

type KimchiMonitor struct {
	mu      sync.RWMutex
	states  map[string]*models.CoinState
	usdtKRW float64
}

func NewKimchiMonitor() *KimchiMonitor {
	return &KimchiMonitor{
		states: make(map[string]*models.CoinState),
	}
}

func (m *KimchiMonitor) OnTicker(ticker models.TickerData) *models.CoinState {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.states[ticker.Symbol]
	if !ok {
		state = &models.CoinState{Symbol: ticker.Symbol}
		m.states[ticker.Symbol] = state
	}

	price := ticker.Price
	switch ticker.Exchange {
	case models.Upbit:
		state.UpbitPrice = cloneDecimal(price)
	case models.Bithumb:
		state.BithumbPrice = cloneDecimal(price)
	case models.Binance:
		state.BinancePrice = cloneDecimal(price)
	case models.Bybit:
		state.BybitPrice = cloneDecimal(price)
	case models.Okx:
		state.OkxPrice = cloneDecimal(price)
	case models.BinanceFutures:
		state.BinanceFuturesPrice = cloneDecimal(price)
	case models.BybitFutures:
		state.BybitFuturesPrice = cloneDecimal(price)
	}

	usdtKRW := decimal.NewFromFloat(m.usdtKRW)
	if state.BinancePrice != nil {
		v := state.BinancePrice.Mul(usdtKRW)
		state.BinanceKrw = &v
	}
	if state.BybitPrice != nil {
		v := state.BybitPrice.Mul(usdtKRW)
		state.BybitKrw = &v
	}
	if state.OkxPrice != nil {
		v := state.OkxPrice.Mul(usdtKRW)
		state.OkxKrw = &v
	}

	recalcPremiums(state, usdtKRW)

	if !ticker.Timestamp.IsZero() {
		state.Timestamp = ticker.Timestamp
	} else {
		state.Timestamp = time.Now()
	}

	return cloneState(state)
}

func (m *KimchiMonitor) SetUsdtKRW(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.usdtKRW = rate
	usdtKRW := decimal.NewFromFloat(rate)
	for _, state := range m.states {
		if state.BinancePrice != nil {
			v := state.BinancePrice.Mul(usdtKRW)
			state.BinanceKrw = &v
		}
		if state.BybitPrice != nil {
			v := state.BybitPrice.Mul(usdtKRW)
			state.BybitKrw = &v
		}
		if state.OkxPrice != nil {
			v := state.OkxPrice.Mul(usdtKRW)
			state.OkxKrw = &v
		}
		recalcPremiums(state, usdtKRW)
	}
}

func (m *KimchiMonitor) GetStates() map[string]*models.CoinState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cloned := make(map[string]*models.CoinState, len(m.states))
	for symbol, state := range m.states {
		cloned[symbol] = cloneState(state)
	}

	return cloned
}

func recalcPremiums(state *models.CoinState, usdtKRW decimal.Decimal) {
	hundred := decimal.NewFromInt(100)

	state.UpbitKimchi = nil
	state.BithumbKimchi = nil
	state.BybitKimchiUp = nil
	state.BybitKimchiBt = nil
	state.OkxKimchiUp = nil
	state.OkxKimchiBt = nil
	state.DomesticGap = nil
	state.FuturesBasis = nil

	if state.UpbitPrice != nil && state.BithumbPrice != nil && !state.BithumbPrice.IsZero() {
		v := state.UpbitPrice.Sub(*state.BithumbPrice).Div(*state.BithumbPrice).Mul(hundred)
		state.DomesticGap = &v
	}

	if state.BinancePrice != nil {
		binanceKRW := state.BinancePrice.Mul(usdtKRW)
		if !binanceKRW.IsZero() {
			if state.UpbitPrice != nil {
				v := state.UpbitPrice.Sub(binanceKRW).Div(binanceKRW).Mul(hundred)
				state.UpbitKimchi = &v
			}
			if state.BithumbPrice != nil {
				v := state.BithumbPrice.Sub(binanceKRW).Div(binanceKRW).Mul(hundred)
				state.BithumbKimchi = &v
			}
		}
	}

	if state.BybitPrice != nil {
		bybitKRW := state.BybitPrice.Mul(usdtKRW)
		if !bybitKRW.IsZero() {
			if state.UpbitPrice != nil {
				v := state.UpbitPrice.Sub(bybitKRW).Div(bybitKRW).Mul(hundred)
				state.BybitKimchiUp = &v
			}
			if state.BithumbPrice != nil {
				v := state.BithumbPrice.Sub(bybitKRW).Div(bybitKRW).Mul(hundred)
				state.BybitKimchiBt = &v
			}
		}
	}

	if state.OkxPrice != nil {
		okxKRW := state.OkxPrice.Mul(usdtKRW)
		if !okxKRW.IsZero() {
			if state.UpbitPrice != nil {
				v := state.UpbitPrice.Sub(okxKRW).Div(okxKRW).Mul(hundred)
				state.OkxKimchiUp = &v
			}
			if state.BithumbPrice != nil {
				v := state.BithumbPrice.Sub(okxKRW).Div(okxKRW).Mul(hundred)
				state.OkxKimchiBt = &v
			}
		}
	}

	if state.BinancePrice != nil {
		if state.BinanceFuturesPrice != nil && !state.BinanceFuturesPrice.IsZero() {
			v := state.BinancePrice.Sub(*state.BinanceFuturesPrice).Div(*state.BinanceFuturesPrice).Mul(hundred)
			state.FuturesBasis = &v
		} else if state.BybitFuturesPrice != nil && !state.BybitFuturesPrice.IsZero() {
			v := state.BinancePrice.Sub(*state.BybitFuturesPrice).Div(*state.BybitFuturesPrice).Mul(hundred)
			state.FuturesBasis = &v
		}
	}
}

func cloneState(state *models.CoinState) *models.CoinState {
	if state == nil {
		return nil
	}

	return &models.CoinState{
		Symbol:              state.Symbol,
		UpbitPrice:          cloneDecimalPtr(state.UpbitPrice),
		BithumbPrice:        cloneDecimalPtr(state.BithumbPrice),
		BinancePrice:        cloneDecimalPtr(state.BinancePrice),
		BinanceKrw:          cloneDecimalPtr(state.BinanceKrw),
		BybitPrice:          cloneDecimalPtr(state.BybitPrice),
		BybitKrw:            cloneDecimalPtr(state.BybitKrw),
		OkxPrice:            cloneDecimalPtr(state.OkxPrice),
		OkxKrw:              cloneDecimalPtr(state.OkxKrw),
		UpbitKimchi:         cloneDecimalPtr(state.UpbitKimchi),
		BithumbKimchi:       cloneDecimalPtr(state.BithumbKimchi),
		BybitKimchiUp:       cloneDecimalPtr(state.BybitKimchiUp),
		BybitKimchiBt:       cloneDecimalPtr(state.BybitKimchiBt),
		OkxKimchiUp:         cloneDecimalPtr(state.OkxKimchiUp),
		OkxKimchiBt:         cloneDecimalPtr(state.OkxKimchiBt),
		DomesticGap:         cloneDecimalPtr(state.DomesticGap),
		BinanceFuturesPrice: cloneDecimalPtr(state.BinanceFuturesPrice),
		BybitFuturesPrice:   cloneDecimalPtr(state.BybitFuturesPrice),
		FuturesBasis:        cloneDecimalPtr(state.FuturesBasis),
		Timestamp:           state.Timestamp,
	}
}

func cloneDecimal(d decimal.Decimal) *decimal.Decimal {
	v := d
	return &v
}

func cloneDecimalPtr(d *decimal.Decimal) *decimal.Decimal {
	if d == nil {
		return nil
	}
	v := *d
	return &v
}
