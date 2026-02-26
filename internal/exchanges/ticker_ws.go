package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hyeokx/Go_GAPTUI/internal/models"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type TickerCallback func(data models.TickerData)

func RunUpbitTicker(ctx context.Context, symbols []string, callback TickerCallback) {
	go func() {
		const wsURL = "wss://api.upbit.com/websocket/v1"

		type upbitTickerMessage struct {
			Code       string  `json:"code"`
			TradePrice float64 `json:"trade_price"`
		}

		codes := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			base := normalizeBaseSymbol(symbol)
			if base == "" {
				continue
			}
			codes = append(codes, "KRW-"+base)
		}

		subscribe := []map[string]json.RawMessage{
			{"ticket": json.RawMessage(`"trade-stream"`)},
			{"type": json.RawMessage(`"trade"`), "codes": mustJSONRaw(codes)},
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := dialWebSocket(wsURL)
			if err != nil {
				log.Error().Err(err).Str("exchange", "Upbit").Msg("ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", "Upbit").Msg("ws subscribe failed")
				_ = conn.Close()
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			readFailed := false
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					readFailed = true
					break
				}

				var payload upbitTickerMessage
				if err := json.Unmarshal(message, &payload); err != nil {
					continue
				}

				parts := strings.Split(payload.Code, "-")
				if len(parts) < 2 {
					continue
				}

				callback(models.TickerData{
					Exchange:  models.Upbit,
					Symbol:    strings.ToUpper(parts[len(parts)-1]),
					Price:     decimal.NewFromFloat(payload.TradePrice),
					Timestamp: time.Now(),
				})
			}

			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "Upbit").Msg("ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunBithumbTicker(ctx context.Context, symbols []string, callback TickerCallback) {
	go func() {
		const wsURL = "wss://pubwss.bithumb.com/pub/ws"

		type bithumbItem struct {
			Symbol    string `json:"symbol"`
			ContPrice string `json:"contPrice"`
		}

		type bithumbMessage struct {
			Type    string `json:"type"`
			Content struct {
				List []bithumbItem `json:"list"`
			} `json:"content"`
		}

		pairs := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			base := normalizeBaseSymbol(symbol)
			if base == "" {
				continue
			}
			pairs = append(pairs, base+"_KRW")
		}

		subscribe := struct {
			Type    string   `json:"type"`
			Symbols []string `json:"symbols"`
		}{
			Type:    "transaction",
			Symbols: pairs,
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := dialWebSocket(wsURL)
			if err != nil {
				log.Error().Err(err).Str("exchange", "Bithumb").Msg("ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", "Bithumb").Msg("ws subscribe failed")
				_ = conn.Close()
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			readFailed := false
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					readFailed = true
					break
				}

				var payload bithumbMessage
				if err := json.Unmarshal(message, &payload); err != nil {
					continue
				}
				if payload.Type != "transaction" {
					continue
				}

				for _, item := range payload.Content.List {
					parts := strings.Split(item.Symbol, "_")
					if len(parts) < 1 {
						continue
					}

					price, err := decimal.NewFromString(strings.ReplaceAll(item.ContPrice, ",", ""))
					if err != nil {
						continue
					}

					callback(models.TickerData{
						Exchange:  models.Bithumb,
						Symbol:    strings.ToUpper(parts[0]),
						Price:     price,
						Timestamp: time.Now(),
					})
				}
			}

			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "Bithumb").Msg("ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunBinanceTicker(ctx context.Context, usdtSymbol string, callback TickerCallback) {
	go runBinanceMiniTicker(ctx, "wss://stream.binance.com:9443/ws/!miniTicker@arr", usdtSymbol, models.Binance, callback)
}

func RunBybitTicker(ctx context.Context, symbols []string, callback TickerCallback) {
	go runBybitTicker(ctx, "wss://stream.bybit.com/v5/public/spot", symbols, models.Bybit, callback)
}

func RunOKXTicker(ctx context.Context, symbols []string, callback TickerCallback) {
	go func() {
		const wsURL = "wss://ws.okx.com:8443/ws/v5/public"

		type okxArg struct {
			Channel string `json:"channel"`
			InstID  string `json:"instId"`
		}

		type okxData struct {
			Last string `json:"last"`
		}

		type okxMessage struct {
			Arg  okxArg    `json:"arg"`
			Data []okxData `json:"data"`
		}

		args := make([]okxArg, 0, len(symbols))
		for _, symbol := range symbols {
			base := normalizeBaseSymbol(symbol)
			if base == "" {
				continue
			}
			args = append(args, okxArg{Channel: "tickers", InstID: base + "-USDT"})
		}

		subscribe := struct {
			Op   string   `json:"op"`
			Args []okxArg `json:"args"`
		}{
			Op:   "subscribe",
			Args: args,
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := dialWebSocket(wsURL)
			if err != nil {
				log.Error().Err(err).Str("exchange", "OKX").Msg("ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", "OKX").Msg("ws subscribe failed")
				_ = conn.Close()
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			stopPing := make(chan struct{})
			go func() {
				ticker := time.NewTicker(25 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-stopPing:
						return
					case <-ticker.C:
						if err := conn.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
							return
						}
					}
				}
			}()

			readFailed := false
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					readFailed = true
					break
				}

				var payload okxMessage
				if err := json.Unmarshal(message, &payload); err != nil {
					continue
				}
				if len(payload.Data) == 0 {
					continue
				}

				parts := strings.Split(payload.Arg.InstID, "-")
				if len(parts) < 1 {
					continue
				}

				price, err := decimal.NewFromString(payload.Data[0].Last)
				if err != nil {
					continue
				}

				callback(models.TickerData{
					Exchange:  models.Okx,
					Symbol:    strings.ToUpper(parts[0]),
					Price:     price,
					Timestamp: time.Now(),
				})
			}

			close(stopPing)
			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "OKX").Msg("ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunBinanceFuturesTicker(ctx context.Context, usdtSymbol string, callback TickerCallback) {
	go runBinanceMiniTicker(ctx, "wss://fstream.binance.com/ws/!miniTicker@arr", usdtSymbol, models.BinanceFutures, callback)
}

