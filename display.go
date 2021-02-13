package ubit

import (
	"image/color"
	"sync"
	"time"

	"machine"

	"github.com/wencode/ubit/font5x5"
	"github.com/wencode/ubit/image5x5"
)

const (
	display_width  = 5
	display_height = 5
)

const (
	animTypeScroll = iota
)

type ModDisplay struct {
	rowPins [5]machine.Pin
	colPins [5]machine.Pin

	// runing at mono-core CPU, no data race problem
	dirty  bool
	buffer [display_width * display_height]uint8
	quitch chan struct{}
	quitWg sync.WaitGroup

	anim struct {
		animType int32
		elapse   int32
		interval int32
		count    int32
	}
	scroll struct {
		value []byte
		cur   int
	}

	rotation int32
	// for update
	lastTime time.Time
}

func NewModDisplay() *ModDisplay {
	return &ModDisplay{
		rowPins: [5]machine.Pin{
			machine.LED_ROW_1,
			machine.LED_ROW_2,
			machine.LED_ROW_3,
			machine.LED_ROW_4,
			machine.LED_ROW_5,
		},
		colPins: [5]machine.Pin{
			machine.LED_COL_1,
			machine.LED_COL_2,
			machine.LED_COL_3,
			machine.LED_COL_4,
			machine.LED_COL_5,
		},
		quitch: make(chan struct{}),
	}
}

func (d *ModDisplay) Init() {
	for i := 0; i < 5; i++ {
		d.rowPins[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
		d.colPins[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	d.Clear()
	go d.bgloop()
}

func (d *ModDisplay) Uninit() {
	d.quitWg.Add(1)
	d.quitch <- struct{}{}
	d.quitWg.Wait()
}

func (d *ModDisplay) SetPixel(x, y int16, c color.RGBA) {

}

func (d *ModDisplay) SetBrightness(x, y int16, value uint8) {
	if x < 0 || x > display_width {
		return
	}
	if y < 0 || y > display_height {
		return
	}
	d.buffer[y*display_width+x] = value
}

func (d *ModDisplay) Clear() {
	for i := 0; i < 5; i++ {
		d.rowPins[i].Low()
		d.colPins[i].High()
	}
}

func (d *ModDisplay) Show(img image5x5.Image) {
	copy(d.buffer[:], []uint8(img[:]))
}

func (d *ModDisplay) ShowCharacter(c byte) {
	d.Show(font5x5.GenImage5x5(c, 255))
}

func (d *ModDisplay) Scroll(img image5x5.Image) {

}

func (d *ModDisplay) ScrollText(text string) {
	d.anim.elapse = 0
	d.anim.interval = 1000
	d.anim.count = 5
	d.scroll.value = []byte(text)
	d.scroll.cur = 0
	d.ShowCharacter(d.scroll.value[0])
}

func (d *ModDisplay) Rotate(num_ccw int) {
	d.rotation = int32(num_ccw) % 4
}

func (d *ModDisplay) bgloop() {
	d.lastTime = time.Now()
LOOP:
	for {
		select {
		case <-d.quitch:
			break LOOP
		default:
		}

		d.update()
		d.render()
	}
	d.Clear()
	d.quitWg.Done()
}

func (d *ModDisplay) update() {
	diff := int32(time.Now().Sub(d.lastTime).Milliseconds())
	d.animUpdate(diff)

}

func (d *ModDisplay) animUpdate(diff int32) {
	if d.anim.interval <= 0 {
		return
	}
	d.anim.elapse += diff
	if d.anim.elapse < d.anim.interval {
		return
	}
	d.anim.count--
	if d.anim.count == 0 {
		d.animEnd()
		return
	}
	if d.anim.animType == animTypeScroll {
		d.scrollUpdate()
	}
}

func (d *ModDisplay) scrollUpdate() {

}

func (d *ModDisplay) animEnd() {

}

func (d *ModDisplay) render() {
	for y := 0; y < display_height; y++ {
		d.Clear()
		d.rowPins[y].High()

		for x := 0; x < display_width; x++ {
			idx := d.bufferIndex(x, y)
			if d.buffer[idx] != 0 {
				d.colPins[x].Low()
			}
		}
		time.Sleep(time.Millisecond * 2)
	}
}

func (d *ModDisplay) bufferIndex(x, y int) int {
	x0 := x
	y0 := y
	switch d.rotation {
	case 1:
		x0 = y
		y0 = display_width - 1 - x
	case 2:
		x0 = display_width - 1 - x
		y0 = display_height - 1 - y
	case 3:
		x0 = display_height - 1 - y
		y0 = x
	}
	return y0*display_width + x0
}
