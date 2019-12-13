// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"device-scale/driver"

	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
)

const (
	version     string = "1.0"
	serviceName string = "device-scale"
)

func main() {
	d := driver.NewScaleDeviceDriver()
	startup.Bootstrap(serviceName, version, d)
}
