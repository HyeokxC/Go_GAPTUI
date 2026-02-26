package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TickerData struct {
	Exchange  Exchange
	Symbol    string
	Price     decimal.Decimal
	Timestamp time.Time
}
