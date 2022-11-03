// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"fmt"
	"reflect"
)

type ServiceConfig struct {
	LossDetector LossDetectorConfig
}

type LossDetectorConfig struct {
	NotificationEmailAddresses string
	NotificationName           string
	NotificationReceiver       string
	SecurityCategory           string
	SubscriptionAdminState     string
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
func (ldconfig *LossDetectorConfig) Validate() error {
	config := reflect.ValueOf(*ldconfig)
	configType := config.Type()

	for i := 0; i < config.NumField(); i++ {
		field := config.Field(i).Interface()
		fieldName := configType.Field(i).Name

		if _, ok := field.(string); ok && len(field.(string)) == 0 {
			return fmt.Errorf("%v is empty", fieldName)
		}
	}
	return nil
}
