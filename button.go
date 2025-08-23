package main

import (
	"machine"
	"time"
)

const DEBOUNCE_DELAY = 10 * time.Millisecond

type Button struct {
	currentState bool
	lastState    bool
	lastDebounce time.Time
	pin          machine.Pin
}

func NewButton(pin machine.Pin) *Button {
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	return &Button{
		pin:          pin,
		lastState:    true, // Pull-up means HIGH when not pressed
		lastDebounce: time.Now(),
		currentState: true,
	}
}

func (b *Button) wasClicked() bool {
	reading := b.pin.Get()

	// If the switch changed, due to noise or pressing
	if reading != b.lastState {
		b.lastDebounce = time.Now()
	}

	// If enough time has passed since last state change
	if time.Since(b.lastDebounce) > DEBOUNCE_DELAY {
		// If the button state has actually changed
		if reading != b.currentState {
			b.currentState = reading
			b.lastState = reading

			// Return true only on falling edge (button press with pull-up)
			return !b.currentState
		}
	}
	b.lastState = reading
	return false
}
