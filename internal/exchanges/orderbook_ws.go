package exchanges

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hyeokx/Go_GAPTUI/internal/models"
	"github.com/rs/zerolog/log"
)

type OrderbookCallback func(exchange models.Exchange, symbol string, asks []models.OrderbookEntry, bids []models.OrderbookEntry)

type OrderbookManager struct {
	mu    sync.RWMutex
	books map[string]map[models.Exchange]*OrderbookSnapshot
}

type OrderbookSnapshot struct {
	Asks []models.OrderbookEntry
	Bids []models.OrderbookEntry
}

func NewOrderbookManager() *OrderbookManager {
	return &OrderbookManager{books: make(map[string]map[models.Exchange]*OrderbookSnapshot)}
}

func (m *OrderbookManager) Update(exchange models.Exchange, symbol string, asks, bids []models.OrderbookEntry) {
	key := orderbookSymbolKey(symbol)
	if key == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.books[key]; !ok {
		m.books[key] = make(map[models.Exchange]*OrderbookSnapshot)
	}
	m.books[key][exchange] = &OrderbookSnapshot{
		Asks: cloneEntries(asks),
		Bids: cloneEntries(bids),
	}
}

func (m *OrderbookManager) Get(exchange models.Exchange, symbol string) *OrderbookSnapshot {
	key := orderbookSymbolKey(symbol)
	if key == "" {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	byExchange, ok := m.books[key]
	if !ok {
		return nil
	}
	snapshot, ok := byExchange[exchange]
	if !ok || snapshot == nil {
		return nil
	}

	return &OrderbookSnapshot{
		Asks: cloneEntries(snapshot.Asks),
		Bids: cloneEntries(snapshot.Bids),
	}
}

func (m *OrderbookManager) GetAll(symbol string) map[models.Exchange]*OrderbookSnapshot {
	key := orderbookSymbolKey(symbol)
	if key == "" {
		return map[models.Exchange]*OrderbookSnapshot{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	byExchange, ok := m.books[key]
	if !ok {
		return map[models.Exchange]*OrderbookSnapshot{}
	}

	out := make(map[models.Exchange]*OrderbookSnapshot, len(byExchange))
	for exchange, snapshot := range byExchange {
		if snapshot == nil {
			continue
		}
		out[exchange] = &OrderbookSnapshot{
			Asks: cloneEntries(snapshot.Asks),
			Bids: cloneEntries(snapshot.Bids),
		}
	}
	return out
}

func RunUpbitOrderbook(ctx context.Context, symbols []string, callback OrderbookCallback) {
	go func() {
		const wsURL = "wss://api.upbit.com/websocket/v1"

		type upbitUnit struct {
			AskPrice float64 `json:"ask_price"`
			BidPrice float64 `json:"bid_price"`
			AskSize  float64 `json:"ask_size"`
			BidSize  float64 `json:"bid_size"`
		}

		type upbitMessage struct {
			Code  string      `json:"code"`
			Units []upbitUnit `json:"orderbook_units"`
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
			{"ticket": json.RawMessage(`"orderbook-stream"`)},
			{"type": json.RawMessage(`"orderbook"`), "codes": mustJSONRaw(codes)},
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := dialWebSocket(wsURL)
			if err != nil {
				log.Error().Err(err).Str("exchange", "Upbit").Msg("orderbook ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", "Upbit").Msg("orderbook ws subscribe failed")
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

				var payload upbitMessage
				if err := json.Unmarshal(message, &payload); err != nil {
					continue
				}
				parts := strings.Split(payload.Code, "-")
				if len(parts) < 2 {
					continue
				}

				symbol := strings.ToUpper(parts[len(parts)-1])
				asks := make([]models.OrderbookEntry, 0, len(payload.Units))
				bids := make([]models.OrderbookEntry, 0, len(payload.Units))
				for _, unit := range payload.Units {
					if unit.AskPrice > 0 && unit.AskSize > 0 {
						asks = append(asks, models.OrderbookEntry{Price: unit.AskPrice, Quantity: unit.AskSize})
					}
					if unit.BidPrice > 0 && unit.BidSize > 0 {
						bids = append(bids, models.OrderbookEntry{Price: unit.BidPrice, Quantity: unit.BidSize})
					}
				}
				sortBook(&asks, &bids, 0)
				callback(models.Upbit, symbol, asks, bids)
			}

			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "Upbit").Msg("orderbook ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunBithumbOrderbook(ctx context.Context, symbols []string, manager *OrderbookManager, callback OrderbookCallback) {
	go func() {
		const wsURL = "wss://pubwss.bithumb.com/pub/ws"

		type bithumbItem struct {
			Symbol    string `json:"symbol"`
			OrderType string `json:"orderType"`
			Price     string `json:"price"`
			Quantity  string `json:"quantity"`
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
			Type:    "orderbookdepth",
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
				log.Error().Err(err).Str("exchange", "Bithumb").Msg("orderbook ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", "Bithumb").Msg("orderbook ws subscribe failed")
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
				if payload.Type != "orderbookdepth" {
					continue
				}

				updates := make(map[string][]bithumbItem)
				for _, item := range payload.Content.List {
					parts := strings.Split(strings.ToUpper(item.Symbol), "_")
					if len(parts) < 1 || parts[0] == "" {
						continue
					}
					updates[parts[0]] = append(updates[parts[0]], item)
				}

				for symbol, items := range updates {
					snapshot := manager.Get(models.Bithumb, symbol)
					asks := []models.OrderbookEntry{}
					bids := []models.OrderbookEntry{}
					if snapshot != nil {
						asks = cloneEntries(snapshot.Asks)
						bids = cloneEntries(snapshot.Bids)
					}

					for _, item := range items {
						price, ok := parseFloat(item.Price)
						if !ok || price <= 0 {
							continue
						}
						qty, ok := parseFloat(item.Quantity)
						if !ok {
							continue
						}

						side := strings.ToLower(item.OrderType)
						switch side {
						case "ask":
							asks = upsertLevel(asks, price, qty)
						case "bid":
							bids = upsertLevel(bids, price, qty)
						}
					}

					sortBook(&asks, &bids, 30)
					manager.Update(models.Bithumb, symbol, asks, bids)
					callback(models.Bithumb, symbol, asks, bids)
				}
			}

			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "Bithumb").Msg("orderbook ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunBinanceOrderbook(ctx context.Context, symbols []string, usdtSymbol string, callback OrderbookCallback) {
	go func() {
		suffix := strings.ToUpper(strings.TrimSpace(usdtSymbol))
		if suffix == "" {
			suffix = "USDT"
		}

		streams := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			base := normalizeBaseSymbol(symbol)
			if base == "" {
				continue
			}
			streams = append(streams, strings.ToLower(base+suffix)+"@depth20@100ms")
		}
		if len(streams) == 0 {
			return
		}

		wsURL := "wss://stream.binance.com:9443/stream?streams=" + strings.Join(streams, "/")

		type binanceData struct {
			Bids [][]string `json:"bids"`
			Asks [][]string `json:"asks"`
		}

		type binanceMessage struct {
			Stream string      `json:"stream"`
			Data   binanceData `json:"data"`
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			conn, err := dialWebSocket(wsURL)
			if err != nil {
				log.Error().Err(err).Str("exchange", "Binance").Msg("orderbook ws connect failed")
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

				var payload binanceMessage
				if err := json.Unmarshal(message, &payload); err != nil {
					continue
				}

				pair := payload.Stream
				if idx := strings.Index(pair, "@"); idx >= 0 {
					pair = pair[:idx]
				}
				pair = strings.ToUpper(pair)
				if !strings.HasSuffix(pair, suffix) {
					continue
				}
				symbol := strings.TrimSuffix(pair, suffix)
				if symbol == "" {
					continue
				}

				asks := parseBookLevels(payload.Data.Asks)
				bids := parseBookLevels(payload.Data.Bids)
				sortBook(&asks, &bids, 20)
				callback(models.Binance, symbol, asks, bids)
			}

			close(stopPing)
			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "Binance").Msg("orderbook ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunBybitOrderbook(ctx context.Context, symbols []string, manager *OrderbookManager, callback OrderbookCallback) {
	go func() {
		const wsURL = "wss://stream.bybit.com/v5/public/spot"

		topics := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			base := normalizeBaseSymbol(symbol)
			if base == "" {
				continue
			}
			topics = append(topics, "orderbook.50."+base+"USDT")
		}

		type bybitData struct {
			Bids [][]string `json:"b"`
			Asks [][]string `json:"a"`
		}

		type bybitMessage struct {
			Topic string    `json:"topic"`
			Type  string    `json:"type"`
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
				log.Error().Err(err).Str("exchange", "Bybit").Msg("orderbook ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			for i := 0; i < len(topics); i += 10 {
				end := min(i+10, len(topics))

				subscribe := struct {
					Op   string   `json:"op"`
					Args []string `json:"args"`
				}{
					Op:   "subscribe",
					Args: topics[i:end],
				}

				if err := conn.WriteJSON(subscribe); err != nil {
					log.Error().Err(err).Str("exchange", "Bybit").Msg("orderbook ws subscribe failed")
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

				topic := strings.TrimPrefix(payload.Topic, "orderbook.50.")
				if topic == payload.Topic {
					continue
				}
				symbol := strings.TrimSuffix(strings.ToUpper(topic), "USDT")
				if symbol == "" {
					continue
				}

				var asks []models.OrderbookEntry
				var bids []models.OrderbookEntry

				switch strings.ToLower(payload.Type) {
				case "snapshot":
					asks = parseBookLevels(payload.Data.Asks)
					bids = parseBookLevels(payload.Data.Bids)
				case "delta":
					snapshot := manager.Get(models.Bybit, symbol)
					if snapshot != nil {
						asks = cloneEntries(snapshot.Asks)
						bids = cloneEntries(snapshot.Bids)
					}
					asks = applyDeltas(asks, payload.Data.Asks)
					bids = applyDeltas(bids, payload.Data.Bids)
				default:
					continue
				}

				sortBook(&asks, &bids, 30)
				manager.Update(models.Bybit, symbol, asks, bids)
				callback(models.Bybit, symbol, asks, bids)
			}

			close(stopPing)
			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "Bybit").Msg("orderbook ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func RunOKXOrderbook(ctx context.Context, symbols []string, callback OrderbookCallback) {
	go func() {
		const wsURL = "wss://ws.okx.com:8443/ws/v5/public"

		type okxArg struct {
			Channel string `json:"channel"`
			InstID  string `json:"instId"`
		}

		type okxBook struct {
			Asks [][]string `json:"asks"`
			Bids [][]string `json:"bids"`
		}

		type okxMessage struct {
			Arg  okxArg     `json:"arg"`
			Data []okxBook  `json:"data"`
		}

		args := make([]okxArg, 0, len(symbols))
		for _, symbol := range symbols {
			base := normalizeBaseSymbol(symbol)
			if base == "" {
				continue
			}
			args = append(args, okxArg{Channel: "books5", InstID: base + "-USDT"})
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
				log.Error().Err(err).Str("exchange", "OKX").Msg("orderbook ws connect failed")
				sleepOrDone(ctx, 3*time.Second)
				continue
			}

			if err := conn.WriteJSON(subscribe); err != nil {
				log.Error().Err(err).Str("exchange", "OKX").Msg("orderbook ws subscribe failed")
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
				if len(parts) < 1 || parts[0] == "" {
					continue
				}
				symbol := strings.ToUpper(parts[0])
				asks := parseBookLevels(payload.Data[0].Asks)
				bids := parseBookLevels(payload.Data[0].Bids)
				sortBook(&asks, &bids, 5)
				callback(models.Okx, symbol, asks, bids)
			}

			close(stopPing)
			_ = conn.Close()
			if readFailed {
				log.Error().Str("exchange", "OKX").Msg("orderbook ws disconnected, reconnecting")
				sleepOrDone(ctx, 3*time.Second)
			}
		}
	}()
}

func orderbookSymbolKey(symbol string) string {
	return strings.ToUpper(normalizeBaseSymbol(symbol))
}

func cloneEntries(entries []models.OrderbookEntry) []models.OrderbookEntry {
	if len(entries) == 0 {
		return []models.OrderbookEntry{}
	}
	out := make([]models.OrderbookEntry, len(entries))
	copy(out, entries)
	return out
}

func parseFloat(value string) (float64, bool) {
	v := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
	if v == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func parseBookLevels(levels [][]string) []models.OrderbookEntry {
	out := make([]models.OrderbookEntry, 0, len(levels))
	for _, level := range levels {
		if len(level) < 2 {
			continue
		}
		price, ok := parseFloat(level[0])
		if !ok || price <= 0 {
			continue
		}
		qty, ok := parseFloat(level[1])
		if !ok || qty <= 0 {
			continue
		}
		out = append(out, models.OrderbookEntry{Price: price, Quantity: qty})
	}
	return out
}

func upsertLevel(levels []models.OrderbookEntry, price, qty float64) []models.OrderbookEntry {
	for i := range levels {
		if levels[i].Price == price {
			if qty <= 0 {
				return append(levels[:i], levels[i+1:]...)
			}
			levels[i].Quantity = qty
			return levels
		}
	}
	if qty <= 0 {
		return levels
	}
	return append(levels, models.OrderbookEntry{Price: price, Quantity: qty})
}

func applyDeltas(levels []models.OrderbookEntry, deltas [][]string) []models.OrderbookEntry {
	out := levels
	for _, delta := range deltas {
		if len(delta) < 2 {
			continue
		}
		price, ok := parseFloat(delta[0])
		if !ok || price <= 0 {
			continue
		}
		qty, ok := parseFloat(delta[1])
		if !ok {
			continue
		}
		out = upsertLevel(out, price, qty)
	}
	return out
}

func sortBook(asks *[]models.OrderbookEntry, bids *[]models.OrderbookEntry, maxLevels int) {
	sort.Slice(*asks, func(i, j int) bool {
		return (*asks)[i].Price < (*asks)[j].Price
	})
	sort.Slice(*bids, func(i, j int) bool {
		return (*bids)[i].Price > (*bids)[j].Price
	})

	if maxLevels > 0 {
		if len(*asks) > maxLevels {
			*asks = (*asks)[:maxLevels]
		}
		if len(*bids) > maxLevels {
			*bids = (*bids)[:maxLevels]
		}
	}
}
