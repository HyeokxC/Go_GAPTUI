package ipc

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

type IpcSnapshot struct {
	CoinStates      map[string]*IpcCoinState     `json:"coin_states"`
	WalletStatus    map[string]*IpcWalletStatus  `json:"wallet_status"`
	KoreanNames     map[string]string            `json:"korean_names"`
	Logs            []IpcLogEntry                `json:"logs"`
	OrderbookInfo   map[string]*IpcOrderbookInfo `json:"orderbook_info"`
	ScenarioThreads []IpcLogThread               `json:"scenario_threads"`
	UsdtKrw         *float64                     `json:"usdt_krw"`
	UsdKrwForex     *float64                     `json:"usd_krw_forex"`
	LastTickerAgeMs int64                        `json:"last_ticker_age_ms"`
	ScenarioConfig  IpcScenarioConfig            `json:"scenario_config"`
	Transfer        *IpcTransferState            `json:"transfer"`
	TransferJobs    []IpcTransferJob             `json:"transfer_jobs"`
}

type IpcCoinState struct {
	Symbol              string   `json:"symbol"`
	UpbitPrice          *float64 `json:"upbit_price"`
	BithumbPrice        *float64 `json:"bithumb_price"`
	BinancePrice        *float64 `json:"binance_price"`
	BinanceKrw          *float64 `json:"binance_krw"`
	BybitPrice          *float64 `json:"bybit_price"`
	BybitKrw            *float64 `json:"bybit_krw"`
	OkxPrice            *float64 `json:"okx_price"`
	OkxKrw              *float64 `json:"okx_krw"`
	UpbitKimchi         *float64 `json:"upbit_kimchi"`
	BithumbKimchi       *float64 `json:"bithumb_kimchi"`
	BybitKimchiUp       *float64 `json:"bybit_kimchi_up"`
	BybitKimchiBt       *float64 `json:"bybit_kimchi_bt"`
	OkxKimchiUp         *float64 `json:"okx_kimchi_up"`
	OkxKimchiBt         *float64 `json:"okx_kimchi_bt"`
	DomesticGap         *float64 `json:"domestic_gap"`
	BinanceFuturesPrice *float64 `json:"binance_futures_price"`
	BybitFuturesPrice   *float64 `json:"bybit_futures_price"`
	FuturesBasis        *float64 `json:"futures_basis"`
}

type IpcWalletStatus struct {
	Upbit   *IpcExchangeWallet `json:"upbit"`
	Bithumb *IpcExchangeWallet `json:"bithumb"`
	Binance *IpcExchangeWallet `json:"binance"`
	Bybit   *IpcExchangeWallet `json:"bybit"`
	Okx     *IpcExchangeWallet `json:"okx"`
}

type IpcExchangeWallet struct {
	Deposit               bool     `json:"deposit"`
	Withdraw              bool     `json:"withdraw"`
	DepositBlockedChains  []string `json:"deposit_blocked_chains"`
	WithdrawBlockedChains []string `json:"withdraw_blocked_chains"`
}

type IpcLogEntry struct {
	Timestamp string `json:"timestamp"`
	Symbol    string `json:"symbol"`
	Message   string `json:"message"`
	LogType   string `json:"log_type"`
}

type IpcOrderbookInfo struct {
	BtSpread       float64            `json:"bt_spread"`
	UpSpread       float64            `json:"up_spread"`
	BtBuySlippage  float64            `json:"bt_buy_slippage"`
	UpSellSlippage float64            `json:"up_sell_slippage"`
	RealGapBtUp    float64            `json:"real_gap_bt_up"`
	RealGapUpBt    float64            `json:"real_gap_up_bt"`
	SurfaceGap     float64            `json:"surface_gap"`
	RealKimpUp     map[string]float64 `json:"real_kimp_up"`
	RealKimpBt     map[string]float64 `json:"real_kimp_bt"`
}

type IpcLogThread struct {
	ID              uint64           `json:"id"`
	Symbol          string           `json:"symbol"`
	Scenario        string           `json:"scenario"`
	Key             string           `json:"key"`
	MainMessage     string           `json:"main_message"`
	MainTimestamp   string           `json:"main_timestamp"`
	SubEntries      []IpcThreadEntry `json:"sub_entries"`
	IsActive        bool             `json:"is_active"`
	ClosedAt        *string          `json:"closed_at"`
	InitialValue    float64          `json:"initial_value"`
	LastLoggedValue float64          `json:"last_logged_value"`
	Label           string           `json:"label"`
}

type IpcThreadEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type IpcScenarioConfig struct {
	GapThresholdPercent  float64 `json:"gap_threshold_percent"`
	DomesticGapThreshold float64 `json:"domestic_gap_threshold"`
	FutBasisThreshold    float64 `json:"fut_basis_threshold"`
}

