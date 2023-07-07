package exporter

import (
	"encoding/binary"
	"fmt"

	"github.com/go-ble/ble"
)

type Measurement struct {
	Addr           string
	DeviceName     string
	Temperature    float64
	Humidity       float64
	BatteryPercent float64
	BatteryVoltage float64
	Count          float64
}

func unmarshalMeasurement(adv ble.Advertisement, data ble.ServiceData) (*Measurement, error) {
	if len(data.Data) != 15 {
		return nil, errBadFrameLength(len(data.Data))
	}

	deviceName := adv.LocalName()
	if deviceName == "" {
		deviceName = "unknown"
	}

	return &Measurement{
		Addr:           adv.Addr().String(),
		DeviceName:     deviceName,
		Temperature:    float64(binary.LittleEndian.Uint16(data.Data[6:8])) * 0.01,
		Humidity:       float64(binary.LittleEndian.Uint16(data.Data[8:10])) * 0.01,
		BatteryVoltage: float64(binary.LittleEndian.Uint16(data.Data[10:12])) * 0.001,
		BatteryPercent: float64(data.Data[12]),
		Count:          float64(data.Data[13]),
	}, nil
}

type errBadFrameLength int

func (v errBadFrameLength) Error() string {
	return fmt.Sprintf("Expected a length of 15, got %d", v)
}
