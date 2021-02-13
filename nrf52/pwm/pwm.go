package pwm

import (
	"device/nrf"
	"machine"
	"runtime/interrupt"
	"runtime/volatile"

	"github.com/wencode/ubit/common"
	"github.com/wencode/ubit/nrf52"
)

const (
	CHANNEL_COUNT     = 4
	PIN_NOT_CONNECTED = 0xFFFFFFFF
	PIN_INVERTED      = 0x80
)

type ID int32

const (
	ID0 ID = iota
	ID1
	ID2
	ID3
)

type Task int32

const (
	TaskStop      Task = 0x4 + 4*iota
	TaskSeqStart0      //0x8
	TaskSeqStart1      //0xc
	TaskNextStep       //0x10
)

type Event int32

const (
	EventStopped      Event = 0x104 + 4*iota
	EventSeqStarted0        //0x108
	EventSeqStarted1        //0x10c
	EventSeqEnd0            //0x110
	EventSeqEnd1            //0x114
	EventPWMPeriodEnd       //0x118
	EventLoopsDone          //0x11c
)

const (
	CLK_16MHz  = nrf.PWM_PRESCALER_PRESCALER_DIV_1
	CLK_8MHz   = nrf.PWM_PRESCALER_PRESCALER_DIV_2
	CLK_4MHz   = nrf.PWM_PRESCALER_PRESCALER_DIV_4
	CLK_2MHz   = nrf.PWM_PRESCALER_PRESCALER_DIV_2
	CLK_1MHz   = nrf.PWM_PRESCALER_PRESCALER_DIV_16
	CLK_500KHz = nrf.PWM_PRESCALER_PRESCALER_DIV_32
	CLK_250KHz = nrf.PWM_PRESCALER_PRESCALER_DIV_64
	CLK_125KHz = nrf.PWM_PRESCALER_PRESCALER_DIV_128
)

var (
	_pwms = [4]PWM{
		{
			PWM_Type: nrf.PWM0,
			id:       -1,
		},
		{
			PWM_Type: nrf.PWM1,
			id:       -1,
		},
		{
			PWM_Type: nrf.PWM2,
			id:       -1,
		},
		{
			PWM_Type: nrf.PWM3,
			id:       -1,
		},
	}
)

type Handler func(event Event, context interface{})

type PWM struct {
	*nrf.PWM_Type
	id      int32
	handler Handler
	context interface{}
	state   volatile.Register32
	flags   uint8

	ir   interrupt.Interrupt
	seq0 *Sequence
	seq1 *Sequence
}

type Config struct {
	output_pins   [CHANNEL_COUNT]machine.Pin
	handler       Handler
	context       interface{}
	base_clock    uint32
	count_mode    uint32
	top_value     uint16
	irq_priority  uint8
	dec_load      uint32
	dec_mode      uint32
	skip_gpio_cfg bool
}

func pwm_defaultConfig() Config {
	return Config{
		output_pins: [CHANNEL_COUNT]machine.Pin{
			machine.NoPin,
			machine.NoPin,
			machine.NoPin,
			machine.NoPin,
		},
		irq_priority:  7,
		base_clock:    CLK_1MHz,
		count_mode:    nrf.PWM_MODE_UPDOWN_Up,
		top_value:     255,
		dec_load:      nrf.PWM_DECODER_LOAD_Common,
		dec_mode:      nrf.PWM_DECODER_MODE_RefreshCount,
		skip_gpio_cfg: false,
	}
}

type Option func(*Config)

func WithOutputPin(pins ...machine.Pin) Option {
	return func(cfg *Config) {
		pins_len := len(pins)
		if pins_len > CHANNEL_COUNT {
			pins_len = CHANNEL_COUNT
		}
		for i := 0; i < pins_len; i++ {
			cfg.output_pins[i] = pins[i]
		}
	}
}

func WithHandler(handler Handler, context interface{}) Option {
	return func(cfg *Config) {
		cfg.handler = handler
		cfg.context = context
	}
}

func WithIRQ_Priority(priority uint8) Option {
	return func(cfg *Config) {
		cfg.irq_priority = priority
	}
}

