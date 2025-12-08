package panel

import (
	"machine"
	"sync"
)

// InputHandler represents digital input interface
type InputHandler interface {
	// IsPressed returns true if the input is currently pressed/active
	IsPressed() bool
}

// PinInputHandler handles digital input from a hardware pin
type PinInputHandler struct {
	pin      machine.Pin
	inverted bool // true if pin reads low when pressed
	mu       sync.RWMutex
}

// NewPinInputHandler creates a new hardware pin input handler
func NewPinInputHandler(pin machine.Pin, inverted bool) *PinInputHandler {
	return &PinInputHandler{
		pin:      pin,
		inverted: inverted,
	}
}

// Configure sets up the pin as input with pull-up resistor
func (p *PinInputHandler) Configure() error {
	p.pin.Configure(machine.PinConfig{
		Mode: machine.PinInputPullup,
	})
	return nil
}

// IsPressed returns true if the input is currently pressed/active
func (p *PinInputHandler) IsPressed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	reading := p.pin.Get()
	if p.inverted {
		return !reading
	}
	return reading
}

// MockInputHandler is a simple implementation for testing
type MockInputHandler struct {
	pressed bool
	mu      sync.RWMutex
}

// NewMockInputHandler creates a new mock input handler
func NewMockInputHandler() *MockInputHandler {
	return &MockInputHandler{}
}

// IsPressed returns the current pressed state
func (m *MockInputHandler) IsPressed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pressed
}

// SetPressed sets the pressed state (for testing)
func (m *MockInputHandler) SetPressed(pressed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pressed = pressed
}
