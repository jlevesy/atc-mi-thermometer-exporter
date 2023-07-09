package exporter_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/rigado/ble"
	"github.com/jlevesy/atc-mi-thermometer-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScanner(t *testing.T) {
	for _, testCase := range []struct {
		desc           string
		filter         exporter.AdvertisementFilter
		advertisements []ble.Advertisement
		wantOutput     string
	}{
		{
			desc:   "reports reading to the registry",
			filter: exporter.AllowAllDevices{},
			advertisements: []ble.Advertisement{
				&stubAdvertisement{
					localName: "bedroom",
					addr:      ble.NewAddr("coucou"),
					serviceData: []ble.ServiceData{
						{
							UUID: exporter.EnvironmentalSensingUUID,
							Data: []byte{169, 148, 32, 56, 193, 164, 227, 10, 243, 17, 127, 12, 100, 36, 4},
						},
					},
				},
			},
			wantOutput: `
				# HELP atc_mi_thermometer_battery_available_percent battery left on the device in %
                                # TYPE atc_mi_thermometer_battery_available_percent gauge
                                atc_mi_thermometer_battery_available_percent{device_address="coucou",device_name="bedroom"} 100
                                # HELP atc_mi_thermometer_battery_voltage_volts Voltage reported by the battery in volt
                                # TYPE atc_mi_thermometer_battery_voltage_volts gauge
                                atc_mi_thermometer_battery_voltage_volts{device_address="coucou",device_name="bedroom"} 3.1990000000000003
                                # HELP atc_mi_thermometer_humidity_percent Humidity reported by the device in percent
                                # TYPE atc_mi_thermometer_humidity_percent gauge
                                atc_mi_thermometer_humidity_percent{device_address="coucou",device_name="bedroom"} 45.95
                                # HELP atc_mi_thermometer_measurements_count Total measurements reported by the device
                                # TYPE atc_mi_thermometer_measurements_count gauge
                                atc_mi_thermometer_measurements_count{device_address="coucou",device_name="bedroom"} 36
                                # HELP atc_mi_thermometer_temperature_celsius_degrees Temperature reported by the device in celsius degrees
                                # TYPE atc_mi_thermometer_temperature_celsius_degrees gauge
                                atc_mi_thermometer_temperature_celsius_degrees{device_address="coucou",device_name="bedroom"} 27.87
				`,
		},
		{
			desc:   "filters unwanted device",
			filter: exporter.MacAddressFilter([]string{"NotAllowd"}),
			advertisements: []ble.Advertisement{
				&stubAdvertisement{
					addr: ble.NewAddr("coucou"),
				},
			},
			wantOutput: "",
		},
	} {
		t.Run(testCase.desc, func(t *testing.T) {
			var (
				ctx      = context.Background()
				registry = prometheus.NewRegistry()
				reporter = exporter.NewPrometheusReporter(
					registry,
					10*time.Millisecond,
					0,
					zap.NewNop(),
				)
				device = stubDevice{
					advertisements: testCase.advertisements,
				}
				scanner = exporter.NewScanner(
					&device,
					reporter,
					testCase.filter,
					zap.NewNop(),
				)
			)

			err := scanner.Run(ctx)
			require.NoError(t, err)

			err = testutil.GatherAndCompare(
				registry,
				bytes.NewBufferString(testCase.wantOutput),
				"atc_mi_thermometer_temperature_celsius_degrees",
				"atc_mi_thermometer_humidity_percent",
				"atc_mi_thermometer_battery_available_percent",
				"atc_mi_thermometer_battery_voltage_volts",
				"atc_mi_thermometer_measurements_count",
			)
			require.NoError(t, err)

			ctx, cancel := context.WithCancel(ctx)
			t.Cleanup(cancel)

			cleanupDone := make(chan struct{})

			go func() {
				err := reporter.CleanInactiveDevices(ctx)
				require.NoError(t, err)
				close(cleanupDone)
			}()

			time.Sleep(100 * time.Millisecond)
			cancel()

			<-cleanupDone
			err = testutil.GatherAndCompare(
				registry,
				bytes.NewBufferString(""),
				"atc_mi_thermometer_temperature_celsius_degrees",
				"atc_mi_thermometer_humidity_percent",
				"atc_mi_thermometer_battery_available_percent",
				"atc_mi_thermometer_battery_voltage_volts",
				"atc_mi_thermometer_measurements_count",
			)
			require.NoError(t, err)

		})
	}
}

type stubAdvertisement struct {
	ble.Advertisement

	addr        ble.Addr
	localName   string
	serviceData []ble.ServiceData
}

func (s *stubAdvertisement) Addr() ble.Addr                 { return s.addr }
func (s *stubAdvertisement) LocalName() string              { return s.localName }
func (s *stubAdvertisement) ServiceData() []ble.ServiceData { return s.serviceData }

type stubDevice struct {
	ble.Device

	advertisements []ble.Advertisement
}

func (s *stubDevice) Scan(_ context.Context, allowDup bool, hdl ble.AdvHandler) error {
	for _, adv := range s.advertisements {
		hdl(adv)
	}

	return nil
}
