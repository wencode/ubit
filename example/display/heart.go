package main

import (
	"time"

	"github.com/wencode/ubit"
	"github.com/wencode/ubit/image5x5"
)

func main() {
	ubit.Display.Init()
	defer ubit.Display.Uninit()

	for i := 0; i < 4; i++ {
		ubit.Display.Rotate(i)
		ubit.Display.ShowCharacter('A')
		time.Sleep(time.Second * 3)
		ubit.Display.Show(image5x5.Heart)
		time.Sleep(time.Second * 3)
		ubit.Display.Show(image5x5.No)
		time.Sleep(time.Second * 3)
		ubit.Display.Clear()
	}
}