func WithBaseCLK(clk uint32) Option {
	return func(cfg *Config) {
		cfg.base_clock = clk
	}
}

func Init(id ID, opts ...Option) (*PWM, error) {
	pwm := &(_pwms[id])
	if pwm.id != -1 {
		return nil, common.ErrInvalidState
	}
	cfg := pwm_defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	pwm.handler = cfg.handler
	pwm.context = cfg.context

	pwm_configurePins(pwm.PWM_Type, &cfg)

	pwm.ENABLE.Set(nrf.PWM_ENABLE_ENABLE_Enabled << nrf.PWM_ENABLE_ENABLE_Pos)
	pwm.PRESCALER.Set(cfg.base_clock)
	pwm.MODE.Set(cfg.count_mode)
	pwm.COUNTERTOP.Set(uint32(cfg.top_value))

	pwm.DECODER.Set((cfg.dec_load << nrf.PWM_DECODER_LOAD_Pos) |
		(cfg.dec_mode << nrf.PWM_DECODER_MODE_Pos))

	pwm.SHORTS.Set(0)
	pwm.INTEN.Set(0)
	pwm.EVENTS_LOOPSDONE.Set(0)
	pwm.EVENTS_SEQEND[0].Set(0)
	pwm.EVENTS_SEQEND[1].Set(0)
	pwm.EVENTS_STOPPED.Set(0)

	if cfg.handler != nil {
		var ir interrupt.Interrupt
		switch id {
		case ID0:
			ir = interrupt.New(nrf.IRQ_PWM0, func(_ir interrupt.Interrupt) {
				_pwms[0].irqHandler(_ir)
			})
		case ID1:
			ir = interrupt.New(nrf.IRQ_PWM1, func(_ir interrupt.Interrupt) {
				_pwms[1].irqHandler(_ir)
			})
		case ID2:
			ir = interrupt.New(nrf.IRQ_PWM2, func(_ir interrupt.Interrupt) {
				_pwms[2].irqHandler(_ir)
			})
		case ID3:
			ir = interrupt.New(nrf.IRQ_PWM3, func(_ir interrupt.Interrupt) {
				_pwms[3].irqHandler(_ir)
			})
		}
		ir.SetPriority(cfg.irq_priority)
		ir.Enable()

		pwm.ir = ir
	}

	pwm.state.Set(nrf52.DriverInitialized)
	pwm.id = int32(id)
	return pwm, nil
}

func (p *PWM) Uninit() {
	if p.handler != nil {
		p.ir.Disable()
		p.handler = nil
	}

	p.ENABLE.Set(nrf.PWM_ENABLE_ENABLE_Disabled << nrf.PWM_ENABLE_ENABLE_Pos)
	pwm_deconfigurePins(p.PWM_Type)

	p.state.Set(nrf52.DriverUninitialized)
}

func (p *PWM) SimplePlayback(seq *Sequence, playback_count uint16) {
	seq.fillto(p.PWM_Type, 0)
	seq.fillto(p.PWM_Type, 1)
	p.seq0 = seq
	p.seq1 = seq
	odd := (playback_count&1 == 1)
	p.LOOP.Set(uint32(playback_count))

	shorts_mask := uint32(0)
	if playback_count > 1 {
		if odd {
			shorts_mask = nrf.PWM_SHORTS_LOOPSDONE_SEQSTART1_Msk
		} else {
			shorts_mask = nrf.PWM_SHORTS_LOOPSDONE_SEQSTART0_Msk
		}
	} else {
		shorts_mask = nrf.PWM_SHORTS_LOOPSDONE_STOP_Msk
	}
	p.SHORTS.Set(shorts_mask)

	task := TaskSeqStart0
	if odd {
		task = TaskSeqStart1
	}
	p.startPlayback(task)
}

