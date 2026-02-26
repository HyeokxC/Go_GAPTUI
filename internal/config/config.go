package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pelletier/go-toml/v2"
)

type AppConfig struct {
	Symbols             []string        `toml:"symbols"`
	UsdtSymbol          string          `toml:"usdt_symbol"`
	ChannelSize         int             `toml:"channel_size"`
	AutoDiscoverSymbols bool            `toml:"auto_discover_symbols"`
	Exchange            ExchangeConfig  `toml:"exchange"`
	Dashboard           DashboardConfig `toml:"dashboard"`
	ApiKeys             ApiKeysConfig
	Proxy               ProxyConfig
}

type ExchangeConfig struct {
	PollIntervalMs     uint64 `toml:"poll_interval_ms"`
	RequestTimeoutSecs uint64 `toml:"request_timeout_secs"`
}

type DashboardConfig struct {
	RefreshMs        uint64 `toml:"refresh_ms"`
	StaleDisplaySecs uint64 `toml:"stale_display_secs"`
}

type ApiKeyPair struct {
	ApiKey     string
	SecretKey  string
	Passphrase string
}

type ApiKeysConfig struct {
	Binance *ApiKeyPair
	Upbit   *ApiKeyPair
	Bithumb *ApiKeyPair
	Bybit   *ApiKeyPair
	Bitget  *ApiKeyPair
	Okx     *ApiKeyPair
	Gate    *ApiKeyPair
}

type ProxyConfig struct {
	DefaultProxy string
	BinanceProxy string
	UpbitProxy   string
	BithumbProxy string
	BybitProxy   string
	OkxProxy     string
	Username     string
	Password     string
}

func Load(path string) (*AppConfig, error) {
	_ = godotenv.Load()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	cfg.ApiKeys = loadAPIKeysFromEnv()
	cfg.Proxy = loadProxyFromEnv()

	return &cfg, nil
}

func applyDefaults(cfg *AppConfig) {
	if cfg.UsdtSymbol == "" {
		cfg.UsdtSymbol = "USDT"
	}
	if cfg.ChannelSize == 0 {
		cfg.ChannelSize = 2000
	}
	if cfg.Exchange.PollIntervalMs == 0 {
		cfg.Exchange.PollIntervalMs = 500
	}
	if cfg.Exchange.RequestTimeoutSecs == 0 {
		cfg.Exchange.RequestTimeoutSecs = 10
	}
	if cfg.Dashboard.RefreshMs == 0 {
		cfg.Dashboard.RefreshMs = 200
	}
	if cfg.Dashboard.StaleDisplaySecs == 0 {
		cfg.Dashboard.StaleDisplaySecs = 30
	}
}

func loadAPIKeysFromEnv() ApiKeysConfig {
	return ApiKeysConfig{
		Binance: loadAPIKeyPair("BINANCE"),
		Upbit:   loadAPIKeyPair("UPBIT"),
		Bithumb: loadAPIKeyPair("BITHUMB"),
		Bybit:   loadAPIKeyPair("BYBIT"),
		Bitget:  loadAPIKeyPair("BITGET"),
		Okx:     loadAPIKeyPair("OKX"),
		Gate:    loadAPIKeyPair("GATE"),
	}
}

func loadAPIKeyPair(exchange string) *ApiKeyPair {
	apiKey := os.Getenv(exchange + "_API_KEY")
	secretKey := os.Getenv(exchange + "_SECRET_KEY")
	if apiKey == "" || secretKey == "" {
		return nil
	}

	return &ApiKeyPair{
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: os.Getenv(exchange + "_PASSPHRASE"),
	}
}

func loadProxyFromEnv() ProxyConfig {
	return ProxyConfig{
		DefaultProxy: os.Getenv("PROXY_URL"),
		BinanceProxy: os.Getenv("BINANCE_PROXY"),
		UpbitProxy:   os.Getenv("UPBIT_PROXY"),
		BithumbProxy: os.Getenv("BITHUMB_PROXY"),
		BybitProxy:   os.Getenv("BYBIT_PROXY"),
		OkxProxy:     os.Getenv("OKX_PROXY"),
		Username:     os.Getenv("PROXY_USERNAME"),
		Password:     os.Getenv("PROXY_PASSWORD"),
	}
}

func (p ProxyConfig) ProxyFor(exchange string) string {
	switch strings.ToUpper(exchange) {
	case "BINANCE":
		if p.BinanceProxy != "" {
			return p.BinanceProxy
		}
	case "UPBIT":
		if p.UpbitProxy != "" {
			return p.UpbitProxy
		}
	case "BITHUMB":
		if p.BithumbProxy != "" {
			return p.BithumbProxy
		}
	case "BYBIT":
		if p.BybitProxy != "" {
			return p.BybitProxy
		}
	case "OKX":
		if p.OkxProxy != "" {
			return p.OkxProxy
		}
	}

	return p.DefaultProxy
}
