// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"fmt"
	"reflect"
	"time"
)

type ServiceConfig struct {
	Reconciler ReconcilerConfig
}

type ReconcilerConfig struct {
	DeviceNames           string
	DevicePos             string
	DeviceScale           string
	DeviceCV              string
	DeviceRFID            string
	ProductLookupEndpoint string
	WebSocketPort         string
	ScaleToScaleTolerance float64
	CvTimeAlignment       string
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (c *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false
	}

	*c = *configuration

	return true
}

// Validate ensures your custom configuration has proper values.
func (bs *ReconcilerConfig) Validate() (time.Duration, error) {
	config := reflect.ValueOf(*bs)
	configType := config.Type()
	var defaultRtnVal time.Duration = 0

	for i := 0; i < config.NumField(); i++ {
		field := config.Field(i).Interface()
		fieldName := configType.Field(i).Name

		if _, ok := field.(string); ok && len(field.(string)) == 0 {
			return defaultRtnVal, fmt.Errorf("%v is empty", fieldName)
		}
	}

	tempDuration, err := time.ParseDuration(bs.CvTimeAlignment)
	if err != nil {
		return defaultRtnVal, fmt.Errorf("failed to parse cvTimeAlignment duration: %v", err)
	}

	return tempDuration, nil
}