func RunBybitFuturesTicker(ctx context.Context, symbols []string, callback TickerCallback) {
	go runBybitTicker(ctx, "wss://stream.bybit.com/v5/public/linear", symbols, models.BybitFutures, callback)
}

func runBinanceMiniTicker(ctx context.Context, wsURL string, usdtSymbol string, exchange models.Exchange, callback TickerCallback) {
	tickerSuffix := strings.ToUpper(strings.TrimSpace(usdtSymbol))
	if tickerSuffix == "" {
		tickerSuffix = "USDT"
	}

	type binanceTicker struct {
		Symbol string `json:"s"`
		Close  string `json:"c"`
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := dialWebSocket(wsURL)
		if err != nil {
			log.Error().Err(err).Str("exchange", exchange.String()).Msg("ws connect failed")
			sleepOrDone(ctx, 3*time.Second)
			continue
		}

		stopPing := make(chan struct{})
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-stopPing:
					return
				case <-ticker.C:
					if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
						return
					}
				}
			}
		}()

		readFailed := false
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				readFailed = true
				break
			}

			var payload []binanceTicker
			if err := json.Unmarshal(message, &payload); err != nil {
				continue
			}

			for _, item := range payload {
				if !strings.HasSuffix(item.Symbol, tickerSuffix) {
					continue
				}
				base := strings.TrimSuffix(strings.ToUpper(item.Symbol), tickerSuffix)
				if base == "" {
					continue
				}

				price, err := decimal.NewFromString(item.Close)
				if err != nil {
					continue
				}

				callback(models.TickerData{
					Exchange:  exchange,
					Symbol:    base,
					Price:     price,
					Timestamp: time.Now(),
				})
			}
		}

		close(stopPing)
		_ = conn.Close()
		if readFailed {
			log.Error().Str("exchange", exchange.String()).Msg("ws disconnected, reconnecting")
			sleepOrDone(ctx, 3*time.Second)
		}
	}
}

func runBybitTicker(ctx context.Context, wsURL string, symbols []string, exchange models.Exchange, callback TickerCallback) {
	topics := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		base := normalizeBaseSymbol(symbol)
		if base == "" {
			continue
		}
		topics = append(topics, "tickers."+base+"USDT")
	}

	type bybitData struct {
		LastPrice string `json:"lastPrice"`
	}

	type bybitMessage struct {
		Topic string    `json:"topic"`
		Data  bybitData `json:"data"`
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := dialWebSocket(wsURL)
		if err != nil {
			log.Error().Err(err).Str("exchange", exchange.String()).Msg("ws connect failed")
			sleepOrDone(ctx, 3*time.Second)
			continue
		}

		for i := 0; i < len(topics); i += 10 {
			end := i + 10
			if end > len(topics) {
				end = len(topics)
			}

			subscribe := struct {
				Op   string   `json:"op"`
				Args []string `json:"args"`
			}{
				Op:   "subscribe",
				Args: topics[i:end],
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", exchange.String()).Msg("ws subscribe failed")
				_ = conn.Close()
				sleepOrDone(ctx, 3*time.Second)
				continue
			}
		}

		stopPing := make(chan struct{})
		go func() {
			ticker := time.NewTicker(20 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-stopPing:
					return
				case <-ticker.C:
					if err := conn.WriteJSON(map[string]string{"op": "ping"}); err != nil {
						return
					}
				}
			}
		}()

		readFailed := false
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				readFailed = true
				break
			}

			var payload bybitMessage
			if err := json.Unmarshal(message, &payload); err != nil {
				continue
			}
			if payload.Topic == "" {
				continue
			}

			topic := strings.TrimPrefix(payload.Topic, "tickers.")
			if topic == payload.Topic {
				continue
			}

			base := strings.TrimSuffix(strings.ToUpper(topic), "USDT")
			if base == "" {
				continue
			}

			price, err := decimal.NewFromString(payload.Data.LastPrice)
			if err != nil {
				continue
			}

			callback(models.TickerData{
				Exchange:  exchange,
				Symbol:    base,
				Price:     price,
				Timestamp: time.Now(),
			})
		}

		close(stopPing)
		_ = conn.Close()
		if readFailed {
			log.Error().Str("exchange", exchange.String()).Msg("ws disconnected, reconnecting")
			sleepOrDone(ctx, 3*time.Second)
		}
	}
}

func dialWebSocket(wsURL string) (*websocket.Conn, error) {
	parsed, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("invalid ws url: %w", err)
	}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(parsed.String(), nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func normalizeBaseSymbol(symbol string) string {
	value := strings.ToUpper(strings.TrimSpace(symbol))
	value = strings.TrimPrefix(value, "KRW-")
	value = strings.TrimSuffix(value, "_KRW")
	value = strings.TrimSuffix(value, "-USDT")
	value = strings.TrimSuffix(value, "USDT")
	return value
}

func sleepOrDone(ctx context.Context, delay time.Duration) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

func mustJSONRaw(v []string) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
