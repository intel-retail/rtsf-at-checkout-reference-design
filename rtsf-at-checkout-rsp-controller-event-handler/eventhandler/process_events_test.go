// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package eventhandler

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/stretchr/testify/assert"
)

func TestTransformRspControllerEventToRfidRoiEvent_SingleReadingMovedEvent(t *testing.T) {
	lc := logger.MockLogger{}

	reading := dtos.NewObjectReading(
		"",
		"rsp-controller",
		"inventory_event",
		RspControllerEvent{
			LaneId:  "123",
			JsonRpc: "2.0",
			Method:  "inventory_event",
			Params: RspControllerEventParams{
				SentOn:    1562972496852,
				GatewayId: "",
				Data: []RspControllerEventParamsData{
					{
						EPCCode:         "30143639F8419145BEEF0009",
						EPCEncodeFormat: "tbd",
						EventType:       "moved",
						FacilityId:      "Entrance",
						Location:        "RSP-15000",
						TimeStamp:       1562972496848,
					},
				},
			},
		},
	)

	rfidEvent, err := transformRspControllerEventToRfidRoiEvent(reading, lc)
	if err != nil {
		t.Fatalf("[FAIL] Error: %s", err.Error())
	}
	assert.Equal(t, 0, len(rfidEvent))
}

func TestTransformRspControllerEventToRfidRoiEvent_MultiReadingsMovedEvent(t *testing.T) {
	lc := logger.MockLogger{}

	reading := dtos.NewObjectReading(
		"",
		"rsp-controller",
		"inventory_event",
		RspControllerEvent{
			LaneId:  "123",
			JsonRpc: "2.0",
			Method:  "inventory_event",
			Params: RspControllerEventParams{
				SentOn:    1562972496852,
				GatewayId: "",
				Data: []RspControllerEventParamsData{
					{
						EPCCode:         "30143639F8419145BEEF0009",
						EPCEncodeFormat: "tbd",
						EventType:       "moved",
						FacilityId:      "Entrance",
						Location:        "RSP-15000",
						TimeStamp:       1562972496848,
					},
					{
						EPCCode:         "30143639F8419145BEEF0009",
						EPCEncodeFormat: "tbd",
						EventType:       "moved",
						FacilityId:      "Entrance",
						Location:        "RSP-15001",
						TimeStamp:       1562972496849,
					},
				},
			},
		},
	)

	rfidEvent, err := transformRspControllerEventToRfidRoiEvent(reading, lc)
	if err != nil {
		t.Fatalf("[FAIL] Error: %s", err.Error())
	}
	assert.Equal(t, 0, len(rfidEvent))
}

func TestTransformRspControllerEventToRfidRoiEvent_4ReadingsArrivalDepartedEvent(t *testing.T) {
	lc := logger.MockLogger{}

	reading := dtos.NewObjectReading(
		"",
		"rsp-controller",
		"inventory_event",
		RspControllerEvent{
			LaneId:  "123",
			JsonRpc: "2.0",
			Method:  "inventory_event",
			Params: RspControllerEventParams{
				SentOn:    1562972496852,
				GatewayId: "",
				Data: []RspControllerEventParamsData{
					{
						EPCCode:         "300C00000000000000000062",
						EPCEncodeFormat: "tbd",
						EventType:       "departed",
						FacilityId:      "Entrance",
						Location:        "RSP-958769-0",
						TimeStamp:       1565132918035,
					},
					{
						EPCCode:         "300C00000000000000000062",
						EPCEncodeFormat: "tbd",
						EventType:       "arrival",
						FacilityId:      "Bagging",
						Location:        "RSP-95a996-0",
						TimeStamp:       1565132919134,
					},
					{
						EPCCode:         "300C00000000000000000060",
						EPCEncodeFormat: "tbd",
						EventType:       "departed",
						FacilityId:      "Entrance",
						Location:        "RSP-958769-0",
						TimeStamp:       1565132909636,
					},
					{
						EPCCode:         "300C00000000000000000060",
						EPCEncodeFormat: "tbd",
						EventType:       "arrival",
						FacilityId:      "Bagging",
						Location:        "RSP-95a996-0",
						TimeStamp:       1565132919176,
					},
				},
			},
		},
	)

	rfidEvent, err := transformRspControllerEventToRfidRoiEvent(reading, lc)
	if err != nil {
		t.Fatalf("[FAIL] Error: %s", err.Error())
	}
	assert.Equal(t, 4, len(rfidEvent))
	assert.Equal(t, "Entrance", rfidEvent[0].ROIName)
	assert.Equal(t, "Bagging", rfidEvent[1].ROIName)
}
