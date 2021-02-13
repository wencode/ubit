package font5x5

import (
	"fmt"
	"testing"

	"github.com/wencode/ubit/image5x5"
)

func display(img image5x5.Image) {
	fmt.Println("")
	idx := 0
	for y := 0; y < image5x5.Height; y++ {
		for x := 0; x < image5x5.Width; x++ {
			if img[idx] != 0 {
				fmt.Printf("1 ")
			} else {
				fmt.Printf("0 ")
			}
			idx++
		}
		fmt.Println("")
	}
}

func TestA(t *testing.T) {
	img := GenImage5x5('A', 255)
	display(img)
}
