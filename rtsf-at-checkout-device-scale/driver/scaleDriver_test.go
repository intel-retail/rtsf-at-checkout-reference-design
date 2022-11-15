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

func TestScaleDriver_processScaleData(t *testing.T) {

	type fields struct {
		lc             logger.LoggingClient
		asyncCh        chan<- *dsModels.AsyncValues
		scaleDevice    *scaleDevice
		httpErrors     chan error
		scaleConnected bool
		config         map[string]string
	}
	type args struct {
		scaleData     map[string]interface{}
		deviceResName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				scaleDevice:    nil,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
			args: args{
				scaleData:     map[string]interface{}{"status": "OK", "total": 2.494, "units": "LB"},
				deviceResName: "testDeviceResource",
			},
			wantErr: false,
		},
		{
			name: "nil scaleData",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				scaleDevice:    nil,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
			args: args{
				scaleData:     nil,
				deviceResName: "testDeviceResource",
			},
			wantErr: true,
		},
		{
			name: "empty device resource",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				scaleDevice:    nil,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
			args: args{
				scaleData:     map[string]interface{}{"status": "OK", "total": 2.494, "units": "LB"},
				deviceResName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := &ScaleDriver{
				lc:             tt.fields.lc,
				asyncCh:        tt.fields.asyncCh,
				scaleDevice:    tt.fields.scaleDevice,
				httpErrors:     tt.fields.httpErrors,
				scaleConnected: tt.fields.scaleConnected,
				config:         tt.fields.config,
			}
			got, err := drv.processScaleData(tt.args.scaleData, tt.args.deviceResName)
			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
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

	type fields struct {
		lc             logger.LoggingClient
		asyncCh        chan<- *dsModels.AsyncValues
		serialDevice   *scale.MockDevice
		httpErrors     chan error
		scaleConnected bool
		config         map[string]string
	}
	type args struct {
		deviceName string
		protocols  map[string]models.ProtocolProperties
		reqs       []dsModels.CommandRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid case",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				serialDevice:   testDevice,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
			args: args{
				deviceName: "testDeviceName",
				protocols:  nil,
				reqs:       nil,
			},
			wantErr: false,
		},
		{
			name: "scale not connected",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				serialDevice:   testDevice,
				httpErrors:     nil,
				scaleConnected: false,
				config:         getDefaultDriverConfig(),
			},
			args: args{
				deviceName: "testDeviceName",
				protocols:  nil,
				reqs:       nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drv := &ScaleDriver{
				lc:      tt.fields.lc,
				asyncCh: tt.fields.asyncCh,
				scaleDevice: &scaleDevice{
					serialDevice: tt.fields.serialDevice,
				},
				httpErrors:     tt.fields.httpErrors,
				scaleConnected: tt.fields.scaleConnected,
				config:         tt.fields.config,
			}
			gotRes, err := drv.HandleReadCommands(tt.args.deviceName,
				map[string]models.ProtocolProperties{},
				[]dsModels.CommandRequest{
					{
						DeviceResourceName: "testDeviceResourceName",
					},
				},
			)
			if tt.wantErr {
				require.NotNil(t, err)
				return
			}
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
				VID:          "0403",
				PID:          "6001",
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
				VID:          "0000",
				PID:          "0000",
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
				require.NotNil(t, err)
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

	type fields struct {
		lc             logger.LoggingClient
		asyncCh        chan<- *dsModels.AsyncValues
		serialDevice   *scale.MockDevice
		httpErrors     chan error
		scaleConnected bool
		config         map[string]string
	}
	type args struct {
		deviceName string
		protocols  map[string]models.ProtocolProperties
		adminState models.AdminState
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "no serial protocol",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				serialDevice:   testDevice,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
			args: args{
				deviceName: "testDeviceName",
				protocols:  nil,
				adminState: "full",
			},
			wantErr: true,
		},
		{
			name: "no pid",
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				serialDevice:   testDevice,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
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
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				serialDevice:   testDevice,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
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
			fields: fields{
				lc:             logger.NewMockClient(),
				asyncCh:        make(chan<- *dsModels.AsyncValues, 16),
				serialDevice:   testDevice,
				httpErrors:     nil,
				scaleConnected: true,
				config:         getDefaultDriverConfig(),
			},
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
			drv := &ScaleDriver{
				lc:      tt.fields.lc,
				asyncCh: tt.fields.asyncCh,
				scaleDevice: &scaleDevice{
					serialDevice: tt.fields.serialDevice,
				},
				httpErrors:     tt.fields.httpErrors,
				scaleConnected: tt.fields.scaleConnected,
				config:         tt.fields.config,
			}
			if err := drv.AddDevice(tt.args.deviceName, tt.args.protocols, tt.args.adminState); (err != nil) != tt.wantErr {
				t.Errorf("ScaleDriver.AddDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
