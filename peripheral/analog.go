package peripheral

import (
	"machine"
)

// ReadAnalogInput reads from analog pin A0 and returns a value between 0-100
// This can be used for variable delay or other analog input needs
func ReadAnalogInput() int {
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

// ReadAnalogInputRaw reads from analog pin A0 and returns the raw ADC value
func ReadAnalogInputRaw() uint16 {
	input := machine.ADC{Pin: machine.A0}
	input.Configure(machine.ADCConfig{})
	return input.Get()
}

// ReadAnalogInputAsDelay reads analog input and converts it to a delay in milliseconds
// scale: the maximum delay value in milliseconds
func ReadAnalogInputAsDelay(scale int) int {
	percentage := ReadAnalogInput()
	return (scale * percentage) / 100
}