type IpcTransferState struct {
	SelectedCoin       string           `json:"selected_coin"`
	FromExchange       string           `json:"from_exchange"`
	ToExchange         string           `json:"to_exchange"`
	AvailableNetworks  []IpcNetworkInfo `json:"available_networks"`
	SelectedNetworkIdx *int             `json:"selected_network_idx"`
	Amount             string           `json:"amount"`
	Balance            *IpcBalanceInfo  `json:"balance"`
	DepositAddress     string           `json:"deposit_address"`
	DepositTag         string           `json:"deposit_tag"`
	AutoBuy            bool             `json:"auto_buy"`
	AutoSell           bool             `json:"auto_sell"`
	MarketOrderPending bool             `json:"market_order_pending"`
	MarketOrderResult  string           `json:"market_order_result"`
	ShowConfirmation   bool             `json:"show_confirmation"`
}

type IpcNetworkInfo struct {
	Network         string   `json:"network"`
	DisplayName     string   `json:"display_name"`
	DepositEnabled  bool     `json:"deposit_enabled"`
	WithdrawEnabled bool     `json:"withdraw_enabled"`
	WithdrawFee     *float64 `json:"withdraw_fee"`
	WithdrawMin     *float64 `json:"withdraw_min"`
	NeedsMemo       bool     `json:"needs_memo"`
}

type IpcBalanceInfo struct {
	Available float64 `json:"available"`
	Locked    float64 `json:"locked"`
}

type IpcTransferJob struct {
	ID           uint64        `json:"id"`
	Coin         string        `json:"coin"`
	Amount       float64       `json:"amount"`
	FromExchange string        `json:"from_exchange"`
	ToExchange   string        `json:"to_exchange"`
	Network      string        `json:"network"`
	CurrentStep  string        `json:"current_step"`
	Steps        []IpcStepInfo `json:"steps"`
	IsExecuting  bool          `json:"is_executing"`
	ErrorMessage string        `json:"error_message"`
}

