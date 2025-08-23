package main

import (
	"machine"
	"time"
)

type Player struct {
	car    Car
	button Button
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

func main() {
	led := machine.LED
	button1 := NewButton(machine.D7)
	car1 := NewCar("green")
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for {
		if button1.wasClicked() {
			blink(led)
			car1.increaseVel()
		}
	}
}
