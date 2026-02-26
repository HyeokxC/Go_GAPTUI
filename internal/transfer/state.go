package transfer

import (
	"time"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

type TransferStep int

const (
	StepIdle TransferStep = iota
	StepWithdrawalRequested
	StepWithdrawalProcessing
	StepWithdrawalConfirmed
	StepDepositProcessing
	StepDepositConfirmed
	StepCompleted
)

func (s TransferStep) String() string {
	switch s {
	case StepIdle:
		return "idle"
	case StepWithdrawalRequested:
		return "withdrawal_requested"
	case StepWithdrawalProcessing:
		return "withdrawal_processing"
	case StepWithdrawalConfirmed:
		return "withdrawal_confirmed"
	case StepDepositProcessing:
		return "deposit_processing"
	case StepDepositConfirmed:
		return "deposit_confirmed"
	case StepCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

type StepStatus int

const (
	StatusPending StepStatus = iota
	StatusInProgress
	StatusCompleted
	StatusFailed
)

type StepInfo struct {
	Step        TransferStep
	Status      StepStatus
	Message     string
	StartedAt   *time.Time
	CompletedAt *time.Time
}

type NetworkInfo struct {
	Network         string
	ToNetwork       string
	DisplayName     string
	DepositEnabled  bool
	WithdrawEnabled bool
	WithdrawFee     *float64
	WithdrawMin     *float64
	NeedsMemo       bool
}

type BalanceInfo struct {
	Available float64
	Locked    float64
}

type DepositAddressInfo struct {
	Address string
	Tag     string
}

type TransferLogEntry struct {
	Timestamp time.Time
	Message   string
}

type TransferJob struct {
	ID                    uint64
	Coin                  string
	Amount                float64
	FromExchange          models.Exchange
	ToExchange            models.Exchange
	Network               string
	NetworkDisplay        string
	Address               string
	Memo                  string
	CurrentStep           TransferStep
	Steps                 []StepInfo
	WithdrawalID          string
	TxHash                string
	TransferNetwork       string
	IsExecuting           bool
	IsCancelled           bool
	ErrorMessage          string
	StartedAt             time.Time
	Logs                  []TransferLogEntry
	AutoBuyBeforeTransfer bool
	MarketBuyResult       string
	AutoSellOnArrival     bool
	MarketSellResult      string
	WithdrawFee           float64
}

type TransferState struct {
	SelectedCoin          string
	FromExchange          models.Exchange
	ToExchange            models.Exchange
	FromNetworks          []NetworkInfo
	ToNetworks            []NetworkInfo
	AvailableNetworks     []NetworkInfo
	SelectedNetworkIdx    *int
	Amount                string
	Balance               *BalanceInfo
	DepositInfo           *DepositAddressInfo
	DepositAddress        string
	DepositTag            string
	ToIsPersonalWallet    bool
	PersonalWalletAddress string
	PersonalWalletTag     string
	AutoBuyBeforeTransfer bool
	AutoSellOnArrival     bool
	MarketOrderPending    bool
	MarketOrderResult     string
	ShowConfirmation      bool
	Logs                  []TransferLogEntry
}
