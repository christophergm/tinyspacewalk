package main

import (
	"image/color"
	"time"

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

	var neoPixel peripheral.NeoPixel
	var boardYellowLight peripheral.BoardYellowLight

	pauseMilliseconds := 300

	rand.Seed(uint64(time.Now().UnixNano()))

	neoPixel = peripheral.NeoPixel{}
	neoPixel.Configure()

	boardYellowLight = peripheral.BoardYellowLight{}
	boardYellowLight.Configure()
	boardYellowLight.StartBlink()

	neoPixel.SetColorAndPause(Off, pauseMilliseconds)

	// Initialize LED strip with new structure
	numLEDs := 144
	ledStrip := peripheral.NewColorLedStrip(numLEDs)
	if err := ledStrip.Configure(); err != nil {
		neoPixel.SetColorAndPause(Red, pauseMilliseconds)
	}

	// Create five batteries with default configuration
	batteries := make([]*battery.Battery, 5)
	for i := 0; i < 5; i++ {
		batteries[i] = battery.NewBattery(battery.FastBatteryConfig())
	}

	// Create mock input handlers for demonstration
	batteryResetButton := peripheral.NewMockButton()
	mockBatteryConnects := make([]*peripheral.MockButton, 5)
	for i := 0; i < 5; i++ {
		mockBatteryConnects[i] = peripheral.NewMockButton()
	}

	// Convert mock buttons to ButtonReader interface for panel
	batteryConnects := make([]peripheral.ButtonReader, 5)
	for i := 0; i < 5; i++ {
		batteryConnects[i] = mockBatteryConnects[i]
	}

	// Create and configure the panel
	panelConfig := panel.PanelConfig{
		Batteries:          batteries,
		LEDStrip:           ledStrip,
		BatteryResetButton: batteryResetButton,
		BatteryConnects:    batteryConnects,
		UpdateRate:         50 * time.Millisecond,
	}
	_ = panel.NewPanel(panelConfig)

	if false {
		go panel.DemoAllBatteries(mockBatteryConnects, neoPixel)
	} else {
		go panel.DemoRandomBatteries(batteryResetButton, mockBatteryConnects, neoPixel)
	}

	select {}
}
