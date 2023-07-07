package ble

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

func NewDevice(opts ...ble.Option) (ble.Device, error) {
	return linux.NewDevice()
}
