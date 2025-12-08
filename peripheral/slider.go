package peripheral

import (
	"machine"
)

// ReadSliderInput reads from slider pin A0 and returns a value between 0-100
// This can be used for variable delay or other slider input needs
func ReadSliderInputPercentage() int {
	input := machine.ADC{Pin: machine.A0}
	input.Configure(machine.ADCConfig{})
	value := input.Get()
	percentage := int((float64(value) / 262140) * 100)
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}
	return percentage
}

// ReadSliderInputRaw reads from slider pin A0 and returns the raw ADC value
func ReadSliderInputRaw() uint16 {
	input := machine.ADC{Pin: machine.A0}
	input.Configure(machine.ADCConfig{})
	return input.Get()
}

// ReadSliderInputScaled reads slider input and returns a number between 0 and scale
func ReadSliderInputScaled(max int) int {
	percentage := ReadSliderInputPercentage()
	return (max * percentage) / 100
}
