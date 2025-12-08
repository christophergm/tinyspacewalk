package panel

import (
	"image/color"
	"math"
	"sync"
	"time"

	"github.com/christophergm/tinyspacewalk/battery"
	"github.com/christophergm/tinyspacewalk/peripheral"
)

// Common colors
var (
	Black  = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	Red    = color.RGBA{R: 5, G: 0, B: 0, A: 255}
	Green  = color.RGBA{R: 0, G: 5, B: 0, A: 255}
	Yellow = color.RGBA{R: 5, G: 5, B: 0, A: 255}
	Blue   = color.RGBA{R: 0, G: 0, B: 5, A: 255}
	White  = color.RGBA{R: 5, G: 5, B: 5, A: 255}
)

// Panel manages the LED display and input handling for the battery system
type Panel struct {
	mu                sync.RWMutex
	battery           *battery.Battery
	ledStrip          *peripheral.ColorLedStrip
	chargedOverrideIn InputHandler
	drainingIn        InputHandler

	// Animation state
	animationTicker *time.Ticker
	stopAnimation   chan struct{}
	running         bool

	// Flash/pulse timing
	flashPhase float64 // 0.0 to 1.0 for flash animations
	pulsePhase float64 // 0.0 to 1.0 for pulse animations
	lastUpdate time.Time
}

// PanelConfig holds configuration for panel creation
type PanelConfig struct {
	Battery           *battery.Battery
	LEDStrip          *peripheral.ColorLedStrip
	ChargedOverrideIn InputHandler
	DrainingIn        InputHandler
	UpdateRate        time.Duration // How often to update animations and check inputs
}

// NewPanel creates a new panel instance
func NewPanel(config PanelConfig) *Panel {
	if config.UpdateRate <= 0 {
		config.UpdateRate = 50 * time.Millisecond // 20 FPS default
	}

	p := &Panel{
		battery:           config.Battery,
		ledStrip:          config.LEDStrip,
		chargedOverrideIn: config.ChargedOverrideIn,
		drainingIn:        config.DrainingIn,
		stopAnimation:     make(chan struct{}),
		lastUpdate:        time.Now(),
	}

	p.start(config.UpdateRate)
	return p
}

// Start begins the panel's update loop
func (p *Panel) start(updateRate time.Duration) {
	if p.running {
		return
	}

	p.running = true
	p.animationTicker = time.NewTicker(updateRate)

	go func() {
		for {
			select {
			case <-p.animationTicker.C:
				p.update()
			case <-p.stopAnimation:
				return
			}
		}
	}()
}

// Stop stops the panel's update loop
func (p *Panel) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		close(p.stopAnimation)
		if p.animationTicker != nil {
			p.animationTicker.Stop()
		}
		p.running = false
		p.ledStrip.Clear()
		p.ledStrip.Show()
	}
}

// update handles input checking, animation updates, and LED display
func (p *Panel) update() {
	now := time.Now()
	deltaTime := now.Sub(p.lastUpdate).Seconds()
	p.lastUpdate = now

	// Check inputs and update battery
	if p.chargedOverrideIn != nil {
		p.battery.SetChargedOverride(p.chargedOverrideIn.IsPressed())
	}
	if p.drainingIn != nil {
		p.battery.SetIsDraining(p.drainingIn.IsPressed())
	}

	// Update animation phases
	p.updateAnimationPhases(deltaTime)

	// Update LED display based on battery state
	info := p.battery.GetInfo()
	p.updateLEDDisplay(info)

	// Show the updated display
	p.ledStrip.Show()
}

// updateAnimationPhases updates the timing for flash and pulse animations
func (p *Panel) updateAnimationPhases(deltaTime float64) {
	// Flash phase: completes a cycle every 1 second
	p.flashPhase += deltaTime
	if p.flashPhase >= 1.0 {
		p.flashPhase -= 1.0
	}

	// Pulse phase: completes a cycle every 2 seconds (slower pulse)
	p.pulsePhase += deltaTime * 0.5
	if p.pulsePhase >= 1.0 {
		p.pulsePhase -= 1.0
	}
}

