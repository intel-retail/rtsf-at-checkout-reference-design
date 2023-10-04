// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// Package driver - This package provides a implementation of a ProtocolDriver interface.
package driver

import (
	"device-scale/scale"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bug.st/serial.v1/enumerator"
)

func getDefaultDriverConfig() map[string]string {
	config := make(map[string]string)
	config["SimulatorPort"] = "8081"
	config["ScaleID"] = "123"
	config["LaneID"] = "123"
	config["TimeOutMilli"] = "500"
	return config
}

func getDefaultScaleDriver() ScaleDriver {
	return ScaleDriver{
		lc:             logger.NewMockClient(),
		asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
		scaleDevice:    nil,
		httpErrors:     nil,
		scaleConnected: true,
		config:         getDefaultDriverConfig(),
	}
}

func TestScaleDriver_processScaleData(t *testing.T) {
	type args struct {
		scaleData     map[string]interface{}
		deviceResName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			args: args{
				scaleData:     map[string]interface{}{"status": "OK", "total": 2.494, "units": "LB"},
				deviceResName: "testDeviceResource",
			},
			wantErr: false,
		},
		{
			name: "nil scaleData",
			args: args{
				scaleData:     nil,
				deviceResName: "testDeviceResource",
			},
			wantErr: true,
		},
		{
			name: "empty device resource",
			args: args{
				scaleData:     map[string]interface{}{"status": "OK", "total": 2.494, "units": "LB"},
				deviceResName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := getDefaultScaleDriver()
			got, err := drv.processScaleData(tt.args.scaleData, tt.args.deviceResName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, got)

		})
	}
}

func TestScaleDriver_HandleReadCommands(t *testing.T) {

	config := scale.Config{
		PortName:        "/dev/tty.usbserial-test",
		BaudRate:        9600,
		DataBits:        7,
		StopBits:        1,
		MinimumReadSize: 1,
		ParityMode:      2,
		TimeOutMilli:    500,
	}

	testDevice := scale.InitializeMockDevice(&config)

	tests := []struct {
		name           string
		scaleConnected bool
		wantRes        bool
	}{
		{
			name:           "valid case",
			scaleConnected: true,
			wantRes:        false,
		},
		{
			name:           "scale not connected",
			scaleConnected: false,
			wantRes:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := getDefaultScaleDriver()
			drv.scaleConnected = tt.scaleConnected
			drv.scaleDevice = &scaleDevice{
				serialDevice: testDevice,
			}

			gotRes, err := drv.HandleReadCommands("testDeviceName",
				map[string]models.ProtocolProperties{},
				[]dsModels.CommandRequest{
					{
						DeviceResourceName: "testDeviceResourceName",
					},
				},
			)
			if tt.wantRes {
				require.Nil(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, gotRes)
		})
	}
}

func Test_findSerialPort(t *testing.T) {

	tests := []struct {
		name     string
		portInfo enumerator.PortDetails
		want     string
		wantErr  bool
	}{
		{
			name: "valid case",
			portInfo: enumerator.PortDetails{
				Name:         "testDevice",
				IsUSB:        true,
				PID:          "6001",
				VID:          "0403",
				SerialNumber: "0123456",
			},
			want:    "testDevice",
			wantErr: false,
		},
		{
			name: "vid pid not found",
			portInfo: enumerator.PortDetails{
				Name:         "testDevice",
				IsUSB:        true,
				PID:          "0000",
				VID:          "0000",
				SerialNumber: "0123456",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ports []*enumerator.PortDetails
			ports = append(ports, &tt.portInfo)
			got, err := findSerialPort(ports, "6001", "0403")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestScaleDriver_AddDevice(t *testing.T) {
	config := scale.Config{
		PortName:        "/dev/tty.usbserial-test",
		BaudRate:        9600,
		DataBits:        7,
		StopBits:        1,
		MinimumReadSize: 1,
		ParityMode:      2,
		TimeOutMilli:    500,
	}

	testDevice := scale.InitializeMockDevice(&config)

	type args struct {
		deviceName string
		protocols  map[string]models.ProtocolProperties
		adminState models.AdminState
	}
	tests := []struct {
		name          string
		args          args
		expectedError string
	}{
		{
			// test case is as happy as possible without a serial device
			// due to inability of mocking serial device
			name: "happy path - port list can not be found",
			args: args{
				deviceName: "testDeviceName",
				protocols: map[string]models.ProtocolProperties{
					"serial": {
						"VID": "0403",
						"PID": "6001",
					},
				},
				adminState: "full",
			},
			expectedError: "unable to find weight scale serial port: serial device with pid:vid 6001:0403 not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := getDefaultScaleDriver()
			drv.scaleDevice = &scaleDevice{
				serialDevice: testDevice,
			}

			err := drv.AddDevice(tt.args.deviceName, tt.args.protocols, tt.args.adminState)
			require.Error(t, err)
			assert.Equal(t, tt.expectedError, err.Error())
		})
	}
}

func TestScaleDriver_ValidateDevice(t *testing.T) {
	config := scale.Config{
		PortName:        "/dev/tty.usbserial-test",
		BaudRate:        9600,
		DataBits:        7,
		StopBits:        1,
		MinimumReadSize: 1,
		ParityMode:      2,
		TimeOutMilli:    500,
	}

	testDevice := scale.InitializeMockDevice(&config)

	tests := []struct {
		name          string
		device        models.Device
		expectedError string
	}{
		{
			name: "happy path protocol",
			device: models.Device{
				Name: "testDeviceName",
				Protocols: map[string]models.ProtocolProperties{
					"serial": {
						"VID": "0403",
						"PID": "0600",
					},
				},
			},
			expectedError: "",
		},
		{
			name: "no serial protocol",
			device: models.Device{
				Name: "testDeviceName",
				Protocols: map[string]models.ProtocolProperties{
					"nonSerial": {
						"VID": "0403",
						"PID": "0600",
					},
				},
			},
			expectedError: "protocols missing serial section",
		},
		{
			name: "no pid",
			device: models.Device{
				Name: "testDeviceName",
				Protocols: map[string]models.ProtocolProperties{
					"serial": {
						"VID": "0403",
					},
				},
			},
			expectedError: "serial Protocol missing PID setting",
		}, {
			name: "empty pid",
			device: models.Device{
				Name: "testDeviceName",
				Protocols: map[string]models.ProtocolProperties{
					"serial": {
						"PID": "",
						"VID": "0403",
					},
				},
			},
			expectedError: "PID is empty",
		},
		{
			name: "no vid",
			device: models.Device{
				Name: "testDeviceName",
				Protocols: map[string]models.ProtocolProperties{
					"serial": {
						"PID": "6001",
					},
				},
			},
			expectedError: "serial Protocol missing VID setting",
		},
		{
			name: "empty vid",
			device: models.Device{
				Name: "testDeviceName",
				Protocols: map[string]models.ProtocolProperties{
					"serial": {
						"PID": "6001",
						"VID": "",
					},
				},
			},
			expectedError: "VID is empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := getDefaultScaleDriver()
			drv.scaleDevice = &scaleDevice{
				serialDevice: testDevice,
			}

			err := drv.ValidateDevice(tt.device)
			if len(tt.expectedError) > 0 {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}
