// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// Package driver - This package provides a implementation of a ProtocolDriver interface.
package driver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"go.bug.st/serial.v1/enumerator"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	edgexcommon "github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

// ScaleDriver the driver for a collection of scales
type ScaleDriver struct {
	lc             logger.LoggingClient
	asyncCh        chan<- *dsModels.AsyncValues
	scaleDevice    *scaleDevice
	httpErrors     chan error
	scaleConnected bool
	config         map[string]string
}

// NewScaleDeviceDriver instantiates a scale driver
func NewScaleDeviceDriver() interfaces.ProtocolDriver {
	return new(ScaleDriver)
}

// DisconnectDevice disconnect from device
func (drv *ScaleDriver) DisconnectDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	return nil
}

// Initialize initialize device
func (drv *ScaleDriver) Initialize(sdk interfaces.DeviceServiceSDK) error {

	drv.lc = sdk.LoggingClient()
	drv.asyncCh = sdk.AsyncValuesChannel()
	drv.httpErrors = make(chan error, 2)
	drv.config = sdk.DriverConfigs()

	return nil
}

func (drv *ScaleDriver) processScaleData(scaleData map[string]interface{}, deviceResName string) (*dsModels.CommandValue, error) {
	if len(scaleData) == 0 {
		return nil, errors.New("scaleData can not be nil")
	}
	if len(deviceResName) == 0 {
		return nil, errors.New("deviceResName can not be empty")
	}
	scaleData["lane_id"] = drv.config["LaneID"]
	scaleData["scale_id"] = drv.config["ScaleID"]
	scaleData["event_time"] = (time.Now().UnixNano() / 1000000)

	scaleBytes, err := json.Marshal(scaleData)
	if err != nil {
		return nil, err
	}

	commandValue, err := dsModels.NewCommandValueWithOrigin(
		deviceResName,
		edgexcommon.ValueTypeString,
		string(scaleBytes),
		time.Now().UnixNano()/int64(time.Millisecond),
	)
	if err != nil {
		return nil, fmt.Errorf("error on NewCommandValueWithOrigin for %v: %v", deviceResName, err)
	}

	return commandValue, nil
}

// HandleReadCommands handle AutoEvents
func (drv *ScaleDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {

	res = make([]*dsModels.CommandValue, len(reqs))

	if !drv.scaleConnected {
		// we need to return nil when scale is not connected for simulator purpose
		// if physical device is not connected, the error will trigger in AddDevice
		drv.lc.Warnf("scale is not connected")
		return nil, nil
	}

	for i, req := range reqs {
		scaleData, err := drv.scaleDevice.readWeight()
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

		result, err := drv.processScaleData(scaleData, req.DeviceResourceName)
		if err != nil {
			return nil, err
		}

		drv.lc.Infof("Scale Reading: %s", result)
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

// Start starts the driver logic
func (drv *ScaleDriver) Start() error {
	return nil
}

func findSerialPort(ports []*enumerator.PortDetails, pid string, vid string) (string, error) {

	for _, port := range ports {

		if port.IsUSB {
			if port.PID == pid && port.VID == vid {
				return port.Name, nil
			}
		}
	}
	return "", fmt.Errorf("serial device with pid:vid %s:%s not found", pid, vid)
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (drv *ScaleDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	// Previously validated by ValidateDevice implicitly called by SDK

	serialProtocol := protocols["serial"]
	pid := serialProtocol["PID"].(string)
	vid := serialProtocol["VID"].(string)

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return err
	}

	serialPort, err := findSerialPort(ports, pid, vid)

	if err != nil {
		drv.lc.Error(err.Error())
		drv.scaleConnected = false
		return fmt.Errorf("unable to find weight scale serial port: %v", err)
	} else {
		drv.lc.Debugf("[serialPort]: %v", serialPort)
		drv.scaleConnected = true
		drv.scaleDevice = newScaleDevice(serialPort, drv.lc, drv.config)
		drv.lc.Debugf("Connecting to scale: %v", serialPort)

		scaleData, err := drv.scaleDevice.readWeight()
		if err != nil {
			return fmt.Errorf("readWeight failed: %v", err)
		}
		for _, v := range scaleData {
			drv.lc.Debugf("[scaleData]: %v", v)
		}
	}

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

// Discover is a callback function that is invoked
// by the SDK but should never be called
func (drv *ScaleDriver) Discover() error {

	return errors.New("discovery not implemented")
}

// Validate is a callback function that is invoked by the SDK
func (drv *ScaleDriver) ValidateDevice(device models.Device) error {

	// Device service protocol validation
	serial, ok := device.Protocols["serial"]
	if !ok {
		return errors.New("protocols missing serial section")
	}

	value, ok := serial["VID"]
	if !ok {
		return errors.New("serial Protocol missing VID setting")
	}
	vid, ok := value.(string)
	if !ok {
		return errors.New("VID value is not a string")
	}
	if len(vid) == 0 {
		return errors.New("VID is empty")
	}

	value, ok = serial["PID"]
	if !ok {
		return errors.New("serial Protocol missing PID setting")
	}
	pid, ok := value.(string)
	if !ok {
		return errors.New("PID value is not a string")
	}
	if len(pid) == 0 {
		return errors.New("PID is empty")
	}
	return nil
}