// updateLEDDisplay updates the LED strip based on battery state
func (p *Panel) updateLEDDisplay(info battery.BatteryInfo) {
	switch info.State {
	case battery.Charged:
		p.displayCharged()
	case battery.Disconnecting:
		p.displayDisconnecting()
	case battery.Draining:
		p.displayDraining(info.BatteryLevel)
	case battery.Dead:
		p.displayDead()
	case battery.Charging:
		p.displayCharging(info.BatteryLevel)
	default:
		p.displayUnknown()
	}
}

// displayCharged shows all green LEDs
func (p *Panel) displayCharged() {
	p.ledStrip.SetAll(Green)
}

// displayDisconnecting shows flashing yellow
func (p *Panel) displayDisconnecting() {
	// Flash on/off every 0.5 seconds
	if p.flashPhase < 0.5 {
		p.ledStrip.SetAll(Yellow)
	} else {
		p.ledStrip.SetAll(Black)
	}
}

// displayDraining shows red bar with decreasing lights and top two pixels pulsing
func (p *Panel) displayDraining(batteryLevel int) {
	length := p.ledStrip.NumLEDs()
	if length == 0 {
		return
	}

	// Clear all pixels first
	p.ledStrip.SetAll(Black)

	// Calculate how many pixels should be lit based on battery level
	pixelsLit := int(math.Ceil(float64(length-2) * float64(batteryLevel) / 100.0))
	if pixelsLit < 0 {
		pixelsLit = 0
	}
	if pixelsLit > length-2 {
		pixelsLit = length - 2
	}

	// Light up the bottom pixels as a red bar
	for i := 0; i < pixelsLit; i++ {
		p.ledStrip.SetPixel(i, Red)
	}

	// Pulse the top two pixels
	pulseBrightness := uint8(127 + 127*math.Sin(p.pulsePhase*2*math.Pi))
	pulseColor := color.RGBA{R: pulseBrightness, G: 0, B: 0, A: 255}

	if length >= 2 {
		p.ledStrip.SetPixel(length-2, pulseColor)
		p.ledStrip.SetPixel(length-1, pulseColor)
	} else if length >= 1 {
		p.ledStrip.SetPixel(length-1, pulseColor)
	}
}

// displayDead shows blinking red
func (p *Panel) displayDead() {
	// Blink on/off every 0.25 seconds (faster than disconnecting)
	if math.Mod(p.flashPhase*2, 1.0) < 0.5 {
		p.ledStrip.SetAll(Red)
	} else {
		p.ledStrip.SetAll(Black)
	}
}

// displayCharging shows a charging animation
func (p *Panel) displayCharging(batteryLevel int) {
	length := p.ledStrip.NumLEDs()
	if length == 0 {
		return
	}

	// Clear all pixels first
	p.ledStrip.SetAll(Black)

	// Show current charge level in green
	pixelsLit := int(math.Ceil(float64(length) * float64(batteryLevel) / 100.0))
	if pixelsLit < 0 {
		pixelsLit = 0
	}
	if pixelsLit > length {
		pixelsLit = length
	}

	for i := 0; i < pixelsLit; i++ {
		p.ledStrip.SetPixel(i, Green)
	}

	// Add a moving "charging" indicator
	if pixelsLit < length {
		// Create a yellow pulse that moves up the strip
		chargePos := int(p.flashPhase * float64(length-pixelsLit))
		if chargePos < 0 {
			chargePos = 0
		}
		if chargePos+pixelsLit < length {
			p.ledStrip.SetPixel(pixelsLit+chargePos, Yellow)
		}
	}
}

// displayUnknown shows a blue pattern to indicate unknown state
func (p *Panel) displayUnknown() {
	// Slow pulse in blue to indicate unknown/error state
	brightness := uint8(64 + 64*math.Sin(p.pulsePhase*2*math.Pi))
	unknownColor := color.RGBA{R: 0, G: 0, B: brightness, A: 255}
	p.ledStrip.SetAll(unknownColor)
}

// GetBatteryInfo returns current battery information
func (p *Panel) GetBatteryInfo() battery.BatteryInfo {
	return p.battery.GetInfo()
}
