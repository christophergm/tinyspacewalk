package battery

import (
	"sync"
	"time"
)

// SystemState represents the five possible battery system states
type SystemState int

const (
	Charged SystemState = iota
	Disconnecting
	Draining
	Dead
	Charging
)

// String returns a string representation of the SystemState
func (s SystemState) String() string {
	switch s {
	case Charged:
		return "Charged"
	case Disconnecting:
		return "Disconnecting"
	case Draining:
		return "Draining"
	case Dead:
		return "Dead"
	case Charging:
		return "Charging"
	default:
		return "Unknown"
	}
}

// BatteryInfo holds all current battery system properties
type BatteryInfo struct {
	State                          SystemState
	BatteryLevel                   float32
	ChargedOverride                bool
	IsDraining                     bool
	DrainRate                      time.Duration
	ChargeRate                     time.Duration
	DisconnectingDuration          time.Duration
	LastUpdateAt                   time.Time
	DisconnectingDurationRemaining time.Duration // Only valid when in Disconnecting state
}

// Config holds configuration parameters for battery creation
type Config struct {
	DrainRate             time.Duration // time to fully drain from 100% to 0%
	ChargeRate            time.Duration // time to fully charge from 0% to 100%
	DisconnectingDuration time.Duration // time to stay in disconnecting state
}

// DefaultBatteryConfig returns a configuration with sensible defaults
func DefaultBatteryConfig() Config {
	return Config{
		DrainRate:             60 * time.Minute, // Default: 60 minutes to fully drain
		ChargeRate:            30 * time.Minute, // Default: 30 minutes to fully charge
		DisconnectingDuration: 30 * time.Second, // Default: 30 seconds in disconnecting state
	}
}

// FastBatteryConfig returns a configuration optimized for demonstrations and testing
func FastBatteryConfig() Config {
	return Config{
		DrainRate:             2 * 60 * time.Second, // 2 minutes to fully drain
		ChargeRate:            30 * time.Second,     // 4 minutes to fully charge
		DisconnectingDuration: 1 * time.Second,      // 1 second in disconnecting state
	}
}

// SlowBatteryConfig returns a configuration for realistic long-term simulation
func StandardBatteryConfig() Config {
	return Config{
		DrainRate:             200 * time.Minute, // 3.33 hours to fully drain
		ChargeRate:            100 * time.Minute, // 100 minutes to fully charge
		DisconnectingDuration: 1 * time.Minute,   // 1 minute in disconnecting state
	}
}

// Battery represents a battery with state machine based on three inputs and time
type Battery struct {
	mu                    sync.RWMutex
	state                 SystemState
	batteryLevel          float32       // 0-100 percentage
	chargedOverride       bool          // Input 1
	isDraining            bool          // Input 2
	drainRate             time.Duration // Input 3: time to fully drain
	chargeRate            time.Duration // time to fully charge
	disconnectingDuration time.Duration // time to stay in disconnecting state

	// State timing
	lastUpdateAt           time.Time
	disconnectingStartTime time.Time

	// Ticker for updates
	ticker     *time.Ticker
	stopTicker chan struct{}
	running    bool
}

// NewBattery creates a new battery instance with the specified configuration
func NewBattery(config Config) *Battery {
	// Validate and set minimums
	if config.DrainRate <= 0 {
		config.DrainRate = time.Minute
	}
	if config.ChargeRate <= 0 {
		config.ChargeRate = time.Minute
	}
	if config.DisconnectingDuration < 0 {
		config.DisconnectingDuration = 0
	}

	b := &Battery{
		state:                 Charged,
		batteryLevel:          100,
		chargedOverride:       false,
		isDraining:            false,
		drainRate:             config.DrainRate,
		chargeRate:            config.ChargeRate,
		disconnectingDuration: config.DisconnectingDuration,
		lastUpdateAt:          time.Now(),
		stopTicker:            make(chan struct{}),
	}
	b.startTicker()
	return b
}

// SetChargedOverride sets the charged override input
// When true, battery level is set to 100 and state transitions to Charged
func (b *Battery) SetChargedOverride(override bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.chargedOverride = override
	if override {
		b.batteryLevel = 100.0
		b.setState(Charged)
	}
	// If turning off override, let the state machine determine next state on next tick
}

