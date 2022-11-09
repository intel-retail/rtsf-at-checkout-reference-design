// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

// ServiceConfig a struct that wraps CustomConfig which holds the values
// for driver configuration
type serviceConfig struct {
	driverConfig config
}

// config holds the configuration options for device-scale
type config struct {
	SimulatorPort string
	ScaleID       string
	LaneID        string
	TimeOutMilli  string
}

// UpdateFromRaw updates the service's full configuration from raw data
// received from the service provider.
func (c *serviceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*serviceConfig)
	if !ok {
		return false
	}

	*c = *configuration
	return true
}
