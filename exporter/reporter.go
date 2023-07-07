package exporter

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	namespace     = "atc_mi_thermometer"
	deviceAddress = "device_address"
	deviceName    = "device_name"
)

var measurementsDimensions = []string{
	deviceAddress,
	deviceName,
}

type PrometheusReporter struct {
	temperature       *prometheus.GaugeVec
	humidity          *prometheus.GaugeVec
	batteryPercent    *prometheus.GaugeVec
	batteryVoltage    *prometheus.GaugeVec
	measurementsTotal *prometheus.GaugeVec

	cleanPeriod time.Duration
	tracker     *activityTracker
	logger      *zap.Logger
}

func NewPrometheusReporter(reg *prometheus.Registry, cleanPeriod, maxUnseen time.Duration, logger *zap.Logger) *PrometheusReporter {
	reporter := &PrometheusReporter{
		temperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "temperature_celsius_degrees",
			Help:      "Temperature reported by the device in celsius degrees",
		},
			measurementsDimensions,
		),
		humidity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "humidity_percent",
				Help:      "Humidity reported by the device in percent",
			},
			measurementsDimensions,
		),
		batteryPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "battery",
				Name:      "available_percent",
				Help:      "battery left on the device in %",
			},
			measurementsDimensions,
		),
		batteryVoltage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "battery",
				Name:      "voltage_volts",
				Help:      "Voltage reported by the battery in volt",
			},
			measurementsDimensions,
		),
		measurementsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "measurements_count",
				Help:      "Total measurements reported by the device",
			},
			measurementsDimensions,
		),
		logger:      logger,
		cleanPeriod: cleanPeriod,
		tracker: &activityTracker{
			maxUnseen: maxUnseen,
			lastSeen:  make(map[string]time.Time),
			logger:    logger,
		},
	}

	reg.MustRegister(
		reporter.temperature,
		reporter.humidity,
		reporter.batteryPercent,
		reporter.batteryVoltage,
		reporter.measurementsTotal,
	)

	return reporter
}

func (r *PrometheusReporter) ReportMeasurement(_ context.Context, m *Measurement) error {
	r.temperature.WithLabelValues(m.Addr, m.DeviceName).Set(m.Temperature)
	r.humidity.WithLabelValues(m.Addr, m.DeviceName).Set(m.Humidity)
	r.batteryPercent.WithLabelValues(m.Addr, m.DeviceName).Set(m.BatteryPercent)
	r.batteryVoltage.WithLabelValues(m.Addr, m.DeviceName).Set(m.BatteryVoltage)
	r.measurementsTotal.WithLabelValues(m.Addr, m.DeviceName).Set(m.Count)

	r.tracker.checkIn(m.Addr)
	return nil
}

func (r *PrometheusReporter) CleanInactiveDevices(ctx context.Context) error {
	ticker := time.NewTicker(r.cleanPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			r.logger.Info("Removing inactive devices")

			inactiveDevices := r.tracker.listInactiveDevices()

			for _, deviceAddr := range inactiveDevices {
				r.flushDevice(deviceAddr)
				r.tracker.forget(deviceAddr)
			}
		}
	}
}

func (r *PrometheusReporter) flushDevice(addr string) {
	filter := prometheus.Labels{deviceAddress: addr}

	r.temperature.DeletePartialMatch(filter)
	r.humidity.DeletePartialMatch(filter)
	r.batteryPercent.DeletePartialMatch(filter)
	r.batteryVoltage.DeletePartialMatch(filter)
	r.measurementsTotal.DeletePartialMatch(filter)
}

type activityTracker struct {
	lastSeenMu sync.RWMutex
	lastSeen   map[string]time.Time
	maxUnseen  time.Duration

	logger *zap.Logger
}

func (t *activityTracker) listInactiveDevices() []string {
	t.lastSeenMu.RLock()
	defer t.lastSeenMu.RUnlock()

	var result []string

	for addr, lastSeen := range t.lastSeen {
		if time.Since(lastSeen) > t.maxUnseen {
			result = append(result, addr)
		}
	}

	return result
}

func (t *activityTracker) checkIn(addr string) {
	t.lastSeenMu.Lock()
	t.lastSeen[addr] = time.Now()
	t.lastSeenMu.Unlock()
	t.logger.Debug("Device is active", zap.String("device_address", addr))
}

func (t *activityTracker) forget(addr string) {
	t.lastSeenMu.Lock()
	delete(t.lastSeen, addr)
	t.lastSeenMu.Unlock()
	t.logger.Debug("Device is inactive", zap.String("device_address", addr))
}
