package main

import (
	"image/color"
	"machine"
	"math"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/ws2812"
)

const (
	ledPin                = machine.D2
	NUM_LEDS              = 4 * 60
	ENERGY_INCREASE       = 100
	STAMINA_MIN           = .1
	STAMINA_START         = 1.0
	STAMINA_PRESS_LOSS    = .1
	STAMINA_CONSTANT_GAIN = .01
	ENERGY_FACTOR         = 25
	FRICTION_DECAY_FACTOR = .95
	MAX_BRIGHTNESS        = 255
	BRIGHTNESS_FACTOR     = .2
	LAPS                  = 5
)

type GameState int

const (
	Running GameState = iota
	Finished
	Waiting
)

type ZoneType int

const (
	ZoneBoost ZoneType = iota
	ZoneNone
)

type Game struct {
	players      []Player
	strip        *LedStrip
	state        GameState
	zone         Zone
	totalPresses int
}

type Zone struct {
	start    int
	length   int
	zoneType ZoneType
}

func NewZone() Zone {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// zone length indicates the number of laps required to win
	return Zone{start: r.Intn(NUM_LEDS - LAPS), length: LAPS, zoneType: ZoneBoost}
}

type Cell struct {
	cars []*Car
	Zone ZoneType
}

type Player struct {
	car           *Car
	button        *Button
	buttonPresses int
}

type Car struct {
	name     string
	carColor color.RGBA
	pos      float64
	energy   float64
	laps     int
	stamina  float64
}

type LedStrip struct {
	numLeds   int
	device    ws2812.Device
	occupancy []*Cell
}

func NewCar(name string, carColor color.RGBA) *Car {
	return &Car{
		name:     name,
		pos:      0,
		carColor: carColor,
		energy:   0,
		laps:     1,
		stamina:  STAMINA_START,
	}
}

func (c *Car) getStaminaColor() color.RGBA {
	exhaustedR := float64(MAX_BRIGHTNESS)
	exhaustedG := 0.0
	exhaustedB := float64(MAX_BRIGHTNESS)
	var fadeFactor float64
	if c.stamina < 0.8 {
		fadeFactor = 1 - c.stamina/0.8
	} else {
		fadeFactor = 0
	}
	finalR := (float64(c.carColor.R)*(1-fadeFactor) + exhaustedR*fadeFactor) * BRIGHTNESS_FACTOR * math.Pow(
		c.stamina,
		2,
	)
	finalG := (float64(c.carColor.G)*(1-fadeFactor) + exhaustedG*fadeFactor) * BRIGHTNESS_FACTOR * math.Pow(
		c.stamina,
		2,
	)
	finalB := (float64(c.carColor.B)*(1-fadeFactor) + exhaustedB*fadeFactor) * BRIGHTNESS_FACTOR * math.Pow(
		c.stamina,
		2,
	)

	return color.RGBA{
		R: uint8(math.Round(finalR)),
		G: uint8(math.Round(finalG)),
		B: uint8(math.Round(finalB)),
	}
}

func NewLedStrip() *LedStrip {
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	strip := &LedStrip{
		numLeds:   NUM_LEDS,
		device:    ws2812.New(ledPin),
		occupancy: make([]*Cell, NUM_LEDS),
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
		for i := range uint8(MAX_BRIGHTNESS * BRIGHTNESS_FACTOR) {
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

func (l *LedStrip) render(cars []Car, zone Zone) {
	leds := make([]color.RGBA, l.numLeds)
	l.occupancy = make([]*Cell, l.numLeds)
	for i := range l.occupancy {
		zType := ZoneNone
		if i >= zone.start && i < zone.start+zone.length {
			zType = zone.zoneType
		}
		l.occupancy[i] = &Cell{
			cars: []*Car{},
			Zone: zType,
		}
	}
	for idx := range cars {
		c := &cars[idx]
		for i := 0; i < c.laps; i++ {
			pos := int(c.pos) + i
			if pos >= l.numLeds {
				pos = pos % l.numLeds
			}
			l.occupancy[pos].cars = append(l.occupancy[pos].cars, c)
		}
	}
	for i, cell := range l.occupancy {
		switch len(cell.cars) {
		case 0:
			switch cell.Zone {
			case ZoneBoost:
				leds[i] = color.RGBA{
					R: 0,
					G: uint8(math.Round(MAX_BRIGHTNESS * BRIGHTNESS_FACTOR)),
					B: 0,
				}
			default:
				leds[i] = color.RGBA{0, 0, 0, 0}
			}
		case 1:
			leds[i] = cell.cars[0].getStaminaColor()
		default:
			b := uint8(MAX_BRIGHTNESS * BRIGHTNESS_FACTOR)
			leds[i] = color.RGBA{R: b, G: b, B: b}
		}
	}
	l.device.WriteColors(leds)
}

func (g *Game) processInputs() {
	switch g.state {
	case Running:
		for i := range g.players {
			p := &g.players[i]
			if p.button.wasClicked() {
				p.car.stamina = math.Max(STAMINA_MIN, p.car.stamina-STAMINA_PRESS_LOSS)
				p.car.energy += ENERGY_INCREASE * p.car.stamina
				p.buttonPresses++
				g.totalPresses++
				println(
					p.car.name,
					": ",
					int(p.car.pos),
					p.car.energy,
					p.car.stamina,
					p.buttonPresses,
					g.totalPresses,
				)
			}
		}
		if g.totalPresses == 120 {
			g.zone = NewZone()
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
	g.zone = NewZone()
	g.totalPresses = 0
	for i := range g.players {
		p := &g.players[i]
		p.car.pos = 0
		p.car.laps = 1
		p.car.energy = 0
		p.buttonPresses = 0
		p.car.stamina = STAMINA_START
	}
	g.state = Waiting
}

func (g *Game) calcNewPos(duration time.Duration) {
	for _, p := range g.players {
		p.car.energy *= FRICTION_DECAY_FACTOR
		vel := math.Sqrt(math.Max(0, p.car.energy) * ENERGY_FACTOR)
		newPos := p.car.pos + vel*duration.Seconds()
		// check if in zone
		if int(newPos) >= g.zone.start && int(newPos) < g.zone.start+g.zone.length &&
			p.car.energy < 50 {
			p.car.energy += 3000
			p.car.stamina = STAMINA_START
		}
		// calculate new position
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
		strip: NewLedStrip(),
		zone:  NewZone(),
		players: []Player{
			{
				car: NewCar(
					"blue",
					color.RGBA{
						R: 0,
						G: uint8(math.Round(0.49 * MAX_BRIGHTNESS)),
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
				p.car.stamina = math.Min(STAMINA_START, p.car.stamina+STAMINA_CONSTANT_GAIN)
			}
			g.strip.render(cars, g.zone)
			g.calcNewPos(interval)
		}
		time.Sleep(interval)
	}
}
