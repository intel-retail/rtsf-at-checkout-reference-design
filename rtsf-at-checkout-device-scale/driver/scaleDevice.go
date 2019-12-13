// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

import (
	"strconv"

	sdk "github.com/edgexfoundry/device-sdk-go"

	"device-scale/scale"
)

type scaleDevice struct {
	serialDevice scale.SerialDevice
}

// readWeight gets called by the auto event to read from the physical scale
// the data read from the scale is wrapped and put on the bus
func (device *scaleDevice) readWeight(deviceResourceName string) (map[string]interface{}, error) {

	scaleReading := make(chan scale.Reading)
	readingErr := make(chan error)

	scale.GetScaleReading(device.serialDevice, scaleReading, readingErr)

	select {
	case err := <-readingErr:
		return nil, err
	case reading := <-scaleReading:

		if reading.Status != "OK" {
			return nil, nil
		}

		scaleData := make(map[string]interface{})
		scaleData["status"] = reading.Status
		// convert total as float64
		if totalWeight, err := strconv.ParseFloat(reading.Value, 64); err == nil {
			scaleData["total"] = totalWeight
		} else {
			scaleData["total"] = 0.0
		}
		scaleData["units"] = reading.Unit
		return scaleData, nil
	}
}

func newScaleDevice(serialPort string) *scaleDevice {

	driver.lc.Debug("Creating new scale device")

	config := sdk.DriverConfigs()

	timeout, err := strconv.ParseInt(config["TimeOutMilli"], 10, 64)
	if err == nil {
		timeout = 500
	}

	options := scale.Config{
		PortName:        serialPort,
		BaudRate:        9600,
		DataBits:        7,
		StopBits:        1,
		MinimumReadSize: 1,
		ParityMode:      2,
		TimeOutMilli:    timeout,
	}

	return &scaleDevice{serialDevice: scale.NewScale(options)}
}
