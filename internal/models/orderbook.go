package models

import "time"

type OrderbookEntry struct {
	Price    float64
	Quantity float64
}

type OrderbookInfo struct {
	BtSpread       float64
	UpSpread       float64
	BtBuySlippage  float64
	UpSellSlippage float64
	RealGapBtUp    float64
	RealGapUpBt    float64
	SurfaceGap     float64
	RealKimpUp     map[Exchange]float64
	RealKimpBt     map[Exchange]float64
	Timestamp      time.Time
}
