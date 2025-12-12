package peripheral

import (
	"context"
	"machine"
	"sync"
	"time"
)

type BoardYellowLight struct {
	Led     machine.Pin
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex
}

func (e *BoardYellowLight) Configure() {
	// Blink yellow board LED
	e.Led = machine.PC30
	e.Led.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

func (e *BoardYellowLight) StartBlink() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return // Already running
	}

	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.running = true

	go func() {
		defer func() {
			e.mu.Lock()
			e.running = false
			e.mu.Unlock()
			// Turn off LED when stopping
			e.Led.Low()
		}()

		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		ledState := false

		for {
			select {
			case <-e.ctx.Done():
				return
			case <-ticker.C:
				if ledState {
					e.Led.High()
				} else {
					e.Led.Low()
				}
				ledState = !ledState
			}
		}
	}()
}

func (e *BoardYellowLight) StopBlink() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return // Not running
	}

	if e.cancel != nil {
		e.cancel()
	}
}

func (e *BoardYellowLight) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}
