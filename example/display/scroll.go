package main

import (
	"time"

	"github.com/wencode/ubit"
	"github.com/wencode/ubit/image5x5"
)

func main() {
	ubit.Display.Init()
	defer ubit.Display.Uninit()

	ubit.Display.ScrollText("Hello")
	time.Sleep(time.Second * 25)
	ubit.Display.Scroll(image5x5.Heart)
	time.Sleep(time.Second * 8)
}
