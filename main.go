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
	numLeds   int
	device    ws2812.Device
	occupancy map[int][]Car
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
		numLeds:   numLeds,
		device:    ws2812.New(ledPin),
		occupancy: make(map[int][]Car),
	}
	strip.clear()
	return strip
}

func (l *LedStrip) clear() {
	leds := []color.RGBA{}
	for range l.numLeds {
		leds = append(leds, color.RGBA{R: 0, G: 0, B: 0})
	}
	l.device.WriteColors(leds)
}

func (l *LedStrip) pulseWhite(repetitions int) {
	const steps uint8 = 20
	const delay = time.Millisecond * 50
	for range repetitions {
		for i := range steps {
			leds := []color.RGBA{}
			for range l.numLeds {
				leds = append(leds, color.RGBA{R: i, G: i, B: i})
			}
			l.device.WriteColors(leds)
			time.Sleep(delay)
		}
		l.clear()
		time.Sleep(time.Millisecond * 1000)
	}
}

func (l *LedStrip) render() {
}

func (g *Game) processInputs() {
	for _, p := range g.players {
		if p.button.wasClicked() {
			p.car.vel += VEL_INCREASE
			println(p.car.name, ": ", int(p.car.pos), p.car.vel)
		}
	}
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
		strip: NewLedStrip(60 * 2),
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

	g.strip.pulseWhite(3)

	interval := 10 * time.Millisecond
	for {
		g.processInputs()
		g.calcNewPos(interval)
		g.strip.render()
		time.Sleep(interval)
	}
}
