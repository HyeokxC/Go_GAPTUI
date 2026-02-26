package models

import "time"

type ScenarioType int

const (
	GapThreshold ScenarioType = iota
	DomesticGap
	FutBasis
)

func (s ScenarioType) String() string {
	switch s {
	case GapThreshold:
		return "gap_threshold"
	case DomesticGap:
		return "domestic_gap"
	case FutBasis:
		return "fut_basis"
	default:
		return "unknown"
	}
}

func (s ScenarioType) Label() string {
	switch s {
	case GapThreshold:
		return "KIMP"
	case DomesticGap:
		return "DOM-GAP"
	case FutBasis:
		return "FUT%"
	default:
		return ""
	}
}

func AllScenarioTypes() []ScenarioType {
	return []ScenarioType{GapThreshold, DomesticGap, FutBasis}
}

type ThreadEntry struct {
	Timestamp time.Time
	Message   string
}

type LogThread struct {
	ID              uint64
	Symbol          string
	Scenario        ScenarioType
	Key             string
	MainMessage     string
	MainTimestamp   time.Time
	SubEntries      []ThreadEntry
	IsActive        bool
	ClosedAt        *time.Time
	InitialValue    float64
	LastLoggedValue float64
}

type ScenarioConfig struct {
	GapThresholdPercent  float64
	GapSubThreadChange   float64
	DomesticGapThreshold float64
	FutBasisThreshold    float64
}

func DefaultScenarioConfig() ScenarioConfig {
	return ScenarioConfig{
		GapThresholdPercent:  5.0,
		GapSubThreadChange:   3.0,
		DomesticGapThreshold: 1.5,
		FutBasisThreshold:    0.5,
	}
}
