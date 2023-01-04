// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package driver

import (
	"strconv"

	"device-scale/scale"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

type scaleDevice struct {
	serialDevice scale.SerialDevice
}

// readWeight gets called by the auto event to read from the physical scale
// the data read from the scale is wrapped and put on the bus
func (device *scaleDevice) readWeight() (map[string]interface{}, error) {

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

func newScaleDevice(serialPort string, lc logger.LoggingClient, config map[string]string) *scaleDevice {

	lc.Debug("Creating new scale device")
	if config == nil {
		lc.Error("config is nil")
		return nil
	}

	timeout, err := strconv.ParseInt(config["TimeOutMilli"], 10, 64)
	if err != nil {
		lc.Warnf("error on parse TimeOutMilli from config: %v", err)
		lc.Info("set TimeOutMilli to default value 500")
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
