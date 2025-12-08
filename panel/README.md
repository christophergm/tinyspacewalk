# Panel Package

The panel package provides LED strip control and input handling for the TinySpaceWalk battery monitoring system. It displays different LED patterns based on battery state and monitors digital inputs to control battery behavior.

## Features

- **LED Strip Control**: Abstract interface for controlling color LED strips
- **Input Handling**: Digital input monitoring for user controls
- **State-based Animations**: Different LED patterns for each battery state
- **Real-time Updates**: Continuous monitoring and display updates

## Battery State LED Patterns

| Battery State | LED Pattern |
|---------------|-------------|
| **Charged** | All LEDs solid green |
| **Disconnecting** | All LEDs flashing yellow (1 Hz) |
| **Draining** | Red bar showing battery level + top 2 LEDs pulsing red |
| **Dead** | All LEDs blinking red (2 Hz) |
| **Charging** | Green bar showing battery level + yellow charging indicator |
| **Unknown** | All LEDs pulsing blue |

## Usage

### Basic Setup

```go
package main

import (
    "time"
    "github.com/christophergm/tinyspacewalk/battery"
    "github.com/christophergm/tinyspacewalk/panel"
)

func main() {
    // Create battery
    bat := battery.NewBattery(battery.DefaultBatteryConfig())
    defer bat.Stop()
    
    // Create LED strip (use your hardware-specific implementation)
    ledStrip := panel.NewMockLEDStrip(10) // 10 pixels for demo
    
    // Create input handlers (use your hardware-specific implementation)
    chargedOverrideInput := panel.NewMockInputHandler()
    drainingInput := panel.NewMockInputHandler()
    
    // Create panel
    panelConfig := panel.PanelConfig{
        Battery:           bat,
        LEDStrip:          ledStrip,
        ChargedOverrideIn: chargedOverrideInput,
        DrainingIn:        drainingInput,
        UpdateRate:        50 * time.Millisecond, // 20 FPS
    }
    p := panel.NewPanel(panelConfig)
    defer p.Stop()
    
    // Panel now runs automatically, monitoring inputs and updating LEDs
    select {} // Keep program running
}
```

### Custom LED Strip Implementation

To use with real hardware, implement the `LEDStrip` interface:

```go
type MyLEDStrip struct {
    // Your hardware-specific fields
}

func (m *MyLEDStrip) SetPixel(index int, color panel.Color) error {
    // Set individual pixel color
    return nil
}

func (m *MyLEDStrip) SetAll(color panel.Color) error {
    // Set all pixels to same color
    return nil
}

func (m *MyLEDStrip) Show() error {
    // Update physical LED strip
    return nil
}

func (m *MyLEDStrip) Clear() error {
    // Turn off all pixels
    return nil
}

func (m *MyLEDStrip) GetLength() int {
    // Return number of pixels
    return 10
}
```

### Custom Input Handler Implementation

To use with real hardware inputs, implement the `InputHandler` interface:

```go
type MyInputHandler struct {
    pin int // GPIO pin number
}

func (m *MyInputHandler) IsPressed() bool {
    // Read digital input state from hardware
    return readGPIO(m.pin)
}
```

## Input Controls

The panel monitors two digital inputs:

1. **Charged Override Input**: When active, forces battery to 100% charge and "Charged" state
2. **Draining Input**: When active, puts battery into draining mode

## Animation Details

### Draining State
- Bottom LEDs show a red "fuel gauge" based on battery level
- Top 2 LEDs pulse red to indicate active draining
- As battery drains, fewer LEDs are lit

### Charging State  
- Green LEDs show current charge level
- Yellow "charging indicator" moves up the strip
- When full, transitions to solid green

### Flash/Pulse Timing
