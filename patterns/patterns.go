package patterns

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/christophergm/tinyspacewalk/battery"
	"github.com/christophergm/tinyspacewalk/peripheral"
)

// Pattern represents a LED pattern that can be started and stopped
type Pattern interface {
	Start(strip *peripheral.ColorLedStrip, done <-chan struct{}) error
	Name() string
}

// PanelStatus represents the status of a solar panel for battery pattern
type PanelStatus struct {
	Status int
}

// BatteryPattern displays a battery-like visualization using Battery struct
type BatteryPattern struct {
	BackgroundColor  color.RGBA
	PanelGapPixels   int
	PanelWidthPixels int
	Battery          *battery.Battery
	DelayScale       int
}

// NewBatteryPattern creates a new battery pattern with default values
func NewBatteryPattern() *BatteryPattern {
	config := battery.Config{
		DrainRate:         10.0, // 10% per minute
		ChargeRate:        5.0,  // 5% per minute
		DisconnectingTime: 0.5,  // 30 seconds
	}
	bat := battery.NewBattery(config)
	bat.SetIsDraining(true) // Start draining

	return &BatteryPattern{
		BackgroundColor:  color.RGBA{R: 50, G: 50, B: 0, A: 255},
		PanelGapPixels:   3,
		PanelWidthPixels: 20,
		Battery:          bat,
		DelayScale:       500,
	}
}

func (p *BatteryPattern) Name() string {
	return "Battery"
}

func (p *BatteryPattern) Start(strip *peripheral.ColorLedStrip, done <-chan struct{}) error {
	ticker := time.NewTicker(time.Duration(peripheral.ReadAnalogInputAsDelay(p.DelayScale)) * time.Millisecond)
	defer ticker.Stop()

	numPanels := 5

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			// Clear buffer with background color
			strip.SetAll(p.BackgroundColor)

			// Get current battery info
			batteryInfo := p.Battery.GetInfo()

			// Calculate how many panels to show as "active" based on battery level
			activePanels := (batteryInfo.BatteryLevel * numPanels) / 100

			// Choose panel color based on battery state
			var panelColor color.RGBA
			switch batteryInfo.State {
			case battery.Charged:
				panelColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
			case battery.Charging:
				panelColor = color.RGBA{R: 0, G: 100, B: 255, A: 255} // Blue
			case battery.Draining:
				if batteryInfo.BatteryLevel > 50 {
					panelColor = color.RGBA{R: 0, G: 200, B: 0, A: 255} // Green
				} else if batteryInfo.BatteryLevel > 20 {
					panelColor = color.RGBA{R: 255, G: 200, B: 0, A: 255} // Orange
				} else {
					panelColor = color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
				}
			case battery.Disconnecting:
				panelColor = color.RGBA{R: 100, G: 0, B: 0, A: 255} // Dark red
			case battery.Dead:
				panelColor = color.RGBA{R: 50, G: 0, B: 0, A: 255} // Very dark red
			default:
				panelColor = color.RGBA{R: 50, G: 50, B: 50, A: 255} // Gray
			}

			// Draw panels
			for i := 0; i < numPanels; i++ {
				var currentPanelColor color.RGBA
				if i < activePanels {
					currentPanelColor = panelColor
				} else {
					// Dim color for inactive panels
					currentPanelColor = color.RGBA{
						R: panelColor.R / 4,
						G: panelColor.G / 4,
						B: panelColor.B / 4,
						A: 255,
					}
				}

				for j := 0; j < p.PanelWidthPixels; j++ {
					pos := (j + i*p.PanelWidthPixels + i*p.PanelGapPixels) % strip.NumLEDs()
					strip.SetPixel(pos, currentPanelColor)
				}
			}

			// Add charging override indicator
			if batteryInfo.ChargedOverride {
				overrideColor := color.RGBA{R: 255, G: 0, B: 255, A: 255} // Magenta
				for i := strip.NumLEDs() - 3; i < strip.NumLEDs(); i++ {
					strip.SetPixel(i, overrideColor)
				}
			}

			strip.Show()
			ticker.Reset(time.Duration(peripheral.ReadAnalogInputAsDelay(p.DelayScale)) * time.Millisecond)
		}
	}
}

