package background

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/hyeokx/Go_GAPTUI/internal/config"
	"github.com/hyeokx/Go_GAPTUI/internal/db"
	"github.com/hyeokx/Go_GAPTUI/internal/exchanges"
	"github.com/hyeokx/Go_GAPTUI/internal/ipc"
	"github.com/hyeokx/Go_GAPTUI/internal/models"
	"github.com/hyeokx/Go_GAPTUI/internal/monitor"
	"github.com/hyeokx/Go_GAPTUI/internal/transfer"
)

type Runner struct {
	cfg       *config.AppConfig
	monitor   *monitor.KimchiMonitor
	detector  *monitor.ScenarioDetector
	obManager *exchanges.OrderbookManager
	executor  *transfer.TransferExecutor
	ipcServer *ipc.IpcServer
	db        *sql.DB

	snapshot     atomic.Pointer[models.AppSnapshot]
	walletStatus sync.Map
	koreanNames  sync.Map
	usdtKRW      atomic.Value
	usdKRWForex  atomic.Value

	lastTickerTime atomic.Value
}

func NewRunner(cfg *config.AppConfig) (*Runner, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	database, err := db.InitDatabase("kimchi.db")
	if err != nil {
		return nil, err
	}

	if err := db.AddSessionStartLog(database); err != nil {
		_ = database.Close()
		return nil, err
	}

	loadedLogs, err := db.LoadLogs(database, 1000)
	if err != nil {
		_ = database.Close()
		return nil, err
	}

	r := &Runner{
		cfg:       cfg,
		monitor:   monitor.NewKimchiMonitor(),
		detector:  monitor.NewScenarioDetector(models.DefaultScenarioConfig()),
		obManager: exchanges.NewOrderbookManager(),
		executor:  transfer.NewTransferExecutor(),
		db:        database,
	}
	r.ipcServer = ipc.NewIpcServer(r.handleCommand)

	r.usdtKRW.Store(0.0)
	r.usdKRWForex.Store(0.0)
	r.lastTickerTime.Store(time.Time{})

	initial := models.NewEmptySnapshot()
	initial.Logs = append(initial.Logs, loadedLogs...)
	r.snapshot.Store(initial)

	return r, nil
}

