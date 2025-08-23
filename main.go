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

type Button struct {
	pin        machine.Pin
	wasPressed bool
}

func NewButton(pin machine.Pin) Button {
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	return Button{
		pin:        pin,
		wasPressed: false,
	}
}

func NewCar(playerColor string) Car {
	return Car{
		playerColor: playerColor,
		velocity:    0,
	}
}

func (c *Car) increaseVel() {
	c.velocity++
	println(c.playerColor, ": ", c.velocity)
}

func (b *Button) wasClicked() bool {
	switch {
	// Get() == false when pressed
	case !b.pin.Get() && !b.wasPressed:
		b.wasPressed = true
		time.Sleep(time.Millisecond * 20)
		return true
	case b.pin.Get():
		b.wasPressed = false
		return false
	default:
		return false
	}
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
