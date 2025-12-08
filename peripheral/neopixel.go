package peripheral

import (
	"image/color"
	"machine"
	"time"

	"golang.org/x/exp/rand"
	"tinygo.org/x/drivers/ws2812"
)

type NeoPixel struct {
	NeoPixelDriver ws2812.Device
}

func (d *NeoPixel) Configure() {
	// Configure the onboard NeoPixel
	neoPixelPin := machine.PC24
	neoPixelPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.NeoPixelDriver = ws2812.NewWS2812(neoPixelPin)
}

// SetRandomColor sets the NeoPixel to a random color
func (d *NeoPixel) SetRandomColorAndPause(pauseMilliseconds int) {
	// Generate random RGB values
	r := uint8(rand.Intn(10))
	g := uint8(rand.Intn(10))
	b := uint8(rand.Intn(10))

	// Write the color to the NeoPixel
	d.NeoPixelDriver.WriteColors([]color.RGBA{{r, g, b, 50}})
	if pauseMilliseconds > 0 {
		time.Sleep(time.Millisecond * time.Duration(pauseMilliseconds))
	}
}

func (d *NeoPixel) SetColorAndPause(col color.RGBA, pauseMilliseconds int) {
	// Write the color to the NeoPixel
	d.NeoPixelDriver.WriteColors([]color.RGBA{{col.R, col.G, col.B, 20}})
	if pauseMilliseconds > 0 {
		time.Sleep(time.Millisecond * time.Duration(pauseMilliseconds))
	}
}