func (r *Runner) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer func() {
		if r.db != nil {
			_ = r.db.Close()
		}
	}()

	var symbolsMu sync.RWMutex
	currentSymbols := append([]string(nil), r.cfg.Symbols...)
	currentKoreanNames := make(map[string]string)

	setSymbols := func(symbols []string, names map[string]string) {
		symbolsMu.Lock()
		currentSymbols = append([]string(nil), symbols...)
		currentKoreanNames = make(map[string]string, len(symbols))
		for _, symbol := range symbols {
			currentKoreanNames[symbol] = strings.TrimSpace(names[symbol])
		}
		symbolsMu.Unlock()

		for _, symbol := range symbols {
			name := strings.TrimSpace(names[symbol])
			r.koreanNames.Store(symbol, name)
		}
	}

	getSymbols := func() []string {
		symbolsMu.RLock()
		defer symbolsMu.RUnlock()
		return append([]string(nil), currentSymbols...)
	}

	getKoreanNames := func() map[string]string {
		symbolsMu.RLock()
		defer symbolsMu.RUnlock()
		copied := make(map[string]string, len(currentKoreanNames))
		for symbol, name := range currentKoreanNames {
			copied[symbol] = name
		}
		return copied
	}

	discoverSymbols := func() {
		client := exchanges.BuildSimpleHTTPClient(r.cfg.Exchange.RequestTimeoutSecs)
		symbols, names, err := exchanges.FetchSymbols(runCtx, client)
		if err != nil {
			log.Warn().Err(err).Msg("symbol discovery failed")
			return
		}
		if len(symbols) == 0 {
			return
		}
		setSymbols(symbols, names)
		log.Info().Int("symbols", len(symbols)).Msg("symbol discovery updated")
	}

	discoverSymbols()

	startupSymbols := getSymbols()
	tickerSymbols := append([]string(nil), startupSymbols...)
	usdtSymbol := strings.ToUpper(strings.TrimSpace(r.cfg.UsdtSymbol))
	if usdtSymbol == "" {
		usdtSymbol = "USDT"
	}
	if !containsSymbol(tickerSymbols, usdtSymbol) {
		tickerSymbols = append(tickerSymbols, usdtSymbol)
	}

	exchanges.RunUpbitTicker(runCtx, tickerSymbols, r.onTicker)
	exchanges.RunBithumbTicker(runCtx, tickerSymbols, r.onTicker)
	exchanges.RunBinanceTicker(runCtx, usdtSymbol, r.onTicker)
	exchanges.RunBybitTicker(runCtx, startupSymbols, r.onTicker)
	exchanges.RunOKXTicker(runCtx, startupSymbols, r.onTicker)
	exchanges.RunBinanceFuturesTicker(runCtx, usdtSymbol, r.onTicker)
	exchanges.RunBybitFuturesTicker(runCtx, startupSymbols, r.onTicker)

	exchanges.RunUpbitOrderbook(runCtx, startupSymbols, r.onOrderbook)
	exchanges.RunBithumbOrderbook(runCtx, startupSymbols, r.obManager, r.onOrderbook)
	exchanges.RunBinanceOrderbook(runCtx, startupSymbols, usdtSymbol, r.onOrderbook)
	exchanges.RunBybitOrderbook(runCtx, startupSymbols, r.obManager, r.onOrderbook)
	exchanges.RunOKXOrderbook(runCtx, startupSymbols, r.onOrderbook)

	var orderbookMu sync.RWMutex
	orderbookInfo := make(map[string]*models.OrderbookInfo)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	spawn := func(fn func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}

	spawn(func() {
		if !r.cfg.AutoDiscoverSymbols {
			return
		}
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				discoverSymbols()
			}
		}
	})

	spawn(func() {
		client := exchanges.BuildSimpleHTTPClient(r.cfg.Exchange.RequestTimeoutSecs)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				usdKrw, err := fetchUSDKRWForex(runCtx, client)
				if err != nil {
					continue
				}
				r.usdKRWForex.Store(usdKrw)

				usdtKrw, err := fetchUSDTKRW(runCtx, client, usdKrw)
				if err != nil || usdtKrw <= 0 {
					usdtKrw = usdKrw
				}
				r.usdtKRW.Store(usdtKrw)
				r.monitor.SetUsdtKRW(usdtKrw)
			}
		}
	})

	spawn(func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				log.Info().Msg("wallet status check")
			}
		}
	})

	spawn(func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				usdtKrw := r.loadUSDTKRW()
				updated := make(map[string]*models.OrderbookInfo)
				for _, symbol := range getSymbols() {
					info := exchanges.ProcessOrderbookInfo(symbol, r.obManager, usdtKrw, 10_000_000)
					if info != nil {
						updated[symbol] = info
					}
				}
				orderbookMu.Lock()
				for symbol, info := range updated {
					orderbookInfo[symbol] = info
				}
				orderbookMu.Unlock()
			}
		}
	})

	spawn(func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				orderbookMu.RLock()
				orderbookSnapshot := make(map[string]*models.OrderbookInfo, len(orderbookInfo))
				for symbol, info := range orderbookInfo {
					orderbookSnapshot[symbol] = info
				}
				orderbookMu.RUnlock()

				snap := &models.AppSnapshot{
					CoinStates:      r.monitor.GetStates(),
					WalletStatus:    map[string]*models.CoinWalletStatus{},
					KoreanNames:     getKoreanNames(),
					Logs:            snapshotLogs(r.snapshot.Load()),
					OrderbookInfo:   orderbookSnapshot,
					ScenarioThreads: r.detector.Threads(),
					UsdtKrw:         floatPtrIfPositive(r.loadUSDTKRW()),
					UsdKrwForex:     floatPtrIfPositive(r.loadUSDKRWForex()),
				}

				r.snapshot.Store(snap)

				ipcSnap := ipc.SnapshotToIpc(
					snap,
					models.DefaultScenarioConfig(),
					nil,
					transferJobsToIpc(r.executor.GetJobs()),
					r.lastTickerAgeMs(),
				)
				r.ipcServer.Broadcast(ipcSnap)
			}
		}
	})

	spawn(func() {
		if err := r.ipcServer.Start(runCtx, "127.0.0.1:9876"); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	})

	select {
	case <-runCtx.Done():
		cancel()
		wg.Wait()
		return nil
	case err := <-errCh:
		cancel()
		wg.Wait()
		return err
	}
}

