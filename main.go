package main

import (
	"machine"
	"time"
)

const statusLed = machine.LED

type Game struct {
	players []Player
}

type Player struct {
	car    *Car
	button *Button
}

type Car struct {
	velocity    int
	playerColor string
}

func NewCar(playerColor string) *Car {
	return &Car{
		playerColor: playerColor,
		velocity:    0,
	}
}

func (c *Car) increaseVel() {
	c.velocity++
	println(c.playerColor, ": ", c.velocity)
}

func blink(led machine.Pin) {
	led.High()
	time.Sleep(time.Millisecond * 10)
	led.Low()
}

func (g *Game) processInputs() {
	for _, p := range g.players {
		if p.button.wasClicked() {
			blink(statusLed)
			p.car.increaseVel()
		}
	}
}

func main() {
	game := Game{
		players: []Player{
			{
				car:    NewCar("green"),
				button: NewButton(machine.D7),
			},
			{
				car:    NewCar("red"),
				button: NewButton(machine.D8),
			},
		},
	}
	statusLed.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for {
		game.processInputs()
	}
}
