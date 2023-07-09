package exporter

import (
	"context"
	"errors"

	"github.com/rigado/ble"
	"go.uber.org/zap"
)

var EnvironmentalSensingUUID = ble.UUID16(0x181A)

type AdvertisementFilter interface {
	Allowed(ctx context.Context, adv ble.Advertisement) bool
}

type MeasurementReporter interface {
	ReportMeasurement(ctx context.Context, measurement *Measurement) error
}

type Scanner struct {
	device   ble.Device
	filter   AdvertisementFilter
	reporter MeasurementReporter

	logger *zap.Logger
}

func NewScanner(device ble.Device, reporter MeasurementReporter, filter AdvertisementFilter, logger *zap.Logger) *Scanner {
	return &Scanner{
		device:   device,
		filter:   filter,
		reporter: reporter,
		logger:   logger,
	}
}

func (s *Scanner) Run(ctx context.Context) error {
	err := s.device.Scan(ctx, true, func(adv ble.Advertisement) {
		if adv.LocalName() == "" {
			return
		}

		if !s.filter.Allowed(ctx, adv) {
			return
		}

		for _, d := range adv.ServiceData() {
			if !d.UUID.Equal(EnvironmentalSensingUUID) {
				continue
			}

			s.logger.Info(
				"Received an update from a device",
				zap.String("device_name", adv.LocalName()),
				zap.String("device_address", adv.Addr().String()),
			)

			reading, err := unmarshalMeasurement(adv, d)
			if err != nil {
				s.logger.Error("Could not read device data", zap.Error(err))
				continue
			}

			if err = s.reporter.ReportMeasurement(ctx, reading); err != nil {
				s.logger.Error("Could not report reading", zap.Error(err))
				continue
			}
		}
	})

	if !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}