func (r *Runner) onTicker(data models.TickerData) {
	r.lastTickerTime.Store(time.Now())

	if strings.EqualFold(strings.TrimSpace(data.Symbol), strings.TrimSpace(r.cfg.UsdtSymbol)) {
		usdtKrw := data.Price.InexactFloat64()
		if usdtKrw > 0 {
			r.usdtKRW.Store(usdtKrw)
			r.monitor.SetUsdtKRW(usdtKrw)
		}
	}

	state := r.monitor.OnTicker(data)
	if state == nil {
		return
	}

	kimp := map[models.Exchange]float64{
		models.Binance: decimalToFloat(state.UpbitKimchi),
		models.Bybit:   decimalToFloat(state.BybitKimchiUp),
		models.Okx:     decimalToFloat(state.OkxKimchiUp),
	}

	r.detector.OnPriceUpdate(
		state.Symbol,
		kimp,
		decimalToFloat(state.DomesticGap),
		decimalToFloat(state.FuturesBasis),
	)
}

func (r *Runner) onOrderbook(exchange models.Exchange, symbol string, asks, bids []models.OrderbookEntry) {
	r.obManager.Update(exchange, symbol, asks, bids)
}

func (r *Runner) handleCommand(cmd ipc.IpcCommand) {
	switch strings.ToLower(strings.TrimSpace(cmd.Type)) {
	case "set_scenario_threshold":
		var params struct {
			GapThresholdPercent  *float64 `json:"gap_threshold_percent"`
			DomesticGapThreshold *float64 `json:"domestic_gap_threshold"`
			FutBasisThreshold    *float64 `json:"fut_basis_threshold"`
			Name                 string   `json:"name"`
			Value                *float64 `json:"value"`
		}
		if err := json.Unmarshal(cmd.Params, &params); err != nil {
			log.Warn().Err(err).Msg("failed to parse set_scenario_threshold params")
			return
		}

		if params.GapThresholdPercent != nil {
			r.detector.SetGapThreshold(*params.GapThresholdPercent)
		}
		if params.DomesticGapThreshold != nil {
			r.detector.SetDomesticGapThreshold(*params.DomesticGapThreshold)
		}
		if params.FutBasisThreshold != nil {
			r.detector.SetFutBasisThreshold(*params.FutBasisThreshold)
		}

		if params.Value != nil {
			switch strings.ToLower(strings.TrimSpace(params.Name)) {
			case "gap", "gap_threshold", "gap_threshold_percent":
				r.detector.SetGapThreshold(*params.Value)
			case "domestic", "domestic_gap", "domestic_gap_threshold":
				r.detector.SetDomesticGapThreshold(*params.Value)
			case "fut", "fut_basis", "fut_basis_threshold":
				r.detector.SetFutBasisThreshold(*params.Value)
			}
		}
	default:
		log.Info().Msgf("unhandled command: %s", cmd.Type)
	}
}

func (r *Runner) loadUSDTKRW() float64 {
	value, ok := r.usdtKRW.Load().(float64)
	if !ok {
		return 0
	}
	return value
}