type IpcStepInfo struct {
	Step        string  `json:"step"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
	StartedAt   *string `json:"started_at"`
	CompletedAt *string `json:"completed_at"`
}

func SnapshotToIpc(
	snap *models.AppSnapshot,
	scenarioConfig models.ScenarioConfig,
	transferState *IpcTransferState,
	transferJobs []IpcTransferJob,
	lastTickerAgeMs int64,
) *IpcSnapshot {
	ipcSnap := &IpcSnapshot{
		CoinStates:      make(map[string]*IpcCoinState),
		WalletStatus:    make(map[string]*IpcWalletStatus),
		KoreanNames:     make(map[string]string),
		Logs:            make([]IpcLogEntry, 0),
		OrderbookInfo:   make(map[string]*IpcOrderbookInfo),
		ScenarioThreads: make([]IpcLogThread, 0),
		UsdtKrw:         nil,
		UsdKrwForex:     nil,
		LastTickerAgeMs: lastTickerAgeMs,
		ScenarioConfig: IpcScenarioConfig{
			GapThresholdPercent:  scenarioConfig.GapThresholdPercent,
			DomesticGapThreshold: scenarioConfig.DomesticGapThreshold,
			FutBasisThreshold:    scenarioConfig.FutBasisThreshold,
		},
		Transfer:     transferState,
		TransferJobs: transferJobs,
	}

	if snap == nil {
		return ipcSnap
	}

	ipcSnap.UsdtKrw = snap.UsdtKrw
	ipcSnap.UsdKrwForex = snap.UsdKrwForex

	for symbol, state := range snap.CoinStates {
		if state == nil {
			ipcSnap.CoinStates[symbol] = nil
			continue
		}

		ipcSnap.CoinStates[symbol] = &IpcCoinState{
			Symbol:              state.Symbol,
			UpbitPrice:          decimalPtrToFloat64(state.UpbitPrice),
			BithumbPrice:        decimalPtrToFloat64(state.BithumbPrice),
			BinancePrice:        decimalPtrToFloat64(state.BinancePrice),
			BinanceKrw:          decimalPtrToFloat64(state.BinanceKrw),
			BybitPrice:          decimalPtrToFloat64(state.BybitPrice),
			BybitKrw:            decimalPtrToFloat64(state.BybitKrw),
			OkxPrice:            decimalPtrToFloat64(state.OkxPrice),
			OkxKrw:              decimalPtrToFloat64(state.OkxKrw),
			UpbitKimchi:         decimalPtrToFloat64(state.UpbitKimchi),
			BithumbKimchi:       decimalPtrToFloat64(state.BithumbKimchi),
			BybitKimchiUp:       decimalPtrToFloat64(state.BybitKimchiUp),
			BybitKimchiBt:       decimalPtrToFloat64(state.BybitKimchiBt),
			OkxKimchiUp:         decimalPtrToFloat64(state.OkxKimchiUp),
			OkxKimchiBt:         decimalPtrToFloat64(state.OkxKimchiBt),
			DomesticGap:         decimalPtrToFloat64(state.DomesticGap),
			BinanceFuturesPrice: decimalPtrToFloat64(state.BinanceFuturesPrice),
			BybitFuturesPrice:   decimalPtrToFloat64(state.BybitFuturesPrice),
			FuturesBasis:        decimalPtrToFloat64(state.FuturesBasis),
		}
	}

	for symbol, status := range snap.WalletStatus {
		if status == nil {
			ipcSnap.WalletStatus[symbol] = nil
			continue
		}

		ipcSnap.WalletStatus[symbol] = &IpcWalletStatus{
			Upbit:   exchangeWalletToIpc(status.Upbit),
			Bithumb: exchangeWalletToIpc(status.Bithumb),
			Binance: exchangeWalletToIpc(status.Binance),
			Bybit:   exchangeWalletToIpc(status.Bybit),
			Okx:     exchangeWalletToIpc(status.Okx),
		}
	}

	for k, v := range snap.KoreanNames {
		ipcSnap.KoreanNames[k] = v
	}

	ipcSnap.Logs = make([]IpcLogEntry, 0, len(snap.Logs))
	for _, entry := range snap.Logs {
		ipcSnap.Logs = append(ipcSnap.Logs, IpcLogEntry{
			Timestamp: entry.Timestamp.Format(time.RFC3339),
			Symbol:    entry.Symbol,
			Message:   entry.Message,
			LogType:   entry.LogType.String(),
		})
	}

	for symbol, info := range snap.OrderbookInfo {
		if info == nil {
			ipcSnap.OrderbookInfo[symbol] = nil
			continue
		}

		realKimpUp := make(map[string]float64, len(info.RealKimpUp))
		for exch, value := range info.RealKimpUp {
			realKimpUp[exch.ShortCode()] = value
		}

		realKimpBt := make(map[string]float64, len(info.RealKimpBt))
		for exch, value := range info.RealKimpBt {
			realKimpBt[exch.ShortCode()] = value
		}

		ipcSnap.OrderbookInfo[symbol] = &IpcOrderbookInfo{
			BtSpread:       info.BtSpread,
			UpSpread:       info.UpSpread,
			BtBuySlippage:  info.BtBuySlippage,
			UpSellSlippage: info.UpSellSlippage,
			RealGapBtUp:    info.RealGapBtUp,
			RealGapUpBt:    info.RealGapUpBt,
			SurfaceGap:     info.SurfaceGap,
			RealKimpUp:     realKimpUp,
			RealKimpBt:     realKimpBt,
		}
	}

	ipcSnap.ScenarioThreads = make([]IpcLogThread, 0, len(snap.ScenarioThreads))
	for _, thread := range snap.ScenarioThreads {
		var closedAt *string
		if thread.ClosedAt != nil {
			formatted := thread.ClosedAt.Format(time.RFC3339)
			closedAt = &formatted
		}

		subEntries := make([]IpcThreadEntry, 0, len(thread.SubEntries))
		for _, sub := range thread.SubEntries {
			subEntries = append(subEntries, IpcThreadEntry{
				Timestamp: sub.Timestamp.Format(time.RFC3339),
				Message:   sub.Message,
			})
		}

		ipcSnap.ScenarioThreads = append(ipcSnap.ScenarioThreads, IpcLogThread{
			ID:              thread.ID,
			Symbol:          thread.Symbol,
			Scenario:        thread.Scenario.String(),
			Key:             thread.Key,
			MainMessage:     thread.MainMessage,
			MainTimestamp:   thread.MainTimestamp.Format(time.RFC3339),
			SubEntries:      subEntries,
			IsActive:        thread.IsActive,
			ClosedAt:        closedAt,
			InitialValue:    thread.InitialValue,
			LastLoggedValue: thread.LastLoggedValue,
			Label:           thread.Scenario.Label(),
		})
	}

	return ipcSnap
}

func decimalPtrToFloat64(v *decimal.Decimal) *float64 {
	if v == nil {
		return nil
	}
	f := v.InexactFloat64()
	return &f
}

func exchangeWalletToIpc(wallet *models.ExchangeWalletStatus) *IpcExchangeWallet {
	if wallet == nil {
		return nil
	}
	depositBlockedChains := make([]string, len(wallet.DepositBlockedChains))
	copy(depositBlockedChains, wallet.DepositBlockedChains)
	withdrawBlockedChains := make([]string, len(wallet.WithdrawBlockedChains))
	copy(withdrawBlockedChains, wallet.WithdrawBlockedChains)

	return &IpcExchangeWallet{
		Deposit:               wallet.Deposit,
		Withdraw:              wallet.Withdraw,
		DepositBlockedChains:  depositBlockedChains,
		WithdrawBlockedChains: withdrawBlockedChains,
	}
}
