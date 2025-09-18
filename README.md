# TinyGo LED Race

A competitive 1D racing game designed for microcontrollers using TinyGo, where players control cars racing along an LED strip using button inputs.

## Features

- **1D LED Strip Racing**: Fast-paced racing action displayed on a single LED strip
- **Stamina System**: Strategic gameplay balancing speed vs. endurance
- **Visual Feedback**: Car colors fade to purple as stamina depletes

## Game Mechanics

### Racing Physics
The game uses an energy-based physics system similar to OpenLED Race:
- Each button press adds energy to your car
- Velocity = √(energy)
- Natural friction causes deceleration over time
- Strategic pressing is more effective than button mashing

### Stamina System
Players must balance speed with endurance:
- **High Stamina**: Car displays in full, bright color
- **Medium Stamina**: Car brightness dims proportionally
- **Low Stamina**: Car color fades toward purple, indicating exhaustion
- **Recovery**: Stamina regenerates slowly when not pressing buttons

This creates strategic depth where players must decide when to sprint and when to coast.

## Installation

### Prerequisites
- [TinyGo](https://tinygo.org/getting-started/install/) installed
- Compatible microcontroller
- LED strip driver library

### Building and Flashing
```bash
# Build and flash
tinygo flash -target=nano-rp2040
```

## Configuration

Adjust these constants in `main.go` for your setup:

```go
const (
	ledPin                      = machine.D2
	ENERGY_INCREASE             = 100
	STAMINA_MIN                 = .1
	STAMINA_START               = 1.0
	STAMINA_PRESS_LOSS          = .1
	STAMINA_CONSTANT_GAIN       = .01
	ENERGY_FACTOR               = 25
	FRICTION_DECAY_FACTOR       = .95
	MAX_BRIGHTNESS        uint8 = 30
	LAPS                        = 5
)
```

## Customization

### Adding More Players
Modify the player array in `main.go`
```go
players: []Player{
    {
        car: NewCar(
            "blue",
            color.RGBA{
                R: 0,
                G: uint8(math.Round(0.49 * float64(MAX_BRIGHTNESS))),
                B: MAX_BRIGHTNESS,
            },
        ),
        button: NewButton(machine.D7),
    },
    {
        car:    NewCar("red", color.RGBA{R: MAX_BRIGHTNESS, G: 0, B: 0}),
        button: NewButton(machine.D8),
    },
},
```

## Acknowledgments

- Heavily inspired by [OpenLED Race](https://openledrace.net/)
- Built with the excellent [TinyGo](https://tinygo.org/) project
- LED strip libraries and community examples
