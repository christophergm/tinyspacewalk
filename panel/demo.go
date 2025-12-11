package panel

import (
	"math/rand"
	"time"

	"github.com/christophergm/tinyspacewalk/peripheral"
)

// demoAllBatteries runs a demo sequence of inputs to mock
// battery connect inputs
func DemoAllBatteries(mockBatteryConnects []*peripheral.MockButton, neoPixel peripheral.NeoPixel) {
	for {
		// Demo sequence - simulate input presses
		time.Sleep(2 * time.Second)

		// Set draining to true for all batteries (simulate button press)
		for _, button := range mockBatteryConnects {
			button.SetPressed(true)
			time.Sleep(time.Second)
		}

		time.Sleep(10 * time.Second)

		for _, button := range mockBatteryConnects {
			button.SetPressed(false)
		}
	}
}

// demoRandomBatteries randomly toggles individual inputs to
// mock battery connect inputs
func DemoRandomBatteries(batteryResetButton *peripheral.MockButton, mockBatteryConnects []*peripheral.MockButton, neoPixel peripheral.NeoPixel) {
	for {
		// Wait between 1-3 seconds before next action
		//waitTime := time.Duration(1+rand.Intn(2)) * time.Second
		waitTime := 1 * time.Second
		time.Sleep(waitTime)

		// Pick a random battery (0-4)
		batteryNum := rand.Intn(5)
		// Pick a random action (true/false)
		pressed := rand.Float32() < 0.5
		// Apply to a random input handler
		mockBatteryConnects[batteryNum].SetPressed(pressed)

		time.Sleep(waitTime)
		mockBatteryConnects[0].SetPressed(true)
	}
}
