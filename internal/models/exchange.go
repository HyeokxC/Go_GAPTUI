package models

type Exchange int

const (
	Bithumb Exchange = iota
	Upbit
	Binance
	Bybit
	Bitget
	Okx
	Gate
	BinanceFutures
	BybitFutures
)

func (e Exchange) String() string {
	switch e {
	case Bithumb:
		return "Bithumb"
	case Upbit:
		return "Upbit"
	case Binance:
		return "Binance"
	case Bybit:
		return "Bybit"
	case Bitget:
		return "Bitget"
	case Okx:
		return "OKX"
	case Gate:
		return "Gate"
	case BinanceFutures:
		return "Binance Futures"
	case BybitFutures:
		return "Bybit Futures"
	default:
		return "Unknown"
	}
}

func (e Exchange) ShortCode() string {
	switch e {
	case Bithumb:
		return "BT"
	case Upbit:
		return "UP"
	case Binance:
		return "BN"
	case Bybit:
		return "BB"
	case Bitget:
		return "BG"
	case Okx:
		return "OK"
	case Gate:
		return "GT"
	case BinanceFutures:
		return "BNF"
	case BybitFutures:
		return "BBF"
	default:
		return ""
	}
}

func (e Exchange) IsDomestic() bool {
	return e == Upbit || e == Bithumb
}

func (e Exchange) IsOverseas() bool {
	return e == Binance || e == Bybit || e == Bitget || e == Okx || e == Gate
}

func (e Exchange) IsFutures() bool {
	return e == BinanceFutures || e == BybitFutures
}

func ExchangeFromShortCode(code string) (Exchange, bool) {
	switch code {
	case "BT":
		return Bithumb, true
	case "UP":
		return Upbit, true
	case "BN":
		return Binance, true
	case "BB":
		return Bybit, true
	case "BG":
		return Bitget, true
	case "OK":
		return Okx, true
	case "GT":
		return Gate, true
	case "BNF":
		return BinanceFutures, true
	case "BBF":
		return BybitFutures, true
	default:
		return 0, false
	}
}

func SpotExchanges() []Exchange {
	return []Exchange{Upbit, Bithumb, Binance, Bybit, Bitget, Okx, Gate}
}

func DomesticExchanges() []Exchange {
	return []Exchange{Upbit, Bithumb}
}

func OverseasExchanges() []Exchange {
	return []Exchange{Binance, Bybit, Bitget, Okx, Gate}
}

func AllExchanges() []Exchange {
	return []Exchange{Bithumb, Upbit, Binance, Bybit, Bitget, Okx, Gate, BinanceFutures, BybitFutures}
}
