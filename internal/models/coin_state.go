package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type CoinState struct {
	Symbol              string
	UpbitPrice          *decimal.Decimal
	BithumbPrice        *decimal.Decimal
	BinancePrice        *decimal.Decimal
	BinanceKrw          *decimal.Decimal
	BybitPrice          *decimal.Decimal
	BybitKrw            *decimal.Decimal
	OkxPrice            *decimal.Decimal
	OkxKrw              *decimal.Decimal
	UpbitKimchi         *decimal.Decimal
	BithumbKimchi       *decimal.Decimal
	BybitKimchiUp       *decimal.Decimal
	BybitKimchiBt       *decimal.Decimal
	OkxKimchiUp         *decimal.Decimal
	OkxKimchiBt         *decimal.Decimal
	DomesticGap         *decimal.Decimal
	BinanceFuturesPrice *decimal.Decimal
	BybitFuturesPrice   *decimal.Decimal
	FuturesBasis        *decimal.Decimal
	Timestamp           time.Time
}
