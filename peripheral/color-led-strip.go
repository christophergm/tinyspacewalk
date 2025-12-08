package peripheral

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/apa102"
)

// ColorLedStrip represents an APA102 LED strip peripheral
type ColorLedStrip struct {
	buffer   []color.RGBA
	numLEDs  int
	ledStrip *apa102.Device
}

// NewColorLedStrip creates a new ColorLedStrip instance
func NewColorLedStrip(numLEDs int) *ColorLedStrip {
	return &ColorLedStrip{
		numLEDs: numLEDs,
		buffer:  make([]color.RGBA, numLEDs),
	}
}

// Configure initializes the SPI interface and LED strip driver
func (d *ColorLedStrip) Configure() error {
	spi := machine.SPI0
	err := spi.Configure(machine.SPIConfig{
		// Default SPI configuration for APA102
		// Frequency, pins, and mode can be customized as needed
	})
	if err != nil {
		return err
	}

	d.ledStrip = apa102.New(spi)
	return nil
}

// SetPixel sets a single pixel to the specified color
func (d *ColorLedStrip) SetPixel(index int, c color.RGBA) {
	if index >= 0 && index < d.numLEDs {
		d.buffer[index] = c
	}
}

// SetAll sets all pixels to the specified color
func (d *ColorLedStrip) SetAll(c color.RGBA) {
	for i := 0; i < d.numLEDs; i++ {
		d.buffer[i] = c
	}
}

// Clear turns off all LEDs (sets them to black)
func (d *ColorLedStrip) Clear() {
	d.SetAll(color.RGBA{R: 0, G: 0, B: 0, A: 255})
}

// GetPixel returns the color of a specific pixel
func (d *ColorLedStrip) GetPixel(index int) color.RGBA {
	if index >= 0 && index < d.numLEDs {
		return d.buffer[index]
	}
	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

// GetBuffer returns a copy of the current buffer
func (d *ColorLedStrip) GetBuffer() []color.RGBA {
	bufferCopy := make([]color.RGBA, len(d.buffer))
	copy(bufferCopy, d.buffer)
	return bufferCopy
}

// SetBuffer sets the entire buffer to the provided colors
func (d *ColorLedStrip) SetBuffer(colors []color.RGBA) {
	minLen := len(colors)
	if minLen > d.numLEDs {
		minLen = d.numLEDs
	}
	for i := 0; i < minLen; i++ {
		d.buffer[i] = colors[i]
	}
}

// SetBufferAt writes colors starting at the specified index with wrap-around
// startIndex: starting position in the LED strip
// colors: slice of colors to write
// Writes at most numLEDs values and uses modulo math to wrap around
func (d *ColorLedStrip) SetBufferAt(startIndex int, colors []color.RGBA) {
	if len(colors) == 0 {
		return
	}

	// Normalize start index to valid range
	startIndex = startIndex % d.numLEDs
	if startIndex < 0 {
		startIndex += d.numLEDs
	}

	// Write at most numLEDs values
	writeLen := len(colors)
	if writeLen > d.numLEDs {
		writeLen = d.numLEDs
	}

	for i := 0; i < writeLen; i++ {
		bufferIndex := (startIndex + i) % d.numLEDs
		d.buffer[bufferIndex] = colors[i]
	}
}

// Show updates the LED strip with the current buffer contents
func (d *ColorLedStrip) Show() {
	if d.ledStrip != nil {
		d.ledStrip.WriteColors(d.buffer)
	}
}

// NumLEDs returns the number of LEDs in the strip
func (d *ColorLedStrip) NumLEDs() int {
	return d.numLEDs
}
