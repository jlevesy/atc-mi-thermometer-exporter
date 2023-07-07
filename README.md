### ATC_MiThermometer Prometheus Exporter

This repository is the source code for a small Prometheus exporter that exposes temparture, humidity and battery stats
reported by devices runnning the [ATC_MiThermoteter](https://github.com/pvvx/ATC_MiThermometer) firmware from pvvx. (NOT atc1441).

It requires the devices to be flashed with this firmware and configured to report using the [custom](https://github.com/pvvx/ATC_MiThermometer#bluetooth-advertising-formats) bluetooth advertising format.

### Exported Timeseries

```
# HELP atc_mi_thermometer_battery_available_percent battery left on the device in %
# TYPE atc_mi_thermometer_battery_available_percent gauge
atc_mi_thermometer_battery_available_percent{device_address="aa:bb:cc:dd:ee:ff",device_name="livingroom"} 100
# HELP atc_mi_thermometer_battery_voltage_volts Voltage reported by the battery in volt
# TYPE atc_mi_thermometer_battery_voltage_volts gauge
atc_mi_thermometer_battery_voltage_volts{device_address="aa:bb:cc:dd:ee:ff",device_name="livingroom"} 3.13
# HELP atc_mi_thermometer_humidity_percent Humidity reported by the device in percent
# TYPE atc_mi_thermometer_humidity_percent gauge
atc_mi_thermometer_humidity_percent{device_address="aa:bb:cc:dd:ee:ff",device_name="livingroom"} 42.61
# HELP atc_mi_thermometer_measurements_count Total measurements reported by the device
# TYPE atc_mi_thermometer_measurements_count gauge
atc_mi_thermometer_measurements_count{device_address="aa:bb:cc:dd:ee:ff",device_name="livingroom"} 229
# HELP atc_mi_thermometer_temperature_celsius_degrees Temperature reported by the device in celsius degrees
# TYPE atc_mi_thermometer_temperature_celsius_degrees gauge
atc_mi_thermometer_temperature_celsius_degrees{device_address="aa:bb:cc:dd:ee:ff",device_name="livingroom"} 28.94
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
```

### Running the exporter

It comes a self contained binary, either build it yourself from source or grab it from the [releases page](https://github.com/jlevesy/atc-mi-thermometer-exporter/releases)

```
Usage of ./exporter:
  -allow-device value
        Allowed device mac address, can be repeated
  -clean-period duration
        Interval between 2 cleanups (default 1m0s)
  -listen-address string
        HTTP Listen address (default ":9977")
  -log-level string
        Log Level (default "info")
  -max-unseen duration
        Maximum duration before a device is considered as inactive (default 5m0s)
```

You can find an example [systemd unit as well](./config/atc-mi-thermometer-exporter.service)
