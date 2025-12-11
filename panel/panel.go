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
	mu                 sync.RWMutex
	batteries          []*battery.Battery
	ledStrip           *peripheral.ColorLedStrip
	airLocktButton     peripheral.ButtonReader
	batteryResetButton peripheral.ButtonReader
	batteryConnects    []peripheral.ButtonReader

	// LED allocation
	batteryLEDCount int // LEDs per battery section
	spacingLEDs     int // LEDs between batteries

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
	Batteries          []*battery.Battery
	LEDStrip           *peripheral.ColorLedStrip
	AirLockButton      peripheral.ButtonReader
	BatteryResetButton peripheral.ButtonReader
	BatteryConnects    []peripheral.ButtonReader
	UpdateRate         time.Duration // How often to update animations and check inputs
}

// NewPanel creates a new panel instance
func NewPanel(config PanelConfig) *Panel {
	if config.UpdateRate <= 0 {
		config.UpdateRate = 50 * time.Millisecond // 20 FPS default
	}

	// Calculate LED allocation
	totalLEDs := config.LEDStrip.NumLEDs()
	numBatteries := len(config.Batteries)
	spacingLEDs := 4
	totalSpacing := spacingLEDs * (numBatteries - 1)
	batteryLEDs := (totalLEDs - totalSpacing) / numBatteries

	p := &Panel{
		batteries:          config.Batteries,
		ledStrip:           config.LEDStrip,
		batteryResetButton: config.BatteryResetButton,
		batteryConnects:    config.BatteryConnects,
		airLocktButton:     config.AirLockButton,
		batteryLEDCount:    batteryLEDs,
		spacingLEDs:        spacingLEDs,
		stopAnimation:      make(chan struct{}),
		lastUpdate:         time.Now(),
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

	// Check inputs and update all batteries
	for i, bat := range p.batteries {
		if p.batteryResetButton.IsPressed() {
			bat.SetChargedOverride(true)
			neoPixel := peripheral.NeoPixel{}
			neoPixel.Configure()
			neoPixel.SetColorAndPause(Red, 50)
			continue
		}
		bat.SetIsDraining(p.batteryConnects[i].IsPressed())
	}

	// Update animation phases
	p.updateAnimationPhases(deltaTime)

	// Clear the strip first
	p.ledStrip.SetAll(Black)

	// Update LED display for each battery
	for i, bat := range p.batteries {
		info := bat.GetInfo()
		p.updateBatterySection(i, info)
	}

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

// getBatteryStartLED returns the starting LED index for a battery section
func (p *Panel) getBatteryStartLED(batteryIndex int) int {
	return batteryIndex * (p.batteryLEDCount + p.spacingLEDs)
}

// updateBatterySection updates the LED section for a specific battery
func (p *Panel) updateBatterySection(batteryIndex int, info battery.BatteryInfo) {
	startLED := p.getBatteryStartLED(batteryIndex)

	switch info.State {
	case battery.Charged:
		p.displayChargedSection(startLED)
	case battery.Disconnecting:
		p.displayDisconnectingSection(startLED, info.BatteryLevel)
	case battery.Draining:
		p.displayDrainingSection(startLED, info.BatteryLevel)
	case battery.Dead:
		p.displayDeadSection(startLED)
	case battery.Charging:
		p.displayChargingSection(startLED, info.BatteryLevel)
	default:
		p.displayUnknownSection(startLED)
	}
}

// displayChargedSection shows green LEDs for a battery section
func (p *Panel) displayChargedSection(startLED int) {
	// Pulse the gree with 1 second period
	// with a subtle pulse from 100% to 80%
	maxBrightness := uint8(40)
	pulseBrightness := uint8(float64(maxBrightness) * (0.9 + 0.1*math.Sin(p.flashPhase*2*math.Pi)))
	pulseColor := color.RGBA{R: 0, G: pulseBrightness, B: 0, A: 255}
	for i := 0; i < p.batteryLEDCount; i++ {
		p.ledStrip.SetPixel(startLED+i, pulseColor)
	}
}

// displayDisconnectingSection shows green flickering out with random pixels turning yellow or off
func (p *Panel) displayDisconnectingSection(startLED int, batteryLevel float32) {
	// Calculate how many pixels should be affected based on battery level
	pixelsAffected := int(math.Ceil(float64(p.batteryLEDCount) * float64(batteryLevel) / 100.0))
	if pixelsAffected < 0 {
		pixelsAffected = 0
	}
	if pixelsAffected > p.batteryLEDCount {
		pixelsAffected = p.batteryLEDCount
	}

	// Use flash phase to control the amount of flickering (more flickering over time)
	flickerIntensity := p.flashPhase // 0.0 to 1.0

	// Only flicker LEDs up to the battery level
	for i := 0; i < pixelsAffected; i++ {
		// Random chance for each pixel to flicker based on intensity
		if rand.Float64() < flickerIntensity*0.5 {
			// Randomly choose between yellow or off
			if rand.Float64() < 0.6 {
				p.ledStrip.SetPixel(startLED+i, Yellow)
			} else {
				p.ledStrip.SetPixel(startLED+i, Black)
			}
		} else {
			// Default to green when not flickering
			p.ledStrip.SetPixel(startLED+i, Green)
		}
	}
}

// displayDrainingSection shows yellow bar getting smaller with pixels incrementally flickering out
func (p *Panel) displayDrainingSection(startLED int, batteryLevel float32) {
	// Calculate how many pixels should be solidly lit based on battery level
	pixelsLit := int(math.Ceil(float64(p.batteryLEDCount) * float64(batteryLevel) / 100.0))
	if pixelsLit < 0 {
		pixelsLit = 0
	}
	if pixelsLit > p.batteryLEDCount {
		pixelsLit = p.batteryLEDCount
	}

	// Light up the solid yellow bar
	for i := 0; i < pixelsLit; i++ {
		p.ledStrip.SetPixel(startLED+i, Yellow)
	}

	// Add flickering effect at the edge of the bar to simulate pixels dying
	flickerZone := 2 // Number of pixels at the edge that can flicker
	for i := pixelsLit; i < pixelsLit+flickerZone && i < p.batteryLEDCount; i++ {
		// Random chance for edge pixels to flicker yellow
		if rand.Float64() < 0.3 {
			p.ledStrip.SetPixel(startLED+i, Yellow)
		}
	}
}

// displayDeadSection shows pulsing red with variable intensity for a battery section
func (p *Panel) displayDeadSection(startLED int) {
	// Pulse the red with 1 second period (same as draining)
	maxBrightness := uint8(10)
	pulseBrightness := uint8(float64(maxBrightness) * (0.5 + 0.5*math.Sin(p.flashPhase*2*math.Pi)))
	pulseColor := color.RGBA{R: pulseBrightness, G: 0, B: 0, A: 255}
	for i := 0; i < p.batteryLEDCount; i++ {
		p.ledStrip.SetPixel(startLED+i, pulseColor)
	}
}

// displayChargingSection shows a charging animation for a battery section
func (p *Panel) displayChargingSection(startLED int, batteryLevel float32) {
	// Show current charge level in green
	pixelsLit := int(math.Ceil(float64(p.batteryLEDCount) * float64(batteryLevel) / 100.0))
	if pixelsLit < 0 {
		pixelsLit = 0
	}
	if pixelsLit > p.batteryLEDCount {
		pixelsLit = p.batteryLEDCount
	}

	for i := 0; i < pixelsLit; i++ {
		p.ledStrip.SetPixel(startLED+i, Green)
	}

	// Add a moving "charging" indicator
	if pixelsLit < p.batteryLEDCount {
		// Create a yellow pulse that moves up the strip
		chargePos := int(p.flashPhase * float64(p.batteryLEDCount-pixelsLit))
		if chargePos < 0 {
			chargePos = 0
		}
		if chargePos+pixelsLit < p.batteryLEDCount {
			p.ledStrip.SetPixel(startLED+pixelsLit+chargePos, Yellow)
		}
	}
}

// displayUnknownSection shows a blue pattern to indicate unknown state for a battery section
func (p *Panel) displayUnknownSection(startLED int) {
	// Slow pulse in blue to indicate unknown/error state
	brightness := uint8(64 + 64*math.Sin(p.pulsePhase*2*math.Pi))
	unknownColor := color.RGBA{R: 0, G: 0, B: brightness, A: 255}
	for i := 0; i < p.batteryLEDCount; i++ {
		p.ledStrip.SetPixel(startLED+i, unknownColor)
	}
}

// GetBatteryInfo returns current battery information for a specific battery
func (p *Panel) GetBatteryInfo(batteryIndex int) battery.BatteryInfo {
	if batteryIndex < 0 || batteryIndex >= len(p.batteries) {
		// Return empty info for invalid index
		return battery.BatteryInfo{}
	}
	return p.batteries[batteryIndex].GetInfo()
}

// GetAllBatteryInfo returns current battery information for all batteries
func (p *Panel) GetAllBatteryInfo() []battery.BatteryInfo {
	infos := make([]battery.BatteryInfo, len(p.batteries))
	for i, bat := range p.batteries {
		infos[i] = bat.GetInfo()
	}
	return infos
}
