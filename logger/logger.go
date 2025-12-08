package logger

import (
	"github.com/chris/tinyspacewalk/peripheral"
)

type Logger struct {
	pixel peripheral.NeoPixel
}

func NewLogger(pixel peripheral.NeoPixel) *Logger {
	return &Logger{
		pixel: pixel,
	}
}

// func (l *Logger) Blink(color color.RGBA) {
// 	if level >= l.level {
// 		l.pixel.SetColor(0, 0, 255)
// 	}
// }
