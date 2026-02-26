package transfer

import (
	"strings"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

type NetworkMapping struct {
	Display    string
	Binance    string
	Upbit      string
	Bithumb    string
	Bybit      string
	Bitget     string
	OKX        string
	Gate       string
	ExplorerTx string
}

var NETWORK_MAP = []NetworkMapping{
	{Display: "Ethereum", Binance: "ETH", Upbit: "ETH", Bithumb: "ETH", Bybit: "ETH", Bitget: "ETH", OKX: "ETH-ERC20", Gate: "ETH", ExplorerTx: "https://etherscan.io/tx/"},
	{Display: "BSC", Binance: "BSC", Upbit: "BSC", Bithumb: "BSC", Bybit: "BSC", Bitget: "BSC", OKX: "BSC-BEP20", Gate: "BSC_BEP20", ExplorerTx: "https://bscscan.com/tx/"},
	{Display: "Tron", Binance: "TRX", Upbit: "TRX", Bithumb: "TRX", Bybit: "TRX", Bitget: "TRC20", OKX: "TRON-TRC20", Gate: "TRC20", ExplorerTx: "https://tronscan.org/#/transaction/"},
	{Display: "Solana", Binance: "SOL", Upbit: "SOL", Bithumb: "SOL", Bybit: "SOL", Bitget: "SOL", OKX: "SOL-Solana", Gate: "SOL", ExplorerTx: "https://solscan.io/tx/"},
	{Display: "Polygon", Binance: "MATIC", Upbit: "MATIC", Bithumb: "POL", Bybit: "MATIC", Bitget: "POLYGON", OKX: "MATIC-Polygon", Gate: "MATIC", ExplorerTx: "https://polygonscan.com/tx/"},
	{Display: "Arbitrum", Binance: "ARBITRUM", Upbit: "ARBITRUM", Bithumb: "ARB_ETH", Bybit: "ARBONE", Bitget: "ARBITRUM", OKX: "ETH-Arbitrum One", Gate: "ARBITRUM", ExplorerTx: "https://arbiscan.io/tx/"},
	{Display: "Optimism", Binance: "OPTIMISM", Upbit: "OPTIMISM", Bithumb: "OP_ETH", Bybit: "OP", Bitget: "OPTIMISM", OKX: "ETH-Optimism", Gate: "OPTIMISM", ExplorerTx: "https://optimistic.etherscan.io/tx/"},
	{Display: "Avalanche C-Chain", Binance: "AVAXC", Upbit: "AVAXC", Bithumb: "AVAX", Bybit: "AVAXC", Bitget: "CAVAX", OKX: "AVAX-C", Gate: "AVAX_C", ExplorerTx: "https://snowtrace.io/tx/"},
	{Display: "Bitcoin", Binance: "BTC", Upbit: "BTC", Bithumb: "BTC", Bybit: "BTC", Bitget: "BTC", OKX: "BTC-Bitcoin", Gate: "BTC", ExplorerTx: "https://mempool.space/tx/"},
	{Display: "Ripple", Binance: "XRP", Upbit: "XRP", Bithumb: "XRP", Bybit: "XRP", Bitget: "XRP", OKX: "XRP-Ripple", Gate: "XRP", ExplorerTx: "https://xrpscan.com/tx/"},
	{Display: "Stellar", Binance: "XLM", Upbit: "XLM", Bithumb: "XLM", Bybit: "XLM", Bitget: "XLM", OKX: "XLM-Stellar", Gate: "XLM", ExplorerTx: "https://stellarchain.io/transactions/"},
	{Display: "Cosmos", Binance: "ATOM", Upbit: "ATOM", Bithumb: "ATOM", Bybit: "ATOM", Bitget: "ATOM", OKX: "ATOM-Cosmos", Gate: "ATOM", ExplorerTx: "https://www.mintscan.io/cosmos/tx/"},
	{Display: "Kaia", Binance: "KLAY", Upbit: "KLAY", Bithumb: "KAIA", Bybit: "KLAY", Bitget: "KLAY", OKX: "KLAY-Klaytn", Gate: "KLAY", ExplorerTx: "https://kaiascope.com/tx/"},
	{Display: "EOS", Binance: "EOS", Upbit: "EOS", Bithumb: "VAULTA", Bybit: "EOS", Bitget: "EOS", OKX: "EOS-EOS", Gate: "EOS", ExplorerTx: "https://bloks.io/transaction/"},
	{Display: "Near", Binance: "NEAR", Upbit: "NEAR", Bithumb: "NEAR", Bybit: "NEAR", Bitget: "NEAR", OKX: "NEAR-NEAR", Gate: "NEAR", ExplorerTx: "https://nearblocks.io/txns/"},
	{Display: "Aptos", Binance: "APT", Upbit: "APT", Bithumb: "APT", Bybit: "APT", Bitget: "APT", OKX: "APT-Aptos", Gate: "APT", ExplorerTx: "https://aptoscan.com/transaction/"},
	{Display: "Sui", Binance: "SUI", Upbit: "SUI", Bithumb: "SUI", Bybit: "SUI", Bitget: "SUI", OKX: "SUI-Sui", Gate: "SUI", ExplorerTx: "https://suiscan.xyz/mainnet/tx/"},
	{Display: "Hedera", Binance: "HBAR", Upbit: "HBAR", Bithumb: "HBAR", Bybit: "HBAR", Bitget: "HBAR", OKX: "HBAR-Hedera", Gate: "HBAR", ExplorerTx: "https://hashscan.io/mainnet/transaction/"},
	{Display: "Base", Binance: "BASE", Upbit: "BASE", Bithumb: "BASE_ETH", Bybit: "BASE", Bitget: "BASE", OKX: "BASE-Base", Gate: "BASE", ExplorerTx: "https://basescan.org/tx/"},
	{Display: "TON", Binance: "TON", Upbit: "TON", Bithumb: "TON", Bybit: "TON", Bitget: "TON", OKX: "TON-TON", Gate: "TON", ExplorerTx: "https://tonscan.org/tx/"},
}

var NETWORK_ALIASES = map[string]string{
	"BASENET":          "Base",
	"BETH":             "Base",
	"BASEEVM":          "Base",
	"BASE_EVM":         "Base",
	"ARB":              "Arbitrum",
	"ARBETH":           "Arbitrum",
	"ARBI":             "Arbitrum",
	"ARBITRUMONE":      "Arbitrum",
	"ARB_ONE":          "Arbitrum",
	"ARBONE":           "Arbitrum",
	"ARB_ETH":          "Arbitrum",
	"ETH-ARBITRUM ONE": "Arbitrum",
	"OPT":              "Optimism",
	"OPETH":            "Optimism",
	"OP_ETH":           "Optimism",
	"POL":              "Polygon",
	"POLYGON":          "Polygon",
	"POLYGON_EVM":      "Polygon",
	"MATIC-POLYGON":    "Polygon",
	"AVAX":             "Avalanche C-Chain",
	"CAVAX":            "Avalanche C-Chain",
	"AVAX_C":           "Avalanche C-Chain",
	"AVAX-C":           "Avalanche C-Chain",
	"CCHAIN":           "Avalanche C-Chain",
	"BEP20":            "BSC",
	"BSC_BEP20":        "BSC",
	"BSC-BEP20":        "BSC",
	"BNB":              "BSC",
	"TRC20":            "Tron",
	"TRON":             "Tron",
	"TRON-TRC20":       "Tron",
	"SOLANA":           "Solana",
	"SOL-SOLANA":       "Solana",
	"KAIA":             "Kaia",
	"KLAYTN":           "Kaia",
	"KLAY-KLAYTN":      "Kaia",
	"VAULTA":           "EOS",
	"EOS-EOS":          "EOS",
	"TONCOIN":          "TON",
	"TON-TON":          "TON",
}

type CommonNetwork struct {
	DisplayName string
	FromNetwork string
	ToNetwork   string
	WithdrawFee *float64
	WithdrawMin *float64
	NeedsMemo   bool
}

var DefaultChainPriority = []string{
	"Solana", "Tron", "Arbitrum", "Optimism", "BSC", "Base",
	"Avalanche C-Chain", "Polygon", "TON", "Ethereum",
}

func NormalizeNetwork(exchangeNetwork string) string {
	for _, mapping := range NETWORK_MAP {
		if strings.EqualFold(exchangeNetwork, mapping.Display) ||
			strings.EqualFold(exchangeNetwork, mapping.Binance) ||
			strings.EqualFold(exchangeNetwork, mapping.Upbit) ||
			strings.EqualFold(exchangeNetwork, mapping.Bithumb) ||
			strings.EqualFold(exchangeNetwork, mapping.Bybit) ||
			strings.EqualFold(exchangeNetwork, mapping.Bitget) ||
			strings.EqualFold(exchangeNetwork, mapping.OKX) ||
			strings.EqualFold(exchangeNetwork, mapping.Gate) {
			return mapping.Display
		}
	}

	if canonical, ok := NETWORK_ALIASES[strings.ToUpper(exchangeNetwork)]; ok {
		return canonical
	}

	return exchangeNetwork
}

func FindCommonNetworks(fromExchange, toExchange models.Exchange, fromNetworks, toNetworks []NetworkInfo) []CommonNetwork {
	commons := make([]CommonNetwork, 0)

	for _, fromNet := range fromNetworks {
		if !fromNet.WithdrawEnabled {
			continue
		}

		fromCanonical := normalizeForExchange(fromExchange, fromNet.Network)

		for _, toNet := range toNetworks {
			if !toNet.DepositEnabled {
				continue
			}

			toCanonical := normalizeForExchange(toExchange, toNet.Network)
			if !strings.EqualFold(fromCanonical, toCanonical) {
				continue
			}

			mappedFrom := exchangeNetworkName(fromExchange, fromCanonical)
			if mappedFrom == "" {
				mappedFrom = fromNet.Network
			}

			mappedTo := exchangeNetworkName(toExchange, toCanonical)
			if mappedTo == "" {
				mappedTo = toNet.Network
			}

			commons = append(commons, CommonNetwork{
				DisplayName: fromCanonical,
				FromNetwork: mappedFrom,
				ToNetwork:   mappedTo,
				WithdrawFee: fromNet.WithdrawFee,
				WithdrawMin: fromNet.WithdrawMin,
				NeedsMemo:   toNet.NeedsMemo,
			})
			break
		}
	}

	return commons
}

func AutoSelectNetwork(commons []CommonNetwork) *CommonNetwork {
	if len(commons) == 0 {
		return nil
	}

	for _, preferred := range DefaultChainPriority {
		for i := range commons {
			if strings.EqualFold(commons[i].DisplayName, preferred) {
				return &commons[i]
			}
		}
	}

	return &commons[0]
}

func normalizeForExchange(exchange models.Exchange, network string) string {
	for _, mapping := range NETWORK_MAP {
		if strings.EqualFold(networkForExchange(exchange, mapping), network) {
			return mapping.Display
		}
	}

	if canonical, ok := NETWORK_ALIASES[strings.ToUpper(network)]; ok {
		return canonical
	}

	return network
}

func exchangeNetworkName(exchange models.Exchange, display string) string {
	for _, mapping := range NETWORK_MAP {
		if strings.EqualFold(mapping.Display, display) {
			return networkForExchange(exchange, mapping)
		}
	}

	return ""
}

func networkForExchange(exchange models.Exchange, mapping NetworkMapping) string {
	switch exchange {
	case models.Binance:
		return mapping.Binance
	case models.Upbit:
		return mapping.Upbit
	case models.Bithumb:
		return mapping.Bithumb
	case models.Bybit:
		return mapping.Bybit
	case models.Bitget:
		return mapping.Bitget
	case models.Okx:
		return mapping.OKX
	case models.Gate:
		return mapping.Gate
	case models.BinanceFutures:
		return mapping.Binance
	case models.BybitFutures:
		return mapping.Bybit
	default:
		return ""
	}
}
