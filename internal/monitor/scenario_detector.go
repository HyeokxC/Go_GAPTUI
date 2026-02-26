package monitor

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/hyeokx/Go_GAPTUI/internal/models"
)

type ScenarioEvent struct {
	ThreadID    uint64
	IsNewThread bool
}

type ScenarioDetector struct {
	mu               sync.Mutex
	config           models.ScenarioConfig
	threads          []models.LogThread
	nextID           uint64
	wasAboveGap      map[string]bool
	wasAboveDomestic map[string][2]bool
	wasAboveFutBasis map[string][2]bool
}

func NewScenarioDetector(config models.ScenarioConfig) *ScenarioDetector {
	return &ScenarioDetector{
		config:           config,
		threads:          make([]models.LogThread, 0),
		nextID:           1,
		wasAboveGap:      make(map[string]bool),
		wasAboveDomestic: make(map[string][2]bool),
		wasAboveFutBasis: make(map[string][2]bool),
	}
}

func (d *ScenarioDetector) Threads() []models.LogThread {
	d.mu.Lock()
	defer d.mu.Unlock()

	out := make([]models.LogThread, len(d.threads))
	for i := range d.threads {
		out[i] = d.threads[i]
		if d.threads[i].SubEntries != nil {
			out[i].SubEntries = append([]models.ThreadEntry(nil), d.threads[i].SubEntries...)
		}
	}
	return out
}

func (d *ScenarioDetector) SetGapThreshold(v float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.GapThresholdPercent = v
}

func (d *ScenarioDetector) SetDomesticGapThreshold(v float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.DomesticGapThreshold = v
}

func (d *ScenarioDetector) SetFutBasisThreshold(v float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config.FutBasisThreshold = v
}

func (d *ScenarioDetector) Refresh() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for i := range d.threads {
		if d.threads[i].IsActive {
			d.threads[i].IsActive = false
			closedAt := now
			d.threads[i].ClosedAt = &closedAt
		}
	}

	d.wasAboveGap = make(map[string]bool)
	d.wasAboveDomestic = make(map[string][2]bool)
	d.wasAboveFutBasis = make(map[string][2]bool)
}

