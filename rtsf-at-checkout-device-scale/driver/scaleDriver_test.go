// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// Package driver - This package provides a implementation of a ProtocolDriver interface.
//
package driver

import (
	"device-scale/scale"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
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
		wantErr        bool
	}{
		{
			name:           "valid case",
			scaleConnected: true,
			wantErr:        false,
		},
		{
			name:           "scale not connected",
			scaleConnected: false,
			wantErr:        true,
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
			if tt.wantErr {
				require.Error(t, err)
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
		pid      string
		vid      string
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
			pid:     "6001",
			vid:     "0403",
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
			pid:     "6001",
			vid:     "0403",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ports []*enumerator.PortDetails
			ports = append(ports, &tt.portInfo)
			got, err := findSerialPort(ports, tt.pid, tt.vid)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			if got != tt.want {
				t.Errorf("findSerialPort() = %v, want %v", got, tt.want)
			}
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
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no serial protocol",
			args: args{
				deviceName: "testDeviceName",
				protocols:  nil,
				adminState: "full",
			},
			wantErr: true,
		},
		{
			name: "no pid",
			args: args{
				deviceName: "testDeviceName",
				protocols: map[string]models.ProtocolProperties{
					"serial": {
						"VID": "0403",
					},
				},
				adminState: "full",
			},
			wantErr: true,
		},
		{
			name: "no vid",
			args: args{
				deviceName: "testDeviceName",
				protocols: map[string]models.ProtocolProperties{
					"serial": {
						"PID": "6001",
					},
				},
				adminState: "full",
			},
			wantErr: true,
		},
		{
			name: "port list can not be found",
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
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := getDefaultScaleDriver()
			drv.scaleDevice = &scaleDevice{
				serialDevice: testDevice,
			}

			if err := drv.AddDevice(tt.args.deviceName, tt.args.protocols, tt.args.adminState); (err != nil) != tt.wantErr {
				t.Errorf("ScaleDriver.AddDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
