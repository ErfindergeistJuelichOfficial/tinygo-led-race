package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	STATUSLED    = machine.LED
	ledPin       = machine.D2
	MAX_SPEED    = 30.0
	FRICTION     = 0.001
	VEL_INCREASE = 1.0
)

type Game struct {
	players []Player
}

type Player struct {
	car    *Car
	button *Button
}

type Car struct {
	playerColor string
	pos         float32 // Position along LED strip
	vel         float32 // Velocity (can be negative for reverse)
	maxSpeed    float32 // Maximum speed
	friction    float32 // Friction coefficient (0.0 - 1.0)
}

func NewCar(playerColor string, maxSpeed, friction float32) *Car {
	return &Car{
		pos:         0,
		playerColor: playerColor,
		vel:         0,
		maxSpeed:    maxSpeed,
		friction:    friction,
	}
}

func (c *Car) increaseVel(increase float32) {
	c.vel += increase
	// Limit maximum speed
	if c.vel > c.maxSpeed {
		c.vel = c.maxSpeed
	} else if c.vel < -c.maxSpeed {
		c.vel = -c.maxSpeed
	}
	println(c.playerColor, ": ", c.vel)
}

func blink(led machine.Pin) {
	led.High()
	time.Sleep(time.Millisecond * 10)
	led.Low()
}

func (g *Game) processInputs() {
	for _, p := range g.players {
		if p.button.wasClicked() {
			blink(STATUSLED)
			p.car.increaseVel(VEL_INCREASE)
		}
	}
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// Apply friction (called every frame)
func (g *Game) applyFriction() {
	for _, p := range g.players {
		p.car.vel -= p.car.friction
		// Stop very slow movement to prevent jitter
		if abs(p.car.vel) < 0.01 {
			p.car.vel = 0
		}
	}
}

func main() {
	game := Game{
		players: []Player{
			{
				car:    NewCar("green", MAX_SPEED, FRICTION),
				button: NewButton(machine.D7),
			},
			{
				car:    NewCar("red", MAX_SPEED, FRICTION),
				button: NewButton(machine.D8),
			},
		},
	}
	STATUSLED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ws := ws2812.New(ledPin)
	for {
		game.processInputs()
		game.applyFriction()
		leds := []color.RGBA{}
		for i := range 60 {
			leds = append(leds, color.RGBA{R: uint8(i), G: uint8(i), B: uint8(i)})
		}
		ws.WriteColors(leds)
	}
}