func (d *ScenarioDetector) OnPriceUpdate(
	symbol string,
	kimpByExchange map[models.Exchange]float64,
	domesticGap float64,
	futuresBasis float64,
) []ScenarioEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	events := make([]ScenarioEvent, 0)

	for exchange, value := range kimpByExchange {
		overseas := exchange.ShortCode()
		if overseas == "" {
			continue
		}

		keys := [2]string{fmt.Sprintf("UP-%s", overseas), fmt.Sprintf("BT-%s", overseas)}
		for _, key := range keys {
			compositeKey := symbol + ":" + key
			threshold := d.config.GapThresholdPercent
			isAbove := math.Abs(value) >= threshold
			wasAbove := d.wasAboveGap[compositeKey]

			if isAbove && !wasAbove {
				threadID, isNew := d.getOrReuseThread(
					symbol,
					models.GapThreshold,
					key,
					fmt.Sprintf("%s %s KIMP %.2f%% (threshold: %.2f%%)", symbol, key, value, threshold),
					value,
				)
				events = append(events, ScenarioEvent{ThreadID: threadID, IsNewThread: isNew})
				d.wasAboveGap[compositeKey] = true
				continue
			}

			if isAbove && wasAbove {
				for i := range d.threads {
					thread := &d.threads[i]
					if thread.Symbol != symbol || thread.Scenario != models.GapThreshold || thread.Key != key || !thread.IsActive {
						continue
					}
					delta := value - thread.LastLoggedValue
					if math.Abs(delta) >= d.config.GapSubThreadChange {
						d.addSubEntry(thread.ID, fmt.Sprintf("%.2f%% (Δ%+.2f%%p)", value, delta), value)
					}
					break
				}
				continue
			}

			if !isAbove && wasAbove {
				for i := range d.threads {
					thread := &d.threads[i]
					if thread.Symbol == symbol && thread.Scenario == models.GapThreshold && thread.Key == key && thread.IsActive {
						d.closeThread(thread.ID, fmt.Sprintf("Closed: %.2f%%", value))
						break
					}
				}
				d.wasAboveGap[compositeKey] = false
			}
		}
	}

	domesticTrackKey := symbol + ":domestic"
	prevDomestic := d.wasAboveDomestic[domesticTrackKey]
	domThreshold := d.config.DomesticGapThreshold

	domCases := []struct {
		key             string
		idx             int
		active          bool
		value           float64
		label           string
		subThreadChange float64
	}{
		{key: "pos", idx: 0, active: domesticGap >= domThreshold, value: domesticGap, label: "UP>BT", subThreadChange: 0.5},
		{key: "neg", idx: 1, active: -domesticGap >= domThreshold, value: domesticGap, label: "BT>UP", subThreadChange: 0.5},
	}

	for _, c := range domCases {
		wasAbove := prevDomestic[c.idx]
		if c.active && !wasAbove {
			threadID, isNew := d.getOrReuseThread(
				symbol,
				models.DomesticGap,
				c.key,
				fmt.Sprintf("%s DOM-GAP %s %.2f%% (threshold: %.2f%%)", symbol, c.label, c.value, domThreshold),
				c.value,
			)
			events = append(events, ScenarioEvent{ThreadID: threadID, IsNewThread: isNew})
			prevDomestic[c.idx] = true
			continue
		}

		if c.active && wasAbove {
			for i := range d.threads {
				thread := &d.threads[i]
				if thread.Symbol != symbol || thread.Scenario != models.DomesticGap || thread.Key != c.key || !thread.IsActive {
					continue
				}
				delta := c.value - thread.LastLoggedValue
				if math.Abs(delta) >= c.subThreadChange {
					d.addSubEntry(thread.ID, fmt.Sprintf("%.2f%% (Δ%+.2f%%p)", c.value, delta), c.value)
				}
				break
			}
			continue
		}

		if !c.active && wasAbove {
			for i := range d.threads {
				thread := &d.threads[i]
				if thread.Symbol == symbol && thread.Scenario == models.DomesticGap && thread.Key == c.key && thread.IsActive {
					d.closeThread(thread.ID, fmt.Sprintf("Closed: %.2f%%", c.value))
					break
				}
			}
			prevDomestic[c.idx] = false
		}
	}
	d.wasAboveDomestic[domesticTrackKey] = prevDomestic

	futTrackKey := symbol + ":futbasis"
	prevFut := d.wasAboveFutBasis[futTrackKey]
	futThreshold := d.config.FutBasisThreshold

	futCases := []struct {
		key             string
		idx             int
		active          bool
		value           float64
		label           string
		subThreadChange float64
	}{
		{key: "pos", idx: 0, active: futuresBasis >= futThreshold, value: futuresBasis, label: "POS", subThreadChange: 0.1},
		{key: "neg", idx: 1, active: -futuresBasis >= futThreshold, value: futuresBasis, label: "NEG", subThreadChange: 0.1},
	}

	for _, c := range futCases {
		wasAbove := prevFut[c.idx]
		if c.active && !wasAbove {
			threadID, isNew := d.getOrReuseThread(
				symbol,
				models.FutBasis,
				c.key,
				fmt.Sprintf("%s FUT%% %s %.2f%% (threshold: %.2f%%)", symbol, c.label, c.value, futThreshold),
				c.value,
			)
			events = append(events, ScenarioEvent{ThreadID: threadID, IsNewThread: isNew})
			prevFut[c.idx] = true
			continue
		}

		if c.active && wasAbove {
			for i := range d.threads {
				thread := &d.threads[i]
				if thread.Symbol != symbol || thread.Scenario != models.FutBasis || thread.Key != c.key || !thread.IsActive {
					continue
				}
				delta := c.value - thread.LastLoggedValue
				if math.Abs(delta) >= c.subThreadChange {
					d.addSubEntry(thread.ID, fmt.Sprintf("%.2f%% (Δ%+.2f%%p)", c.value, delta), c.value)
				}
				break
			}
			continue
		}

		if !c.active && wasAbove {
			for i := range d.threads {
				thread := &d.threads[i]
				if thread.Symbol == symbol && thread.Scenario == models.FutBasis && thread.Key == c.key && thread.IsActive {
					d.closeThread(thread.ID, fmt.Sprintf("Closed: %.2f%%", c.value))
					break
				}
			}
			prevFut[c.idx] = false
		}
	}
	d.wasAboveFutBasis[futTrackKey] = prevFut

	return events
}

func (d *ScenarioDetector) getOrReuseThread(symbol string, scenario models.ScenarioType, key string, message string, value float64) (uint64, bool) {
	for i := range d.threads {
		thread := &d.threads[i]
		if thread.Symbol == symbol && thread.Scenario == scenario && thread.Key == key && thread.IsActive {
			return thread.ID, false
		}
	}

	if len(d.threads) >= 200 {
		for i := range d.threads {
			if !d.threads[i].IsActive {
				d.threads = append(d.threads[:i], d.threads[i+1:]...)
				break
			}
		}
	}

	id := d.nextID
	d.nextID++
	now := time.Now()
	d.threads = append(d.threads, models.LogThread{
		ID:              id,
		Symbol:          symbol,
		Scenario:        scenario,
		Key:             key,
		MainMessage:     message,
		MainTimestamp:   now,
		SubEntries:      make([]models.ThreadEntry, 0),
		IsActive:        true,
		ClosedAt:        nil,
		InitialValue:    value,
		LastLoggedValue: value,
	})

	return id, true
}

func (d *ScenarioDetector) addSubEntry(threadID uint64, message string, value float64) {
	for i := range d.threads {
		thread := &d.threads[i]
		if thread.ID != threadID {
			continue
		}
		thread.SubEntries = append(thread.SubEntries, models.ThreadEntry{Timestamp: time.Now(), Message: message})
		thread.LastLoggedValue = value
		return
	}
}

func (d *ScenarioDetector) closeThread(threadID uint64, message string) {
	for i := range d.threads {
		thread := &d.threads[i]
		if thread.ID != threadID {
			continue
		}
		now := time.Now()
		thread.IsActive = false
		thread.ClosedAt = &now
		thread.SubEntries = append(thread.SubEntries, models.ThreadEntry{Timestamp: now, Message: message})
		return
	}
}
