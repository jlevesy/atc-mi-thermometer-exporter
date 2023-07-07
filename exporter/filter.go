package exporter

import (
	"context"
	"strings"

	"github.com/go-ble/ble"
)

type AllowAllDevices struct{}

func (AllowAllDevices) Allowed(context.Context, ble.Advertisement) bool { return true }

func MacAddressFilter(macs []string) AdvertisementFilter {
	filter := make(macAddressFilter, len(macs))

	for i, vv := range macs {
		filter[i] = strings.ToLower(strings.ReplaceAll(vv, ":", ""))
	}

	return filter
}

type macAddressFilter []string

func (m macAddressFilter) Allowed(_ context.Context, adv ble.Advertisement) bool {
	for _, addr := range m {
		if strings.EqualFold(addr, adv.Addr().String()) {
			return true
		}
	}

	return false
}
