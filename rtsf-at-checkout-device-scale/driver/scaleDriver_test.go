// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

// Package driver - This package provides a implementation of a ProtocolDriver interface.
//
package driver

import (
	"reflect"
	"testing"

	dsModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
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

	// goodCV, err := dsModels.NewCommandValueWithOrigin(
	// 	"testDeviceResource",
	// 	edgexcommon.ValueTypeString,
	// 	string(scaleBytes),
	// 	time.Now().UnixNano()/int64(time.Millisecond),
	// )
	// require.NoError(t, err)

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
		want    *dsModels.CommandValue
		wantErr bool
	}{
		// {
		// 	name: "valid case",
		// 	fields: fields{
		// 		lc: logger.NewMockClient(),
		// 		asyncCh: make(chan<- *dsModels.AsyncValues, 16),
		// 		scaleDevice: nil,

		// 	},
		// 	args: args{
		// 		scaleData: ,
		// 		deviceResName: "testDeviceResource",
		// 	},

		// },
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ScaleDriver.processScaleData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScaleDriver.processScaleData() = %v, want %v", got, tt.want)
			}
		})
	}
}