func (p *PWM) Playback(seq0, seq1 *Sequence, playback_count uint16) {
	seq0.fillto(p.PWM_Type, 0)
	seq1.fillto(p.PWM_Type, 1)
	p.seq0 = seq0
	p.seq1 = seq1
	p.LOOP.Set(uint32(playback_count))

	shorts_mask := uint32(0)
	if playback_count > 1 {
		shorts_mask = nrf.PWM_SHORTS_LOOPSDONE_SEQSTART0_Msk
	} else {
		shorts_mask = nrf.PWM_SHORTS_LOOPSDONE_STOP_Msk
	}
	p.SHORTS.Set(shorts_mask)

	p.startPlayback(TaskSeqStart0)
}

func (p *PWM) Stop(wait_until_stopped bool) bool {
	// Deactivate shortcuts before triggering the STOP task, otherwise the PWM
	// could be immediately started again if the LOOPSDONE event occurred in
	// the same peripheral clock cycle as the STOP task was triggered.
	p.SHORTS.Set(0)

	pwm_taskTrigger(p.PWM_Type, TaskStop)

	if p.IsStopped() {
		return true
	}

	for wait_until_stopped {
		if p.IsStopped() {
			break
		}
	}
	return true
}

func (p *PWM) IsStopped() bool {
	if p.state.Get() != nrf52.DriverStatePoweredOn {
		return true
	}
	if p.EVENTS_STOPPED.Get() == 0 {
		return false
	}
	p.state.Set(nrf52.DriverInitialized)
	return true
}

func (p *PWM) irqHandler(ir interrupt.Interrupt) {
	if e := common.Volatile32_GetAndClear(&p.EVENTS_SEQEND[0]); e != 0 {
		if p.handler != nil {
			p.handler(EventSeqEnd0, p.context)
		}
	}
	if e := common.Volatile32_GetAndClear(&p.EVENTS_SEQEND[1]); e != 0 {
		if p.handler != nil {
			p.handler(EventSeqEnd1, p.context)
		}
	}
	if e := common.Volatile32_GetAndClear(&p.EVENTS_LOOPSDONE); e != 0 {
		if p.handler != nil {
			p.handler(EventLoopsDone, p.context)
		}
	}
	if e := common.Volatile32_GetAndClear(&p.EVENTS_STOPPED); e != 0 {
		p.state.Set(nrf52.DriverInitialized)
		if p.handler != nil {
			p.handler(EventStopped, p.context)
		}
	}
}

func (p *PWM) startPlayback(starting_task Task) {
	p.state.Set(nrf52.DriverStatePoweredOn)
	p.EVENTS_STOPPED.Set(0)
	pwm_taskTrigger(p.PWM_Type, starting_task)
}

func pwm_configurePins(p *nrf.PWM_Type, cfg *Config) {
	for i := 0; i < CHANNEL_COUNT; i++ {
		pin := uint32(PIN_NOT_CONNECTED)
		if output_pin := cfg.output_pins[i]; output_pin != machine.NoPin {
			pin = uint32(output_pin) & (^uint32(PIN_INVERTED))

			if !cfg.skip_gpio_cfg {
				inverted := (output_pin&PIN_INVERTED != 0)
				port, pin_number := nrf52.GetPortPin(machine.Pin(pin))
				if inverted {
					port.OUTSET.Set(uint32(1) << pin_number)
				} else {
					port.OUTCLR.Set(uint32(1) << pin_number)
				}
				machine.Pin(pin).Configure(machine.PinConfig{machine.PinOutput})
			}

		}
		p.PSEL.OUT[i].Set(pin)
	}
}

func pwm_deconfigurePins(p *nrf.PWM_Type) {
	for i := 0; i < CHANNEL_COUNT; i++ {
		output_in := p.PSEL.OUT[i].Get()
		if output_in != PIN_NOT_CONNECTED {
			nrf52.GPIO_Cfg_Default(output_in)
		}
	}
}

func pwm_taskTrigger(p *nrf.PWM_Type, task Task) {
	switch task {
	case TaskStop:
		p.TASKS_STOP.Set(1)
	case TaskSeqStart0:
		p.TASKS_SEQSTART[0].Set(1)
	case TaskSeqStart1:
		p.TASKS_SEQSTART[1].Set(1)
	case TaskNextStep:
		p.TASKS_NEXTSTEP.Set(1)
	}
}