// SpinPattern creates a spinning tail effect with twinkling background
type SpinPattern struct {
	TailColors    []color.RGBA
	TwinkleColor  color.RGBA
	TwinkleChance int
	DelayScale    int
	position      int
	tailLength    int
}

// NewSpinPattern creates a new spin pattern with default values
func NewSpinPattern() *SpinPattern {
	return &SpinPattern{
		TailColors: []color.RGBA{
			{R: 50, G: 50, B: 0, A: 255},  // Tail segment 1
			{R: 90, G: 10, B: 50, A: 255}, // Tail segment 2
			{R: 10, G: 0, B: 100, A: 255}, // Tail segment 3
		},
		TwinkleColor:  color.RGBA{R: 0, G: 50, B: 0, A: 255},
		TwinkleChance: 8, // Out of 10
		DelayScale:    500,
		tailLength:    rand.Intn(100) + 1,
	}
}

func (p *SpinPattern) Name() string {
	return "Spin"
}

func (p *SpinPattern) Start(strip *peripheral.ColorLedStrip, done <-chan struct{}) error {
	ticker := time.NewTicker(time.Duration(peripheral.ReadSliderInputScaled(p.DelayScale)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			for i := 0; i < strip.NumLEDs(); i++ {
				var col color.RGBA

				if i < p.tailLength/3 {
					col = p.TailColors[0]
				} else if i < p.tailLength/3*2 {
					col = p.TailColors[1]
				} else if i < p.tailLength {
					col = p.TailColors[2]
				} else {
					// Background with occasional twinkle
					if rand.Intn(10) > p.TwinkleChance {
						col = p.TwinkleColor
					} else {
						col = color.RGBA{R: 0, G: 0, B: 0, A: 255}
					}
				}

				pos := (i + p.position) % strip.NumLEDs()
				strip.SetPixel(pos, col)
			}

			strip.Show()
			p.position++
			ticker.Reset(time.Duration(peripheral.ReadSliderInputScaled(p.DelayScale)) * time.Millisecond)
		}
	}
}

// TwinklePattern creates a twinkling star effect
type TwinklePattern struct {
	BackgroundColor color.RGBA
	TwinkleColor    color.RGBA
	TwinkleChance   int // Percentage chance (0-100)
	DelayScale      int
}

// NewTwinklePattern creates a new twinkle pattern with default values
func NewTwinklePattern() *TwinklePattern {
	return &TwinklePattern{
		BackgroundColor: color.RGBA{R: 10, G: 10, B: 0, A: 255},
		TwinkleColor:    color.RGBA{R: 0, G: 70, B: 0, A: 255},
		TwinkleChance:   20,
		DelayScale:      1000,
	}
}

func (p *TwinklePattern) Name() string {
	return "Twinkle"
}

func (p *TwinklePattern) Start(strip *peripheral.ColorLedStrip, done <-chan struct{}) error {
	ticker := time.NewTicker(time.Duration(peripheral.ReadSliderInputScaled(p.DelayScale)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			for i := 0; i < strip.NumLEDs(); i++ {
				if rand.Intn(100) < p.TwinkleChance {
					strip.SetPixel(i, p.TwinkleColor)
				} else {
					strip.SetPixel(i, p.BackgroundColor)
				}
			}

			strip.Show()
			ticker.Reset(time.Duration(peripheral.ReadSliderInputScaled(p.DelayScale)) * time.Millisecond)
		}
	}
}

// ExplodePattern creates an explosion effect radiating from a center point
type ExplodePattern struct {
	CenterPosition int
	MaxMagnitude   int
	Iterations     int
	IterationDelay time.Duration
}

// NewExplodePattern creates a new explode pattern with default values
func NewExplodePattern(centerPosition int) *ExplodePattern {
	return &ExplodePattern{
		CenterPosition: centerPosition,
		MaxMagnitude:   10,
		Iterations:     10,
		IterationDelay: 20 * time.Millisecond,
	}
}

func (p *ExplodePattern) Name() string {
	return "Explode"
}

