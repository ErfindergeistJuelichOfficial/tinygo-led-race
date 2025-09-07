package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	ledPin       = machine.D2
	MAX_SPEED    = 30.0
	FRICTION     = 0.001
	VEL_INCREASE = 1.0
)

type Game struct {
	players []Player
	strip   *LedStrip
}

type Player struct {
	car    *Car
	button *Button
}

type Car struct {
	name        string
	playerColor color.RGBA
	pos         int
	vel         float32
	maxSpeed    float32
	friction    float32 // Friction coefficient (0.0 - 1.0)
}

type LedStrip struct {
	numLeds int
	device  ws2812.Device
}

func NewCar(name string, playerColor color.RGBA, maxSpeed, friction float32) *Car {
	return &Car{
		name:        name,
		pos:         0,
		playerColor: playerColor,
		vel:         0,
		maxSpeed:    maxSpeed,
		friction:    friction,
	}
}

func NewLedStrip(numLeds int) *LedStrip {
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip := &LedStrip{
		numLeds: numLeds,
		device:  ws2812.New(ledPin),
	}
	strip.reset()
	return strip
}

func (l *LedStrip) reset() {
	leds := []color.RGBA{}
	for range l.numLeds {
		leds = append(leds, color.RGBA{R: 0, G: 0, B: 0})
		l.device.WriteColors(leds)
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
}

func blink(led machine.Pin) {
	led.High()
	time.Sleep(time.Millisecond * 10)
	led.Low()
}

func (g *Game) processInputs() {
	for _, p := range g.players {
		if p.button.wasClicked() {
			p.car.pos++
			println(p.car.name, ": ", p.car.pos)
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

func haveSamePosition(players []Player) bool {
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[i].car.pos == players[j].car.pos {
				return true
			}
		}
	}
	return false
}

func (g *Game) draw() {
	leds := make([]color.RGBA, g.strip.numLeds)
	if haveSamePosition(g.players) {
		leds[g.players[0].car.pos] = color.RGBA{R: 255, G: 255, B: 255}
	} else {
		for _, p := range g.players {
			leds[p.car.pos] = p.car.playerColor
		}
	}
	g.strip.device.WriteColors(leds)
}

func main() {
	game := Game{
		strip: NewLedStrip(60 * 4),
		players: []Player{
			{
				car:    NewCar("green", color.RGBA{R: 0, G: 255, B: 0}, MAX_SPEED, FRICTION),
				button: NewButton(machine.D7),
			},
			{
				car:    NewCar("red", color.RGBA{R: 255, G: 0, B: 0}, MAX_SPEED, FRICTION),
				button: NewButton(machine.D8),
			},
		},
	}

	for {
		game.processInputs()
		game.applyFriction()
		game.draw()
	}
}
