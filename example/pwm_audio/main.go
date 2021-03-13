package main

import (
	"fmt"
	"time"

	"machine"

	"github.com/wencode/ubit"
	"github.com/wencode/ubit/nrf52/pwm"
)

func main() {
	ubit.Display.Init()
	ubit.Display.ShowCharacter('S')

	p, err := pwm.Init(pwm.ID0,
		pwm.WithHandler(
			func(event pwm.Event, context interface{}) {
				fmt.Printf("event %d\n", event)
				switch event {
				case pwm.EventStopped:
					ubit.Display.ShowCharacter('T')
				case pwm.EventSeqStarted0:
					ubit.Display.ShowCharacter('0')
				case pwm.EventSeqStarted1:
					ubit.Display.ShowCharacter('1')
				case pwm.EventSeqEnd0:
					ubit.Display.ShowCharacter('E')
				case pwm.EventSeqEnd1:
					ubit.Display.ShowCharacter('F')
				case pwm.EventPWMPeriodEnd:
					ubit.Display.ShowCharacter('P')
				case pwm.EventLoopsDone:
					ubit.Display.ShowCharacter('L')
				}
			}, nil),
		pwm.WithOutputPin(machine.SPEAKER_PIN),
	)
	if err != nil {
		fmt.Printf("pwm init error: %v\n", err)
		return
	}

	seq := pwm.NewSequence(airtel[:])

	p.SimplePlayback(seq, 1)

	for !p.IsStopped() {
		time.Sleep(time.Millisecond * 1)
	}

	time.Sleep(time.Second * 1)

	switch pwm.TestEvent {
	case pwm.EventStopped:
		ubit.Display.ShowCharacter('9')
	case pwm.EventSeqStarted0:
		ubit.Display.ShowCharacter('8')
	case pwm.EventSeqStarted1:
		ubit.Display.ShowCharacter('7')
	case pwm.EventSeqEnd0:
		ubit.Display.ShowCharacter('6')
	case pwm.EventSeqEnd1:
		ubit.Display.ShowCharacter('5')
	case pwm.EventPWMPeriodEnd:
		ubit.Display.ShowCharacter('4')
	case pwm.EventLoopsDone:
		ubit.Display.ShowCharacter('3')
	}
	time.Sleep(time.Second * 1)

	p.Uninit()
	ubit.Display.Uninit()
}
