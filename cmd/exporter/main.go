package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jlevesy/atc-mi-thermometer-exporter/ble"
	"github.com/jlevesy/atc-mi-thermometer-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var version = "unknown"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		logLevel       string
		cleanPeriod    time.Duration
		maxUnseen      time.Duration
		allowedDevices stringsVar
		listenAddress  string
		printVersion   bool
	)

	flag.StringVar(&logLevel, "log-level", "info", "Log Level")
	flag.Var(&allowedDevices, "allow-device", "Allowed device mac address, can be repeated")
	flag.DurationVar(&cleanPeriod, "clean-period", time.Minute, "Interval between 2 cleanups")
	flag.DurationVar(&maxUnseen, "max-unseen", 5*time.Minute, "Maximum duration before a device is considered as inactive")
	flag.StringVar(&listenAddress, "listen-address", ":9977", "HTTP Listen address")
	flag.BoolVar(&printVersion, "version", false, "Print version an exit")
	flag.Parse()

	logger := zap.Must(newLogger(logLevel))

	if printVersion {
		logger.Info("atc-pi-thermometer-exporter", zap.String("version", version))
		return 0
	}

	logger.Info(
		"Starting the exporter",
		zap.String("version", version),
		zap.String("log_level", logLevel),
		zap.Duration("clean_period", cleanPeriod),
		zap.Duration("max_unseen", maxUnseen),
		zap.String("listen_address", listenAddress),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	device, err := ble.NewDevice()
	if err != nil {
		logger.Error("Could not open the BLE device", zap.Error(err))
		return 1
	}

	defer func() {
		if err := device.Stop(); err != nil {
			logger.Error("Could not stop the BLE device", zap.Error(err))
		}
	}()

	var filter exporter.AdvertisementFilter = exporter.AllowAllDevices{}

	if len(allowedDevices) > 0 {
		logger.Info("Allowing only devices", zap.Strings("mac_addresses", allowedDevices))
		filter = exporter.MacAddressFilter(allowedDevices)
	}

	registry := prometheus.NewRegistry()
	reporter := exporter.NewPrometheusReporter(
		registry,
		cleanPeriod,
		maxUnseen,
		logger,
	)

	scanner := exporter.NewScanner(device, reporter, filter, logger)

	server := newServer(listenAddress, registry)

	group, groupContext := errgroup.WithContext(ctx)

	group.Go(func() error { return scanner.Run(groupContext) })
	group.Go(func() error { return reporter.CleanInactiveDevices(groupContext) })
	group.Go(func() error {
		shutdownDone := make(chan struct{})

		go func() {
			<-groupContext.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP server reported an error while shutting down, forcing stop", zap.Error(err))

				server.Close()
				return
			}

			close(shutdownDone)
		}()

		logger.Info("HTTP server listening", zap.String("addr", server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		<-shutdownDone

		logger.Info("HTTP server stopped")

		return nil
	})

	if err := group.Wait(); err != nil {
		logger.Error("Exporter reported an error", zap.Error(err))
		return 1
	}

	logger.Info("Exporter received a signal, bye bye...")

	return 0
}

func newServer(addr string, registry *prometheus.Registry) *http.Server {
	var mux http.ServeMux

	mux.Handle("/metrics", promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{Registry: registry},
	))

	srv := http.Server{
		Addr:    addr,
		Handler: &mux,
	}

	return &srv
}

func newLogger(lvl string) (*zap.Logger, error) {
	if lvl == "debug" {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}

type stringsVar []string

func (v *stringsVar) String() string {
	return strings.Join(*v, ".")
}

func (v *stringsVar) Set(value string) error {
	*v = append(*v, value)
	return nil
}