// SetIsDraining sets the draining input
func (b *Battery) SetIsDraining(draining bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.isDraining = draining
}

// Stop stops the battery's internal ticker and operations
func (b *Battery) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		close(b.stopTicker)
		if b.ticker != nil {
			b.ticker.Stop()
		}
		b.running = false
	}
}

// setState sets the internal state and updates timing (must be called with mutex locked)
func (b *Battery) setState(newState SystemState) {
	if b.state != newState {
		b.state = newState
		b.lastUpdateAt = time.Now()

		// Special handling for disconnecting state
		if newState == Disconnecting {
			b.disconnectingStartTime = time.Now()
		}
	}
}

// startTicker begins the internal state machine ticker
func (b *Battery) startTicker() {
	if b.running {
		return
	}

	b.running = true
	b.ticker = time.NewTicker(100 * time.Millisecond) // Update every 100ms

	go func() {
		for {
			select {
			case <-b.ticker.C:
				b.updateStateMachine()
			case <-b.stopTicker:
				return
			}
		}
	}()
}

// updateStateMachine implements the state machine logic
func (b *Battery) updateStateMachine() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	deltaMinutes := now.Sub(b.lastUpdateAt).Minutes()

	// Rule 1: ChargedOverride always forces Charged state with 100% battery
	if b.chargedOverride {
		b.batteryLevel = 100.0
		b.setState(Charged)
		return
	}

	// State machine transitions based on current state
	switch b.state {
	case Charged:
		// if in ChargedState and IsDraining == true transition to Disconnecting
		if b.isDraining {
			b.setState(Disconnecting)
		}

	case Disconnecting:
		// if in Disconnecting, after DisconnectingDuration has elapsed then transition to Draining
		disconnectingElapsed := now.Sub(b.disconnectingStartTime)
		if disconnectingElapsed >= b.disconnectingDuration {
			b.setState(Draining)
		}

	case Draining:
		// if in Draining, then reduce BatteryLevel by drainRate
		drainPercentPerMinute := 100.0 / b.drainRate.Minutes()
		drainAmount := drainPercentPerMinute * deltaMinutes
		newLevel := float64(b.batteryLevel) - drainAmount

		if newLevel <= 0 {
			// if in Draining and BatteryLevel reaches 0 then transition to Dead
			b.batteryLevel = 0.0
			b.setState(Dead)
		} else {
			b.batteryLevel = float32(newLevel)

			// if in Draining and isDraining is set to false then transition to Charging
			if !b.isDraining {
				b.setState(Charging)
			}
		}

	case Dead:
		// Dead state - can only exit via ChargedOverride or if isDraining becomes false
		if !b.isDraining {
			b.setState(Charging)
		}

	case Charging:
		// if in Charging, then increment battery level by charge rate
		chargePercentPerMinute := 100.0 / b.chargeRate.Minutes()
		chargeAmount := chargePercentPerMinute * deltaMinutes
		newLevel := float64(b.batteryLevel) + chargeAmount

		if newLevel >= 100 {
			// if in Charging and battery level reaches 100 then transition to Charged
			b.batteryLevel = 100.0
			b.setState(Charged)
		} else {
			b.batteryLevel = float32(newLevel)

			// If isDraining becomes true while charging, transition to Disconnecting
			if b.isDraining {
				b.setState(Disconnecting)
			}
		}
	}

	b.lastUpdateAt = now
}

// GetInfo returns a comprehensive summary of the current battery state
func (b *Battery) GetInfo() BatteryInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	info := BatteryInfo{
		State:                 b.state,
		BatteryLevel:          b.batteryLevel,
		ChargedOverride:       b.chargedOverride,
		IsDraining:            b.isDraining,
		DrainRate:             b.drainRate,
		ChargeRate:            b.chargeRate,
		DisconnectingDuration: b.disconnectingDuration,
		LastUpdateAt:          b.lastUpdateAt,
	}

	// Add state-specific information
	if b.state == Disconnecting {
		elapsed := time.Since(b.disconnectingStartTime)
		remaining := b.disconnectingDuration - elapsed
		remaining = max(remaining, 0)
		info.DisconnectingDurationRemaining = remaining
	}

	return info
}
