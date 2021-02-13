package main

import (
	"fmt"
	"time"

	"machine"

	"github.com/wencode/ubit/nrf52/pwm"
)

func main() {
	machine.LED_ROW_5.Configure(machine.PinConfig{machine.PinOutput})
	machine.LED_COL_4.Configure(machine.PinConfig{machine.PinOutput})
	machine.LED_COL_5.Configure(machine.PinConfig{machine.PinOutput})

	machine.LED_ROW_5.High()
	machine.LED_COL_4.Low()
	time.Sleep(time.Millisecond * 2)

	p, err := pwm.Init(pwm.ID0,
		pwm.WithHandler(
			func(event pwm.Event, context interface{}) {
				fmt.Printf("event %d\n", event)
			}, nil),
		pwm.WithOutputPin(machine.SPEAKER_PIN),
	)
	if err != nil {
		fmt.Printf("pwm init error: %v\n", err)
		return
	}
	machine.LED_COL_5.Low()
	time.Sleep(time.Millisecond * 2)

	seq := pwm.NewSequence(airtel[:])

	p.SimplePlayback(seq, 1)

	for !p.IsStopped() {
		time.Sleep(time.Millisecond * 1)
	}

	p.Uninit()
}
