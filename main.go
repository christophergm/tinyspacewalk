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
	Red    = color.RGBA{25, 0, 0, 20}
	Green  = color.RGBA{0, 25, 0, 20}
	Blue   = color.RGBA{0, 0, 25, 20}
	Yellow = color.RGBA{25, 25, 0, 20}
)

var (
	neoPixel         peripheral.NeoPixel
	boardYellowLight peripheral.BoardYellowLight
	batteryPanel     *panel.Panel
)

func main() {

	pauseMilliseconds := 300

	rand.Seed(uint64(time.Now().UnixNano()))

	neoPixel = peripheral.NeoPixel{}
	neoPixel.Configure()

	boardYellowLight = peripheral.BoardYellowLight{}
	boardYellowLight.Configure()
	boardYellowLight.StartBlink()

	neoPixel.SetColorAndPause(Green, pauseMilliseconds)

	neoPixel.SetColorAndPause(Red, pauseMilliseconds)
	neoPixel.SetColorAndPause(Yellow, pauseMilliseconds)
	neoPixel.SetColorAndPause(Blue, pauseMilliseconds)

	// Initialize LED strip with new structure
	numLEDs := 144
	ledStrip := peripheral.NewColorLedStrip(numLEDs)
	if err := ledStrip.Configure(); err != nil {
		for {
			neoPixel.SetRandomColorAndPause(pauseMilliseconds)
		}
	}

	// Create five batteries with default configuration
	batteries := make([]*battery.Battery, 5)
	for i := 0; i < 5; i++ {
		batteries[i] = battery.NewBattery(battery.FastBatteryConfig())
	}

	// Create mock input handlers for demonstration
	chargedOverrideInputs := make([]panel.InputHandler, 5)
	drainingInputs := make([]panel.InputHandler, 5)
	for i := 0; i < 5; i++ {
		chargedOverrideInputs[i] = panel.NewMockInputHandler()
		drainingInputs[i] = panel.NewMockInputHandler()
	}

	// Create and configure the panel
	panelConfig := panel.PanelConfig{
		Batteries:          batteries,
		LEDStrip:           ledStrip,
		ChargedOverrideIns: chargedOverrideInputs,
		DrainingIns:        drainingInputs,
		UpdateRate:         50 * time.Millisecond,
	}
	batteryPanel = panel.NewPanel(panelConfig)

	// Panel will now monitor the hardware inputs and update battery state accordingly
	// - Press button on D2 to force charged state (charged override)
	// - Press button on D3 to start draining

	for {

		// Demo sequence - simulate input presses
		time.Sleep(2 * time.Second)

		// Set draining to true for all batteries (simulate button press)
		for _, input := range drainingInputs {
			if mockInput, ok := input.(*panel.MockInputHandler); ok {
				mockInput.SetPressed(true)
			}
		}

		time.Sleep(10 * time.Second)

		for _, input := range drainingInputs {
			if mockInput, ok := input.(*panel.MockInputHandler); ok {
				mockInput.SetPressed(false)
			}
		}
	}

	// Set charged override to true (simulate button press)
	//chargedOverrideInput.SetPressed(true)

	select {}
}