func (r *Runner) loadUSDKRWForex() float64 {
	value, ok := r.usdKRWForex.Load().(float64)
	if !ok {
		return 0
	}
	return value
}

func (r *Runner) lastTickerAgeMs() int64 {
	value, ok := r.lastTickerTime.Load().(time.Time)
	if !ok || value.IsZero() {
		return -1
	}
	return time.Since(value).Milliseconds()
}

func snapshotLogs(existing *models.AppSnapshot) []models.LogEntry {
	if existing == nil {
		return []models.LogEntry{}
	}
	out := make([]models.LogEntry, len(existing.Logs))
	copy(out, existing.Logs)
	return out
}

func transferJobsToIpc(jobs []*transfer.TransferJob) []ipc.IpcTransferJob {
	result := make([]ipc.IpcTransferJob, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}

		steps := make([]ipc.IpcStepInfo, 0, len(job.Steps))
		for _, step := range job.Steps {
			steps = append(steps, ipc.IpcStepInfo{
				Step:        step.Step.String(),
				Status:      transferStepStatusString(step.Status),
				Message:     step.Message,
				StartedAt:   timeToRFC3339(step.StartedAt),
				CompletedAt: timeToRFC3339(step.CompletedAt),
			})
		}

		result = append(result, ipc.IpcTransferJob{
			ID:           job.ID,
			Coin:         job.Coin,
			Amount:       job.Amount,
			FromExchange: job.FromExchange.String(),
			ToExchange:   job.ToExchange.String(),
			Network:      job.NetworkDisplay,
			CurrentStep:  job.CurrentStep.String(),
			Steps:        steps,
			IsExecuting:  job.IsExecuting,
			ErrorMessage: job.ErrorMessage,
		})
	}
	return result
}

func transferStepStatusString(status transfer.StepStatus) string {
	switch status {
	case transfer.StatusPending:
		return "pending"
	case transfer.StatusInProgress:
		return "in_progress"
	case transfer.StatusCompleted:
		return "completed"
	case transfer.StatusFailed:
		return "failed"
	default:
		return "pending"
	}
}

func timeToRFC3339(t *time.Time) *string {
	if t == nil {
		return nil
	}
	v := t.Format(time.RFC3339)
	return &v
}

func decimalToFloat(value *decimal.Decimal) float64 {
	if value == nil {
		return 0
	}
	return value.InexactFloat64()
}

func floatPtrIfPositive(v float64) *float64 {
	if v <= 0 {
		return nil
	}
	out := v
	return &out
}

func containsSymbol(symbols []string, target string) bool {
	t := strings.ToUpper(strings.TrimSpace(target))
	for _, symbol := range symbols {
		if strings.ToUpper(strings.TrimSpace(symbol)) == t {
			return true
		}
	}
	return false
}

func fetchUSDKRWForex(ctx context.Context, client *http.Client) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.manana.kr/exchange/rate.json?base=KRW&code=USD", nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, errors.New("unexpected forex status")
	}

	var payload []struct {
		Name string  `json:"name"`
		Rate float64 `json:"rate"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	for _, entry := range payload {
		if entry.Name == "USDKRW=X" && entry.Rate > 0 {
			return entry.Rate, nil
		}
	}

	return 0, errors.New("usdkrw not found")
}

func fetchUSDTKRW(ctx context.Context, client *http.Client, usdKrw float64) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.binance.com/api/v3/ticker/price?symbol=USDTUSDT", nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, errors.New("unexpected binance status")
	}

	var payload struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	price, err := strconv.ParseFloat(strings.TrimSpace(payload.Price), 64)
	if err != nil || price <= 0 {
		return 0, errors.New("invalid usdt ticker")
	}

	if usdKrw <= 0 {
		return 0, errors.New("invalid usdkrw fallback")
	}

	return price * usdKrw, nil
}
