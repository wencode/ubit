package ubit

type ModAudio struct {
	is_playing bool
}

func NewModAudio() *ModAudio {
	return &ModAudio{}
}

func (m *ModAudio) Play(source string) error {
	return nil
}

func (m *ModAudio) IsPlaying() bool { return m.is_playing }

func (m *ModAudio) Stop() {}
