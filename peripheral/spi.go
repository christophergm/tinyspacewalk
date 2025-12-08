package peripheral

import (
	"machine"
	"time"
)

type Spi struct {
	Spi machine.SPI
}

func (s *Spi) Configure() error {
	s.Spi = *machine.SPI0
	return s.Spi.Configure(machine.SPIConfig{
		// Frequency: 4000000,      // 4 MHz, typical for APA102
		// SCK:       machine.PD09, // SCK
		// SDO:       machine.PD08, // MOSI
	})
}

func (s *Spi) Start() {
	for {
		s.Spi.Transfer(byte(0x53))
		time.Sleep(500 * time.Millisecond)

	}
}
