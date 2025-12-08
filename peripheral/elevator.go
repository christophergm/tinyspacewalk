package peripheral

import (
	"machine"
	"time"
)

type Elevator struct {
	Period   int32
	pwmTimer *machine.TCC
	// Blue Channel
	chB uint8
	// Red Channel
	chR         uint8
	ButtonInput machine.Pin
}

func (e *Elevator) Configure() {
	// Configure the pins
	buttonLedB := machine.PC17
	buttonLedB.Configure(machine.PinConfig{Mode: machine.PinTimer})

	buttonLedR := machine.PC16
	buttonLedR.Configure(machine.PinConfig{Mode: machine.PinTimer})

	buttonInput := machine.PB13
	buttonInput.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	e.ButtonInput = buttonInput

	// Set up PWM timer
	e.pwmTimer = machine.TCC0
	e.pwmTimer.Configure(machine.PWMConfig{})

	ch, err := e.pwmTimer.Channel(buttonLedB)
	if err != nil {
		println("Failed to get PWM channel:", err)
		return
	}
	e.chB = ch

	ch, err = e.pwmTimer.Channel(buttonLedR)
	if err != nil {
		println("Failed to get PWM channel:", err)
		return
	}
	e.chR = ch

}

func (e *Elevator) Run() {
	// Set the PWM period (frequency)
	e.pwmTimer.SetPeriod(1000) // 1 kHz

	// Ramp up and down the PWM duty cycle
	max := e.pwmTimer.Top()

	i := uint32(0)
	onCount := max / uint32(10)
	direction := 1
	e.pwmTimer.Set(e.chR, max/10)
	for {
		if direction == 1 {
			i = i + onCount
		} else {
			i = i - onCount
		}
		if i >= max {
			i = max
			direction = -1
		}
		if i <= 0 {
			i = 0
			direction = 1
		}
		if e.ButtonInput.Get() == true {
			e.pwmTimer.Set(e.chR, 0)
		} else {
			e.pwmTimer.Set(e.chR, max/10)
		}
		e.pwmTimer.Set(e.chB, i)
		time.Sleep(time.Millisecond * 25)
	}
}
