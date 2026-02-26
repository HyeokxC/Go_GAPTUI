package transfer

import (
	"sync"
	"time"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

type TransferExecutor struct {
	mu     sync.Mutex
	jobs   []*TransferJob
	nextID uint64
}

type TransferJobParams struct {
	Coin           string
	Amount         float64
	FromExchange   models.Exchange
	ToExchange     models.Exchange
	Network        string
	NetworkDisplay string
	Address        string
	Memo           string
	AutoBuy        bool
	AutoSell       bool
	WithdrawFee    float64
}

func NewTransferExecutor() *TransferExecutor {
	return &TransferExecutor{
		jobs:   make([]*TransferJob, 0),
		nextID: 1,
	}
}

func (e *TransferExecutor) CreateJob(params TransferJobParams) *TransferJob {
	e.mu.Lock()
	defer e.mu.Unlock()

	job := &TransferJob{
		ID:                    e.nextID,
		Coin:                  params.Coin,
		Amount:                params.Amount,
		FromExchange:          params.FromExchange,
		ToExchange:            params.ToExchange,
		Network:               params.Network,
		NetworkDisplay:        params.NetworkDisplay,
		Address:               params.Address,
		Memo:                  params.Memo,
		CurrentStep:           StepIdle,
		Steps:                 initialStepInfos(),
		TransferNetwork:       params.NetworkDisplay,
		IsExecuting:           false,
		IsCancelled:           false,
		StartedAt:             time.Now(),
		AutoBuyBeforeTransfer: params.AutoBuy,
		AutoSellOnArrival:     params.AutoSell,
		WithdrawFee:           params.WithdrawFee,
	}

	e.jobs = append(e.jobs, job)
	e.nextID++

	return job
}

func (e *TransferExecutor) GetJobs() []*TransferJob {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]*TransferJob, len(e.jobs))
	copy(result, e.jobs)
	return result
}

func (e *TransferExecutor) CancelJob(jobID uint64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, job := range e.jobs {
		if job.ID == jobID {
			job.IsCancelled = true
			return true
		}
	}

	return false
}

func initialStepInfos() []StepInfo {
	steps := []TransferStep{
		StepWithdrawalRequested,
		StepWithdrawalProcessing,
		StepWithdrawalConfirmed,
		StepDepositProcessing,
		StepDepositConfirmed,
		StepCompleted,
	}

	result := make([]StepInfo, 0, len(steps))
	for _, step := range steps {
		result = append(result, StepInfo{
			Step:    step,
			Status:  StatusPending,
			Message: "",
		})
	}

	return result
}
