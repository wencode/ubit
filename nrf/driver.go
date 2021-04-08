package nrf

import (
	"unsafe"

	"device/nrf"
	"machine"
)

const (
	DriverUninitialized = iota
	DriverInitialized
	DriverStatePoweredOn
)

const (
	DefaultPinMode = machine.PinMode(nrf.GPIO_PIN_CNF_DIR_Input |
		nrf.GPIO_PIN_CNF_INPUT_Disconnect |
		nrf.GPIO_PIN_CNF_PULL_Disabled |
		nrf.GPIO_PIN_CNF_DRIVE_S0S1 |
		nrf.GPIO_PIN_CNF_SENSE_Disabled)
)

func IRQ_Number(peripherals unsafe.Pointer) uint8 {
	return uint8(uintptr(peripherals) >> 12)
}

func GetPortPin(p machine.Pin) (*nrf.GPIO_Type, uint32) {
	if p >= 32 {
		return nrf.P1, uint32(p - 32)
	} else {
		return nrf.P0, uint32(p)
	}
}

func GPIO_Cfg_Default(pin_number uint32) {
	pin := machine.Pin(pin_number)
	pin.Configure(machine.PinConfig{DefaultPinMode})
}
