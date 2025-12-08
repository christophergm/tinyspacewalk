package peripheral

import (
	"machine"
	"time"
)

type BoardYellowLight struct {
	Led machine.Pin
}

func (e *BoardYellowLight) Configure() {
	// Blink yellow board LED
	e.Led = machine.PC30
}

func (e *BoardYellowLight) StartBlink() {

	e.Led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	go func() {
		for {
			e.Led.Low()
			time.Sleep(time.Millisecond * 250)

			e.Led.High()
			time.Sleep(time.Millisecond * 250)
		}
	}()
}
