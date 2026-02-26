package exchanges

import (
	"time"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

const defaultSlippageBudgetKRW = 10_000_000.0

func CalcBuyAvg(asks []models.OrderbookEntry, budgetKRW float64) (avgPrice float64, totalQty float64, ok bool) {
	if budgetKRW <= 0 {
		return 0, 0, false
	}

	remaining := budgetKRW
	totalCost := 0.0
	totalBought := 0.0

	for _, ask := range asks {
		if remaining <= 0 {
			break
		}
		if ask.Price <= 0 || ask.Quantity <= 0 {
			continue
		}

		levelCost := ask.Price * ask.Quantity
		if levelCost <= remaining {
			totalCost += levelCost
			totalBought += ask.Quantity
			remaining -= levelCost
			continue
		}

		partialQty := remaining / ask.Price
		if partialQty <= 0 {
			break
		}
		totalCost += partialQty * ask.Price
		totalBought += partialQty
		remaining = 0
	}

	if totalBought <= 0 {
		return 0, 0, false
	}
	return totalCost / totalBought, totalBought, true
}

func CalcSellAvg(bids []models.OrderbookEntry, sellQty float64) (avgPrice float64, ok bool) {
	if sellQty <= 0 {
		return 0, false
	}

	remaining := sellQty
	totalRevenue := 0.0
	totalSold := 0.0

	for _, bid := range bids {
		if remaining <= 0 {
			break
		}
		if bid.Price <= 0 || bid.Quantity <= 0 {
			continue
		}

		if bid.Quantity <= remaining {
			totalRevenue += bid.Price * bid.Quantity
			totalSold += bid.Quantity
			remaining -= bid.Quantity
			continue
		}

		partialQty := remaining
		totalRevenue += bid.Price * partialQty
		totalSold += partialQty
		remaining = 0
	}

	if totalSold <= 0 {
		return 0, false
	}
	return totalRevenue / totalSold, true
}

func CalcSpread(bestBid float64, bestAsk float64) float64 {
	if bestAsk <= 0 {
		return 0
	}
	return (bestAsk - bestBid) / bestAsk * 100.0
}

func CalcRealKimp(
	domesticBids []models.OrderbookEntry,
	overseasAsks []models.OrderbookEntry,
	budgetKRW float64,
	usdtKRW float64,
) (float64, bool) {
	if budgetKRW <= 0 || usdtKRW <= 0 {
		return 0, false
	}

	budgetUSDT := budgetKRW / usdtKRW
	overseasBuyAvg, qty, ok := CalcBuyAvg(overseasAsks, budgetUSDT)
	if !ok || qty <= 0 || overseasBuyAvg <= 0 {
		return 0, false
	}

	overseasKRW := overseasBuyAvg * usdtKRW
	if overseasKRW <= 0 {
		return 0, false
	}

	domesticSellAvg, ok := CalcSellAvg(domesticBids, qty)
	if !ok {
		return 0, false
	}

	realKimp := (domesticSellAvg - overseasKRW) / overseasKRW * 100.0
	return realKimp, true
}

func ProcessOrderbookInfo(
	symbol string,
	manager *OrderbookManager,
	usdtKRW float64,
	budgetKRW float64,
) *models.OrderbookInfo {
	if budgetKRW <= 0 {
		budgetKRW = defaultSlippageBudgetKRW
	}

	info := &models.OrderbookInfo{
		RealKimpUp: map[models.Exchange]float64{
			models.Binance: 0,
			models.Bybit:   0,
			models.Okx:     0,
		},
		RealKimpBt: map[models.Exchange]float64{
			models.Binance: 0,
			models.Bybit:   0,
			models.Okx:     0,
		},
		Timestamp: time.Now(),
	}
	if manager == nil {
		return info
	}

	bt := manager.Get(models.Bithumb, symbol)
	up := manager.Get(models.Upbit, symbol)
	bn := manager.Get(models.Binance, symbol)
	bb := manager.Get(models.Bybit, symbol)
	okxBook := manager.Get(models.Okx, symbol)

	if bt != nil && len(bt.Bids) > 0 && len(bt.Asks) > 0 {
		info.BtSpread = CalcSpread(bt.Bids[0].Price, bt.Asks[0].Price)
		if btBuyAvg, _, ok := CalcBuyAvg(bt.Asks, budgetKRW); ok && bt.Asks[0].Price > 0 {
			info.BtBuySlippage = (btBuyAvg - bt.Asks[0].Price) / bt.Asks[0].Price * 100.0
		}
	}

	if up != nil && len(up.Bids) > 0 && len(up.Asks) > 0 {
		info.UpSpread = CalcSpread(up.Bids[0].Price, up.Asks[0].Price)
		if _, upQty, ok := CalcBuyAvg(up.Asks, budgetKRW); ok && up.Bids[0].Price > 0 {
			if upSellAvg, sellOK := CalcSellAvg(up.Bids, upQty); sellOK {
				info.UpSellSlippage = (up.Bids[0].Price - upSellAvg) / up.Bids[0].Price * 100.0
			}
		}
	}

	if bt != nil && up != nil {
		if btBuyAvg, qty, ok := CalcBuyAvg(bt.Asks, budgetKRW); ok && btBuyAvg > 0 {
			if upSellAvg, sellOK := CalcSellAvg(up.Bids, qty); sellOK {
				info.RealGapBtUp = (upSellAvg - btBuyAvg) / btBuyAvg * 100.0
			}
		}
		if upBuyAvg, qty, ok := CalcBuyAvg(up.Asks, budgetKRW); ok && upBuyAvg > 0 {
			if btSellAvg, sellOK := CalcSellAvg(bt.Bids, qty); sellOK {
				info.RealGapUpBt = (btSellAvg - upBuyAvg) / upBuyAvg * 100.0
			}
		}
	}

	if up != nil && bn != nil && len(up.Bids) > 0 && len(bn.Asks) > 0 && usdtKRW > 0 {
		overseasKRW := bn.Asks[0].Price * usdtKRW
		if overseasKRW > 0 {
			info.SurfaceGap = (up.Bids[0].Price - overseasKRW) / overseasKRW * 100.0
		}
	}

	if up != nil {
		if bn != nil {
			if value, ok := CalcRealKimp(up.Bids, bn.Asks, budgetKRW, usdtKRW); ok {
				info.RealKimpUp[models.Binance] = value
			}
		}
		if bb != nil {
			if value, ok := CalcRealKimp(up.Bids, bb.Asks, budgetKRW, usdtKRW); ok {
				info.RealKimpUp[models.Bybit] = value
			}
		}
		if okxBook != nil {
			if value, valid := CalcRealKimp(up.Bids, okxBook.Asks, budgetKRW, usdtKRW); valid {
				info.RealKimpUp[models.Okx] = value
			}
		}
	}

	if bt != nil {
		if bn != nil {
			if value, ok := CalcRealKimp(bt.Bids, bn.Asks, budgetKRW, usdtKRW); ok {
				info.RealKimpBt[models.Binance] = value
			}
		}
		if bb != nil {
			if value, ok := CalcRealKimp(bt.Bids, bb.Asks, budgetKRW, usdtKRW); ok {
				info.RealKimpBt[models.Bybit] = value
			}
		}
		if okxBook != nil {
			if value, valid := CalcRealKimp(bt.Bids, okxBook.Asks, budgetKRW, usdtKRW); valid {
				info.RealKimpBt[models.Okx] = value
			}
		}
	}

	return info
}
