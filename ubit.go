package ubit

var (
	// Display drive the led matrix
	Display *ModDisplay
	// Audio use to play sounds
	Audio *ModAudio
)

func init() {
	Display = NewModDisplay()
	Audio = NewModAudio()

}
