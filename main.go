package main

import (
	"image/color"
	"machine"
	"math"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	ledPin                      = machine.D2
	ENERGY_INCREASE             = 100
	ENERGY_FACTOR               = 25
	FRICTION_DECAY_FACTOR       = 0.95
	MAX_BRIGHTNESS        uint8 = 30
	LAPS                        = 5
)

type GameState int

const (
	Running GameState = iota
	Finished
	Waiting
)

type Game struct {
	players []Player
	strip   *LedStrip
	state   GameState
}

type Player struct {
	car    *Car
	button *Button
}

type Car struct {
	name     string
	carColor color.RGBA
	pos      float64
	energy   float64
	laps     int
}

type LedStrip struct {
	numLeds   int
	device    ws2812.Device
	occupancy [][]Car
}

func NewCar(name string, carColor color.RGBA) *Car {
	return &Car{
		name:     name,
		pos:      0,
		carColor: carColor,
		energy:   0,
		laps:     1,
	}
}

func NewLedStrip(numLeds int) *LedStrip {
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip := &LedStrip{
		numLeds:   numLeds,
		device:    ws2812.New(ledPin),
		occupancy: make([][]Car, numLeds),
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

func (l *LedStrip) illuminate(c color.RGBA) {
	leds := []color.RGBA{}
	for range l.numLeds {
		leds = append(leds, c)
	}
	l.device.WriteColors(leds)
	time.Sleep(1 * time.Second)
}

func (l *LedStrip) pulseWhite(repetitions int) {
	const delay = time.Millisecond * 10
	for range repetitions {
		time.Sleep(time.Millisecond * 1000)
		for i := range MAX_BRIGHTNESS {
			leds := []color.RGBA{}
			for range l.numLeds {
				leds = append(leds, color.RGBA{R: i, G: i, B: i})
			}
			l.device.WriteColors(leds)
			time.Sleep(delay)
		}
		l.clear()
	}
}

func (l *LedStrip) render(cars []Car) {
	leds := make([]color.RGBA, l.numLeds)
	l.occupancy = make([][]Car, l.numLeds)
	for _, c := range cars {
		for i := range c.laps {
			pos := int(c.pos) + i
			if pos >= l.numLeds {
				pos = pos % l.numLeds
				l.occupancy[pos] = append(l.occupancy[pos], c)
			} else {
				l.occupancy[pos] = append(l.occupancy[pos], c)
			}
		}
	}
	for i, cars := range l.occupancy {
		switch {
		case len(cars) == 1:
			leds[i] = cars[0].carColor
		case len(cars) > 1:
			leds[i] = color.RGBA{R: MAX_BRIGHTNESS, G: MAX_BRIGHTNESS, B: MAX_BRIGHTNESS}
		}
	}
	l.device.WriteColors(leds)
}

func (g *Game) processInputs() {
	switch g.state {
	case Running:
		for _, p := range g.players {
			if p.button.wasClicked() {
				p.car.energy += ENERGY_INCREASE
				println(p.car.name, ": ", int(p.car.pos), p.car.energy)
			}
		}
	case Waiting:
		for _, p := range g.players {
			if p.button.wasClicked() {
				g.start()
			}
		}
	}
}

func (g *Game) start() {
	g.strip.pulseWhite(3)
	g.state = Running
}

func (g *Game) end(winner Player) {
	g.state = Finished
	g.strip.illuminate(winner.car.carColor)
	for _, p := range g.players {
		p.car.pos = 0
		p.car.laps = 1
		p.car.energy = 0
	}
	g.state = Waiting
}

func (g *Game) calcNewPos(duration time.Duration) {
	for _, p := range g.players {
		p.car.energy *= FRICTION_DECAY_FACTOR
		vel := math.Sqrt(math.Max(0, p.car.energy) * ENERGY_FACTOR)
		newPos := p.car.pos + vel*duration.Seconds()
		if int(newPos) >= g.strip.numLeds {
			p.car.laps++
			p.car.pos = float64(int(newPos) % g.strip.numLeds)
		} else {
			p.car.pos = newPos
		}
		if p.car.laps >= LAPS {
			g.end(p)
		}
	}
}

func main() {
	g := Game{
		strip: NewLedStrip(60 * 4),
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
	}

	g.start()
	interval := 10 * time.Millisecond
	for {
		g.processInputs()
		switch g.state {
		case Running:
			cars := []Car{}
			for _, p := range g.players {
				cars = append(cars, *p.car)
			}
			g.strip.render(cars)
			g.calcNewPos(interval)
		}
		time.Sleep(interval)
	}
}
