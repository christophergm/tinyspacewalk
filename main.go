package main

import (
	"context"
	"image/color"
	"time"

	"machine"

	"golang.org/x/exp/rand"

	"github.com/christophergm/tinyspacewalk/battery"
	"github.com/christophergm/tinyspacewalk/panel"
	"github.com/christophergm/tinyspacewalk/peripheral"
)

var (
	Red    = color.RGBA{25, 0, 0, 255}
	Green  = color.RGBA{0, 25, 0, 255}
	Blue   = color.RGBA{0, 0, 25, 255}
	Yellow = color.RGBA{25, 25, 0, 255}
	Off    = color.RGBA{0, 0, 0, 255}
)

func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configuration - set to true to use real GPIO pins instead of demo mode
	useRealPins := true
	runDemoAllBatteries := false   // Only used when useRealPins is false
	runDemoRandomBatteries := true // Only used when useRealPins is false

	var neoPixel peripheral.NeoPixel
	var boardYellowLight peripheral.BoardYellowLight

	pauseMilliseconds := 300

	// Use simpler seed to avoid overflow on microcontroller
	rand.Seed(uint64(time.Now().Unix()))

	neoPixel = peripheral.NeoPixel{}
	neoPixel.Configure()

	boardYellowLight = peripheral.BoardYellowLight{}
	boardYellowLight.Configure()
	boardYellowLight.StartBlink()

	// Cleanup function for board yellow light
	defer func() {
		boardYellowLight.StopBlink()
	}()

	neoPixel.SetColorAndPause(Off, pauseMilliseconds)

	// Initialize LED strip with new structure
	numLEDs := 144
	ledStrip := peripheral.NewColorLedStrip(numLEDs)
	if err := ledStrip.Configure(); err != nil {
		neoPixel.SetColorAndPause(Red, pauseMilliseconds)
		return // Exit on configuration error
	}

	// Create five batteries with default configuration
	batteries := make([]*battery.Battery, 5)
	for i := 0; i < 5; i++ {
		batteries[i] = battery.NewBattery(battery.FastBatteryConfig())
	}

	var batteryResetButton peripheral.ButtonReader
	var batteryConnects []peripheral.ButtonReader
	var mockBatteryConnects []*peripheral.MockButton
	var mockResetButton *peripheral.MockButton

	if useRealPins {
		// Configure real GPIO pins D0-D5
		// D0: Board reset button
		resetButton := peripheral.NewButton(machine.D40, true) // inverted - pressed when low
		resetButton.Configure()
		batteryResetButton = resetButton

		// D1-D5: Battery connect signals
		batteryConnects = make([]peripheral.ButtonReader, 5)
		pins := []machine.Pin{machine.D30, machine.D32, machine.D34, machine.D36, machine.D38}

		for i, pin := range pins {
			button := peripheral.NewButton(pin, false) // inverted - pressed when low
			button.Configure()
			batteryConnects[i] = button
		}
	} else {
		// Create mock input handlers for demonstration
		mockResetButton = peripheral.NewMockButton()
		batteryResetButton = mockResetButton
		mockBatteryConnects = make([]*peripheral.MockButton, 5)
		for i := 0; i < 5; i++ {
			mockBatteryConnects[i] = peripheral.NewMockButton()
		}

		// Convert mock buttons to ButtonReader interface for panel
		batteryConnects = make([]peripheral.ButtonReader, 5)
		for i := 0; i < 5; i++ {
			batteryConnects[i] = mockBatteryConnects[i]
		}
	}

	// Create and configure the panel
	panelConfig := panel.PanelConfig{
		Batteries:          batteries,
		LEDStrip:           ledStrip,
		BatteryResetButton: batteryResetButton,
		BatteryConnects:    batteryConnects,
		UpdateRate:         50 * time.Millisecond,
	}
	mainPanel := panel.NewPanel(panelConfig)

	// Ensure panel cleanup on exit
	defer mainPanel.Stop()

	// Only run demo sequences when using mock buttons
	if !useRealPins {
		if runDemoAllBatteries {
			mainPanel.StartDemoAllBatteries(neoPixel)
		} else if runDemoRandomBatteries {
			mainPanel.StartDemoRandomBatteries(neoPixel)
		}
	}

	// Run until context is cancelled or panel stops
	select {
	case <-ctx.Done():
		// Context was cancelled
	case <-mainPanel.GetContext().Done():
		// Panel stopped itself
		cancel()
	}

	// Cleanup already handled by defer statements
}
