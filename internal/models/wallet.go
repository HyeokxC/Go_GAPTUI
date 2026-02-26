package models

type ExchangeWalletStatus struct {
	Deposit               bool
	Withdraw              bool
	DepositBlockedChains  []string
	WithdrawBlockedChains []string
}

type CoinWalletStatus struct {
	Upbit   *ExchangeWalletStatus
	Bithumb *ExchangeWalletStatus
	Binance *ExchangeWalletStatus
	Bybit   *ExchangeWalletStatus
	Okx     *ExchangeWalletStatus
}
