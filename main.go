package main

import (
	"image/color"
	"machine"
	"math"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	ledPin       = machine.D2
	VEL_INCREASE = 20
	AIRRES       = 0.02
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
	pos         float64
	vel         float64
	laps        int
}

type LedStrip struct {
	numLeds int
	device  ws2812.Device
}

func NewCar(name string, playerColor color.RGBA) *Car {
	return &Car{
		name:        name,
		pos:         0,
		playerColor: playerColor,
		vel:         0,
		laps:        0,
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

func (g *Game) processInputs() {
	for _, p := range g.players {
		if p.button.wasClicked() {
			p.car.vel += VEL_INCREASE
			println(p.car.name, ": ", int(p.car.pos), p.car.vel)
		}
	}
}

func haveSamePosition(players []Player) bool {
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if int(players[i].car.pos) == int(players[j].car.pos) {
				return true
			}
		}
	}
	return false
}

func (g *Game) draw() {
	leds := make([]color.RGBA, g.strip.numLeds)
	if haveSamePosition(g.players) {
		leds[int(g.players[0].car.pos)] = color.RGBA{R: 255, G: 255, B: 255}
	} else {
		for _, p := range g.players {
			leds[int(p.car.pos)] = p.car.playerColor
		}
	}
	g.strip.device.WriteColors(leds)
}

func (g *Game) calcNewPos(duration time.Duration) {
	for _, p := range g.players {
		a := -AIRRES * math.Pow(p.car.vel, 2)
		p.car.vel = p.car.vel + a*duration.Seconds()
		newPos := p.car.pos + p.car.vel*duration.Seconds()
		if int(newPos) >= g.strip.numLeds {
			p.car.laps++
			p.car.pos = float64(int(newPos) % g.strip.numLeds)
		} else {
			p.car.pos = newPos
		}
	}
}

func main() {
	g := Game{
		strip: NewLedStrip(60 * 4),
		players: []Player{
			{
				car:    NewCar("green", color.RGBA{R: 0, G: 255, B: 0}),
				button: NewButton(machine.D7),
			},
			{
				car:    NewCar("red", color.RGBA{R: 255, G: 0, B: 0}),
				button: NewButton(machine.D8),
			},
		},
	}

	interval := 10 * time.Millisecond
	for {
		g.processInputs()
		g.calcNewPos(interval)
		g.draw()
		time.Sleep(interval)
	}
}
