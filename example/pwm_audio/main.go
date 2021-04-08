package main

import (
	"fmt"
	"time"

	"machine"

	"github.com/wencode/ubit"
	"github.com/wencode/ubit/nrf/pwm"
)

var (
	testEvent2 pwm.Event
)

var (
	//ch0_duty = []uint16{16000}//,8000,4000,2000}
	ch0_duty = []uint16{127, 127, 127, 127}
)

func main() {
	ubit.Display.Init()
	ubit.Display.ShowCharacter('S')

	p, err := pwm.Init(pwm.ID0,
		pwm.WithHandler(
			func(event pwm.Event, context interface{}) {
				fmt.Printf("in handler %x\n", event)
				testEvent2 = event
			}, nil),
		pwm.WithBaseCLK(pwm.CLK_125KHz),
		pwm.WithTopValue(255),
		pwm.WithOutputPin(machine.SPEAKER_PIN),
	)
	if err != nil {
		fmt.Printf("pwm init error: %v\n", err)
		return
	}

	seq := pwm.NewSequence(ch0_duty[:])

	p.SimplePlayback(seq, 2)

	for !p.IsStopped() {
		time.Sleep(time.Millisecond * 1)
	}
	println("stopped")

	time.Sleep(time.Second * 1)

	p.Uninit()
	ubit.Display.Uninit()
}