func (p *ExplodePattern) Start(strip *peripheral.ColorLedStrip, done <-chan struct{}) error {
	for j := 0; j < p.Iterations; j++ {
		select {
		case <-done:
			return nil
		default:
			for i := 0; i < strip.NumLEDs(); i++ {
				distance := (p.CenterPosition - i) % strip.NumLEDs()
				magnitude := 3 * (strip.NumLEDs() - distance) / strip.NumLEDs()
				magnitude = magnitude + rand.Intn(9) - j

				if magnitude < 0 {
					magnitude = 0
				}

				col := color.RGBA{
					R: uint8(3 * magnitude),
					G: uint8(2 * magnitude),
					B: uint8(magnitude),
					A: 255,
				}
				strip.SetPixel(i, col)
			}

			strip.Show()
			time.Sleep(p.IterationDelay)
		}
	}
	return nil
}

// PatternManager manages multiple patterns and provides control functionality
type PatternManager struct {
	strip          *peripheral.ColorLedStrip
	currentPattern Pattern
	stopChan       chan struct{}
	running        bool
}

// NewPatternManager creates a new pattern manager
func NewPatternManager(strip *peripheral.ColorLedStrip) *PatternManager {
	return &PatternManager{
		strip: strip,
	}
}

// StartPattern starts a new pattern, stopping any currently running pattern
func (pm *PatternManager) StartPattern(pattern Pattern) error {
	pm.StopPattern()

	pm.currentPattern = pattern
	pm.stopChan = make(chan struct{})
	pm.running = true

	go func() {
		defer func() {
			pm.running = false
		}()
		pattern.Start(pm.strip, pm.stopChan)
	}()

	return nil
}

// StopPattern stops the currently running pattern
func (pm *PatternManager) StopPattern() {
	if pm.running && pm.stopChan != nil {
		close(pm.stopChan)
		pm.running = false
	}
}

// IsRunning returns whether a pattern is currently running
func (pm *PatternManager) IsRunning() bool {
	return pm.running
}

// CurrentPattern returns the currently running pattern
func (pm *PatternManager) CurrentPattern() Pattern {
	return pm.currentPattern
}

// ClearStrip turns off all LEDs
func (pm *PatternManager) ClearStrip() {
	pm.StopPattern()
	pm.strip.Clear()
	pm.strip.Show()
}

// WavePattern creates a wave effect that moves around the strip using SetBufferAt
type WavePattern struct {
	WaveColors []color.RGBA
	WaveLength int
	Speed      int // milliseconds between moves
	position   int
}

// NewWavePattern creates a new wave pattern with default values
func NewWavePattern() *WavePattern {
	return &WavePattern{
		WaveColors: []color.RGBA{
			{R: 0, G: 0, B: 1, A: 255}, // Dark blue
			{R: 0, G: 5, B: 1, A: 255}, // Medium blue
			{R: 0, G: 1, B: 1, A: 255}, // Bright blue
			{R: 0, G: 0, B: 0, A: 255}, // Light blue
			{R: 1, G: 1, B: 1, A: 255}, // Very light blue
			{R: 1, G: 0, B: 0, A: 255}, // Light blue
			{R: 0, G: 5, B: 5, A: 255}, // Bright blue
			{R: 0, G: 5, B: 5, A: 255}, // Medium blue
		},
		WaveLength: 8,
		Speed:      100,
		position:   0,
	}
}

func (p *WavePattern) Name() string {
	return "Wave"
}

func (p *WavePattern) Start(strip *peripheral.ColorLedStrip, done <-chan struct{}) error {
	ticker := time.NewTicker(time.Duration(p.Speed) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-ticker.C:
			// Clear the strip
			strip.Clear()

			// Use SetBufferAt to place the wave at the current position
			// This demonstrates wrap-around functionality
			strip.SetBufferAt(p.position, p.WaveColors)

			strip.Show()

			// Move the wave position
			p.position = (p.position + 1) % strip.NumLEDs()

			// Adjust speed based on analog input (inverted for more responsive control)
			analogValue := peripheral.ReadSliderInputPercentage()
			newSpeed := (p.Speed * (100 - analogValue)) / 100
			if newSpeed < 10 {
				newSpeed = 10 // Minimum speed
			}
			ticker.Reset(time.Duration(newSpeed) * time.Millisecond)
		}
	}
}
