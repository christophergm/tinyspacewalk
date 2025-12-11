package peripheral

import (
	"machine"
	"sync"
)

// ButtonReader represents digital input interface
type ButtonReader interface {
	// IsPressed returns true if the input is currently pressed/active
	IsPressed() bool
}

// Compile-time assertion that PinInputHandler implements InputHandler
var _ ButtonReader = (*Button)(nil)

// Button handles digital input from a hardware pin
type Button struct {
	pin      machine.Pin
	inverted bool // true if pin reads low when pressed
}

// NewButton creates a new hardware pin input handler
func NewButton(pin machine.Pin, inverted bool) *Button {
	return &Button{
		pin:      pin,
		inverted: inverted,
	}
}

// Configure sets up the pin as input with pull-up resistor
func (p *Button) Configure() error {
	p.pin.Configure(machine.PinConfig{
		Mode: machine.PinInputPullup,
	})
	return nil
}

// IsPressed returns true if the input is currently pressed/active
// The battery number parameter is ignored for hardware pins since
// each pin represents input for all batteries connected to it
func (p *Button) IsPressed() bool {
	reading := p.pin.Get()
	if p.inverted {
		return !reading
	}
	return reading
}

var _ ButtonReader = (*MockButton)(nil)

// MockButton is a simple implementation for testing
type MockButton struct {
	pressed bool // map of battery number to pressed state
	mu      sync.RWMutex
}

// NewMockButton creates a new mock input handler
func NewMockButton() *MockButton {
	return &MockButton{
		pressed: false,
	}
}

// IsPressed returns the pressed state for a specific battery
func (m *MockButton) IsPressed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pressed
}

// SetPressed sets the pressed state for a specific battery (for testing)
func (m *MockButton) SetPressed(pressed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pressed = pressed
}
