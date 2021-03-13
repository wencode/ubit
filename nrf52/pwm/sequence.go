package pwm

import (
	"reflect"
	"unsafe"

	"device/nrf"

	"github.com/wencode/ubit/common"
)

// Sequence defining a sequence of PWM duty cycles
// see: nrf-sdk/include/nrf_pwm.h
// When the sequence is set, the provided duty cycle values are not copied.
// The valuses pointer is stored in the internal register of the peripheral,
// and the values are loaded from RAM during the sequence playback.
type Sequence struct {
	values    []uint16
	repeated  uint32
	end_delay uint32
}

func NewSequence(values []uint16) *Sequence {
	return &Sequence{
		values: values,
	}
}

func NewSequenceWithMulitChannel(values ...[]uint16) (*Sequence, error) {
	values_num := len(values)
	if values_num > 4 {
		values_num = 4
	}
	for i := 1; i < values_num; i++ {
		if len(values[i-1]) != len(values[i]) {
			return nil, common.ErrInvalidArgument
		}
	}
	seq := &Sequence{
		values: make([]uint16, 0, len(values[0])*values_num),
	}
	for i := 0; i < values_num; i++ {
		seq.values = append(seq.values, values[i]...)
	}
	return seq, nil
}

func (seq *Sequence) SetRepeated(n int) { seq.repeated = uint32(n) }

func (seq *Sequence) SetEndDelay(v int) { seq.end_delay = uint32(v) }

func (seq *Sequence) setTo(p *nrf.PWM_Type, seq_id int) {
	values_header := (*reflect.SliceHeader)(unsafe.Pointer(&seq.values))
	p.SEQ[seq_id].PTR.Set(uint32(values_header.Data))
	p.SEQ[seq_id].CNT.Set(uint32(values_header.Len))
	p.SEQ[seq_id].REFRESH.Set(seq.repeated)
	p.SEQ[seq_id].ENDDELAY.Set(seq.end_delay)
}
