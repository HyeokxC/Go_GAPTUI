package models

type AppSnapshot struct {
	CoinStates      map[string]*CoinState
	WalletStatus    map[string]*CoinWalletStatus
	KoreanNames     map[string]string
	Logs            []LogEntry
	OrderbookInfo   map[string]*OrderbookInfo
	ScenarioThreads []LogThread
	UsdtKrw         *float64
	UsdKrwForex     *float64
}

func NewEmptySnapshot() *AppSnapshot {
	return &AppSnapshot{
		CoinStates:      make(map[string]*CoinState),
		WalletStatus:    make(map[string]*CoinWalletStatus),
		KoreanNames:     make(map[string]string),
		Logs:            make([]LogEntry, 0),
		OrderbookInfo:   make(map[string]*OrderbookInfo),
		ScenarioThreads: make([]LogThread, 0),
	}
}
