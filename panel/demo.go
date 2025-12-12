package panel

import (
	"math/rand"
	"time"

	"github.com/christophergm/tinyspacewalk/peripheral"
)

// DemoAllBatteries runs a demo sequence of inputs to mock
// battery connect inputs with context support for graceful shutdown
func (p *Panel) DemoAllBatteries(mockBatteryConnects []*peripheral.MockButton, neoPixel peripheral.NeoPixel) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			// Clean up: turn off all mock buttons
			for _, button := range mockBatteryConnects {
				button.SetPressed(false)
			}
			return
		case <-ticker.C:
			// Demo sequence - simulate input presses
			// Set draining to true for all batteries (simulate button press)
			for _, button := range mockBatteryConnects {
				select {
				case <-p.ctx.Done():
					return
				default:
					button.SetPressed(true)
					// Use a shorter sleep with context checking
					if !p.sleepWithContext(1 * time.Second) {
						return
					}
				}
			}

			// Wait longer period with context checking
			if !p.sleepWithContext(10 * time.Second) {
				return
			}

			// Turn off all batteries
			for _, button := range mockBatteryConnects {
				select {
				case <-p.ctx.Done():
					return
				default:
					button.SetPressed(false)
				}
			}
		}
	}
}

// DemoRandomBatteries randomly toggles individual inputs to
// mock battery connect inputs with context support for graceful shutdown
func (p *Panel) DemoRandomBatteries(batteryResetButton *peripheral.MockButton, mockBatteryConnects []*peripheral.MockButton, neoPixel peripheral.NeoPixel) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	defer func() {
		// Clean up: turn off all mock buttons
		batteryResetButton.SetPressed(false)
		for _, button := range mockBatteryConnects {
			button.SetPressed(false)
		}
	}()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			// Pick a random battery (0-4)
			batteryNum := rand.Intn(5)
			// Pick a random action (true/false)
			pressed := rand.Float32() < 0.5
			// Apply to a random input handler
			mockBatteryConnects[batteryNum].SetPressed(pressed)

			// Wait a bit more with context checking
			if !p.sleepWithContext(1 * time.Second) {
				return
			}

			// Always set first battery to true (as in original)
			mockBatteryConnects[0].SetPressed(true)
		}
	}
}

// sleepWithContext sleeps for the given duration while respecting context cancellation
// Returns true if sleep completed normally, false if context was cancelled
func (p *Panel) sleepWithContext(duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-p.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// StartDemoAllBatteries starts the DemoAllBatteries routine in a goroutine
// This is a convenience method that checks if we're using mock buttons
func (p *Panel) StartDemoAllBatteries(neoPixel peripheral.NeoPixel) {
	// Try to cast batteryConnects to mock buttons
	mockButtons := make([]*peripheral.MockButton, 0, len(p.batteryConnects))
	for _, buttonReader := range p.batteryConnects {
		if mockButton, ok := buttonReader.(*peripheral.MockButton); ok {
			mockButtons = append(mockButtons, mockButton)
		}
	}

	// Only start demo if all buttons are mock buttons
	if len(mockButtons) == len(p.batteryConnects) {
		go p.DemoAllBatteries(mockButtons, neoPixel)
	}
}

// StartDemoRandomBatteries starts the DemoRandomBatteries routine in a goroutine
// This is a convenience method that checks if we're using mock buttons
func (p *Panel) StartDemoRandomBatteries(neoPixel peripheral.NeoPixel) {
	// Try to cast batteryConnects to mock buttons
	mockButtons := make([]*peripheral.MockButton, 0, len(p.batteryConnects))
	for _, buttonReader := range p.batteryConnects {
		if mockButton, ok := buttonReader.(*peripheral.MockButton); ok {
			mockButtons = append(mockButtons, mockButton)
		}
	}

	// Try to cast battery reset button to mock button
	if mockResetButton, ok := p.batteryResetButton.(*peripheral.MockButton); ok {
		// Only start demo if all buttons are mock buttons
		if len(mockButtons) == len(p.batteryConnects) {
			go p.DemoRandomBatteries(mockResetButton, mockButtons, neoPixel)
		}
	}
}
