// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// Package driver - This package provides a implementation of a ProtocolDriver interface.
//
package driver

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	device "github.com/edgexfoundry/device-sdk-go/v2/pkg/service"
	"go.bug.st/serial.v1/enumerator"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// ScaleDriver the driver for a collection of scales
type ScaleDriver struct {
	lc             logger.LoggingClient
	asyncCh        chan<- *dsModels.AsyncValues
	scaleDevice    *scaleDevice
	httpErrors     chan error
	scaleConnected bool
}

var once sync.Once
var driver *ScaleDriver

// NewScaleDeviceDriver instantiates a scale driver
func NewScaleDeviceDriver() dsModels.ProtocolDriver {
	once.Do(func() {
		driver = new(ScaleDriver)
	})
	return driver
}

// DisconnectDevice disconnect from device
func (drv *ScaleDriver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	return nil
}

// Initialize initialize device
func (drv *ScaleDriver) Initialize(lc logger.LoggingClient, asyncCh chan<- *dsModels.AsyncValues, deviceCh chan<- []dsModels.DiscoveredDevice) error {

	config := device.DriverConfigs()

	drv.lc = lc
	drv.asyncCh = asyncCh
	drv.httpErrors = make(chan error, 2)

	serialPort, err := findSerialPort(config["ScalePID"], config["ScaleVID"])

	fmt.Printf("[serialPort]: %v, err: %v", serialPort, err)

	if serialPort == "" || err != nil {
		driver.lc.Warn(err.Error())
		drv.scaleConnected = false
	} else {
		drv.scaleConnected = true
		drv.scaleDevice = newScaleDevice(serialPort)
		driver.lc.Debug(fmt.Sprintf("Connecting to scale: %v", serialPort))

		//
		//
		scaleData, err := drv.scaleDevice.readWeight("scale-item")
		for _, v := range scaleData {
			fmt.Printf("[scaleData]: %v, err: %v", v, err)
		}

		//
		//
	}
	return nil
}

func processScaleData(scaleData map[string]interface{}, deviceResName string) (*dsModels.CommandValue, error) {
	config := device.DriverConfigs()
	scaleData["lane_id"] = config["LaneID"]
	scaleData["scale_id"] = config["ScaleID"]
	scaleData["event_time"] = (time.Now().UnixNano() / 1000000)

	scaleBytes, err := json.Marshal(scaleData)
	if err != nil {
		return nil, err
	}

	commandvalue, err := dsModels.NewCommandValueWithOrigin(
		deviceResName,
		edgexcommon.ValueTypeString,
		string(scaleBytes),
		time.Now().UnixNano()/int64(time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("error on NewCommandValueWithOrigin for %v: %v", deviceResName, err)
	}

	return commandvalue, nil
}

// HandleReadCommands handle AutoEvents
func (drv *ScaleDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {

	res = make([]*dsModels.CommandValue, len(reqs))

	if !drv.scaleConnected {
		return nil, nil
	}

	for i, req := range reqs {
		scaleData, err := drv.scaleDevice.readWeight(req.DeviceResourceName)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				// scale is unplugged or unreachable
				// returning nil prevents the logger from spamming the logs everytime an auto-event fires
				return nil, nil
			}
			return nil, err
		}

		if scaleData == nil {
			// Scale is in motion
			return nil, nil
		}

		result, err := processScaleData(scaleData, req.DeviceResourceName)
		if err != nil {
			return nil, err
		}

		drv.lc.Info(fmt.Sprintf("Scale Reading: %s", result))
		res[i] = result
	}

	return res, nil
}

// HandleWriteCommands handle incoming write commands
func (drv *ScaleDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	return nil
}

// Stop stop a device
func (drv *ScaleDriver) Stop(force bool) error {
	return nil
}

func findSerialPort(pid string, vid string) (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return "", err
	}

	for _, port := range ports {

		if port.IsUSB {
			if port.PID == pid && port.VID == vid {
				return port.Name, nil
			}
		}
	}
	return "", fmt.Errorf("Serial device with pid:vid %s:%s not found", pid, vid)
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (drv *ScaleDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Nothing required to do for AddDevice since new devices will be available
	// when data is posted to REST endpoint
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (drv *ScaleDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Nothing required to do for UpdateDevice since device update will be available
	// when data is posted to REST endpoint.
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (drv *ScaleDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	// Nothing required to do for RemoveDevice since removed device will no longer be available
	// when data is posted to REST endpoint.
	return nil
}
