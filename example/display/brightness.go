package main

import (
	"time"

	"github.com/wencode/ubit"
)

func main() {
	d := ubit.NewModDisplay()
	d.Init()
	defer d.Uninit()

	d.SetBrightness(1, 2, 1)
	time.Sleep(time.Second * 3)
	d.SetBrightness(2, 2, 1)
	time.Sleep(time.Second * 10)
}
