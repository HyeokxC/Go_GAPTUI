package models

import "time"

type LogType int

const (
	WalletBlocked LogType = iota
	WalletUnblocked
	SessionStart
)

func (l LogType) String() string {
	switch l {
	case WalletBlocked:
		return "wallet_blocked"
	case WalletUnblocked:
		return "wallet_unblocked"
	case SessionStart:
		return "session_start"
	default:
		return "session_start"
	}
}

func LogTypeFromString(s string) LogType {
	switch s {
	case "wallet_blocked":
		return WalletBlocked
	case "wallet_unblocked":
		return WalletUnblocked
	case "session_start":
		return SessionStart
	default:
		return SessionStart
	}
}

type LogEntry struct {
	Timestamp time.Time
	Symbol    string
	Message   string
	LogType   LogType
}
