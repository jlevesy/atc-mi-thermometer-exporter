package ble

import (
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/darwin"
)

func NewDevice(opts ...ble.Option) (ble.Device, error) {
	return darwin.NewDevice(opts...)
}
