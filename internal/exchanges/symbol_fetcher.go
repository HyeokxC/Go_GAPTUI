package exchanges

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

func FetchSymbols(ctx context.Context, client *http.Client) (symbols []string, koreanNames map[string]string, err error) {
	if client == nil {
		return nil, nil, fmt.Errorf("http client is nil")
	}

	binance, err := fetchBinanceSymbols(ctx, client)
	if err != nil {
		return nil, nil, err
	}

	upbit, upbitNames, err := fetchUpbitSymbols(ctx, client)
	if err != nil {
		return nil, nil, err
	}

	bithumb, err := fetchBithumbSymbols(ctx, client)
	if err != nil {
		return nil, nil, err
	}

	bybit, err := fetchBybitSymbols(ctx, client)
	if err != nil {
		return nil, nil, err
	}

	okx, err := fetchOKXSymbols(ctx, client)
	if err != nil {
		return nil, nil, err
	}

	domestic := intersectSets(upbit, bithumb)
	overseas := unionSets(binance, bybit, okx)
	finalSet := intersectSets(domestic, overseas)

	symbols = make([]string, 0, len(finalSet))
	for symbol := range finalSet {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	koreanNames = make(map[string]string, len(symbols))
	for _, symbol := range symbols {
		if name, ok := upbitNames[symbol]; ok {
			koreanNames[symbol] = name
		} else {
			koreanNames[symbol] = ""
		}
	}

	return symbols, koreanNames, nil
}

func fetchBinanceSymbols(ctx context.Context, client *http.Client) (map[string]struct{}, error) {
	type symbolInfo struct {
		BaseAsset  string `json:"baseAsset"`
		QuoteAsset string `json:"quoteAsset"`
		Status     string `json:"status"`
	}
	type response struct {
		Symbols []symbolInfo `json:"symbols"`
	}

	var payload response
	if err := fetchJSON(ctx, client, "https://api.binance.com/api/v3/exchangeInfo", &payload); err != nil {
		return nil, fmt.Errorf("fetch binance symbols: %w", err)
	}

	set := make(map[string]struct{})
	for _, item := range payload.Symbols {
		if item.QuoteAsset != "USDT" || item.Status != "TRADING" {
			continue
		}
		base := strings.ToUpper(strings.TrimSpace(item.BaseAsset))
		if base != "" {
			set[base] = struct{}{}
		}
	}

	return set, nil
}

func fetchUpbitSymbols(ctx context.Context, client *http.Client) (map[string]struct{}, map[string]string, error) {
	type marketInfo struct {
		Market     string `json:"market"`
		KoreanName string `json:"korean_name"`
	}

	var payload []marketInfo
	if err := fetchJSON(ctx, client, "https://api.upbit.com/v1/market/all", &payload); err != nil {
		return nil, nil, fmt.Errorf("fetch upbit symbols: %w", err)
	}

	set := make(map[string]struct{})
	names := make(map[string]string)
	for _, item := range payload {
		market := strings.ToUpper(strings.TrimSpace(item.Market))
		if !strings.HasPrefix(market, "KRW-") {
			continue
		}
		base := strings.TrimPrefix(market, "KRW-")
		if base == "" {
			continue
		}
		set[base] = struct{}{}
		names[base] = strings.TrimSpace(item.KoreanName)
	}

	return set, names, nil
}

func fetchBithumbSymbols(ctx context.Context, client *http.Client) (map[string]struct{}, error) {
	type response struct {
		Data map[string]json.RawMessage `json:"data"`
	}

	var payload response
	if err := fetchJSON(ctx, client, "https://api.bithumb.com/public/ticker/ALL_KRW", &payload); err != nil {
		return nil, fmt.Errorf("fetch bithumb symbols: %w", err)
	}

	set := make(map[string]struct{})
	for key := range payload.Data {
		symbol := strings.ToUpper(strings.TrimSpace(key))
		if symbol == "" || symbol == "DATE" {
			continue
		}
		set[symbol] = struct{}{}
	}

	return set, nil
}

func fetchBybitSymbols(ctx context.Context, client *http.Client) (map[string]struct{}, error) {
	type instrument struct {
		BaseCoin  string `json:"baseCoin"`
		QuoteCoin string `json:"quoteCoin"`
		Status    string `json:"status"`
	}
	type listWrapper struct {
		List []instrument `json:"list"`
	}
	type response struct {
		Result listWrapper `json:"result"`
	}

	var payload response
	if err := fetchJSON(ctx, client, "https://api.bybit.com/v5/market/instruments-info?category=spot&status=Trading", &payload); err != nil {
		return nil, fmt.Errorf("fetch bybit symbols: %w", err)
	}

	set := make(map[string]struct{})
	for _, item := range payload.Result.List {
		if item.QuoteCoin != "USDT" || item.Status != "Trading" {
			continue
		}
		base := strings.ToUpper(strings.TrimSpace(item.BaseCoin))
		if base != "" {
			set[base] = struct{}{}
		}
	}

	return set, nil
}

func fetchOKXSymbols(ctx context.Context, client *http.Client) (map[string]struct{}, error) {
	type instrument struct {
		InstID string `json:"instId"`
		State  string `json:"state"`
	}
	type response struct {
		Data []instrument `json:"data"`
	}

	var payload response
	if err := fetchJSON(ctx, client, "https://www.okx.com/api/v5/public/instruments?instType=SPOT", &payload); err != nil {
		return nil, fmt.Errorf("fetch okx symbols: %w", err)
	}

	set := make(map[string]struct{})
	for _, item := range payload.Data {
		instID := strings.ToUpper(strings.TrimSpace(item.InstID))
		if item.State != "live" || !strings.HasSuffix(instID, "-USDT") {
			continue
		}
		parts := strings.Split(instID, "-")
		if len(parts) < 2 {
			continue
		}
		base := strings.TrimSpace(parts[0])
		if base != "" {
			set[base] = struct{}{}
		}
	}

	return set, nil
}

func fetchJSON(ctx context.Context, client *http.Client, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(out); err != nil {
		return err
	}

	return nil
}

func intersectSets(a, b map[string]struct{}) map[string]struct{} {
	if len(a) > len(b) {
		a, b = b, a
	}

	out := make(map[string]struct{})
	for k := range a {
		if _, ok := b[k]; ok {
			out[k] = struct{}{}
		}
	}

	return out
}

func unionSets(sets ...map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	for _, set := range sets {
		for k := range set {
			out[k] = struct{}{}
		}
	}
	return out
}
