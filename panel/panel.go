package panel

import (
	"image/color"
	"math"
	"math/rand"
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
		p.displayDisconnecting(info.BatteryLevel)
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

// displayDisconnecting shows green flickering out with random pixels turning yellow or off
func (p *Panel) displayDisconnecting(batteryLevel float32) {
	length := p.ledStrip.NumLEDs()
	if length == 0 {
		return
	}

	// Calculate how many pixels should be affected based on battery level
	pixelsAffected := int(math.Ceil(float64(length) * float64(batteryLevel) / 100.0))
	if pixelsAffected < 0 {
		pixelsAffected = 0
	}
	if pixelsAffected > length {
		pixelsAffected = length
	}

	// Start with all LEDs off
	p.ledStrip.SetAll(Black)

	// Use flash phase to control the amount of flickering (more flickering over time)
	flickerIntensity := p.flashPhase // 0.0 to 1.0

	// Only flicker LEDs up to the battery level
	for i := 0; i < pixelsAffected; i++ {
		// Random chance for each pixel to flicker based on intensity
		if rand.Float64() < flickerIntensity*0.5 {
			// Randomly choose between yellow or off
			if rand.Float64() < 0.6 {
				p.ledStrip.SetPixel(i, Yellow)
			} else {
				p.ledStrip.SetPixel(i, Black)
			}
		} else {
			// Default to green when not flickering
			p.ledStrip.SetPixel(i, Green)
		}
	}
}

// displayDraining shows yellow bar getting smaller with pixels incrementally flickering out
func (p *Panel) displayDraining(batteryLevel float32) {
	length := p.ledStrip.NumLEDs()
	if length == 0 {
		return
	}

	// Clear all pixels first
	p.ledStrip.SetAll(Black)

	// Calculate how many pixels should be solidly lit based on battery level
	pixelsLit := int(math.Ceil(float64(length) * float64(batteryLevel) / 100.0))
	if pixelsLit < 0 {
		pixelsLit = 0
	}
	if pixelsLit > length {
		pixelsLit = length
	}

	// Light up the solid yellow bar
	for i := 0; i < pixelsLit; i++ {
		p.ledStrip.SetPixel(i, Yellow)
	}

	// Add flickering effect at the edge of the bar to simulate pixels dying
	flickerZone := 3 // Number of pixels at the edge that can flicker
	for i := pixelsLit; i < pixelsLit+flickerZone && i < length; i++ {
		// Random chance for edge pixels to flicker yellow
		if rand.Float64() < 0.3 {
			p.ledStrip.SetPixel(i, Yellow)
		}
	}
}

// displayDead shows pulsing red with variable intensity
func (p *Panel) displayDead() {
	// Pulse the red with 1 second period (same as draining)
	pulseBrightness := uint8(127 + 127*math.Sin(p.flashPhase*2*math.Pi))
	pulseColor := color.RGBA{R: pulseBrightness, G: 0, B: 0, A: 255}
	p.ledStrip.SetAll(pulseColor)
}

// displayCharging shows a charging animation
func (p *Panel) displayCharging(batteryLevel float32) {
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
