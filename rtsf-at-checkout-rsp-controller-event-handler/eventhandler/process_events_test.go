// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package eventhandler

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) func(t *testing.T) {
	oldLoggingClient := loggingClient
	loggingClient = logger.NewClient("rsp-controller-event-handler", false, "./test.log", models.InfoLog)

	return func(t *testing.T) {
		loggingClient = oldLoggingClient
	}
}

func TestTransformRspControllerEventToRfidRoiEvent_SingleReadingMovedEvent(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)

	reading := models.Reading{
		Id:     "20a8c57b-f353-448c-92c6-5b4eeae213a5",
		Origin: 1564611885582447220,
		Device: "rsp-controller",
		Name:   "inventory_event",
		Value:  "{\"lane_id\":\"123\",\"jsonrpc\":\"2.0\",\"method\":\"inventory_event\",\"params\":{\"data\":[{\"epc_code\":\"30143639F8419145BEEF0009\",\"epc_encode_format\":\"tbd\",\"event_type\":\"moved\",\"facility_id\":\"Entrance\",\"location\":\"RSP-15000\",\"tid\":null,\"timestamp\":1562972496848}],\"gateway_id\":\"\",\"sent_on\":1562972496852}}",
	}

	rfidEvent, err := transformRspControllerEventToRfidRoiEvent(reading)
	if err != nil {
		t.Fatalf("[FAIL] Error: %s", err.Error())
	}
	assert.Equal(t, 0, len(rfidEvent))
}

func TestTransformRspControllerEventToRfidRoiEvent_MultiReadingsMovedEvent(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)

	reading := models.Reading{
		Id:     "20a8c57b-f353-448c-92c6-5b4eeae213a5",
		Origin: 1564611885582447220,
		Device: "rsp-controller",
		Name:   "inventory_event",
		Value:  "{\"lane_id\":\"123\",\"jsonrpc\":\"2.0\",\"method\":\"inventory_event\",\"params\":{\"data\":[{\"epc_code\":\"30143639F8419145BEEF0009\",\"epc_encode_format\":\"tbd\",\"event_type\":\"moved\",\"facility_id\":\"Entrance\",\"location\":\"RSP-15000\",\"tid\":null,\"timestamp\":1562972496848},{\"epc_code\":\"30143639F8419145BEEF0009\",\"epc_encode_format\":\"tbd\",\"event_type\":\"moved\",\"facility_id\":\"Entrance\",\"location\":\"RSP-15001\",\"tid\":null,\"timestamp\":1562972496849}],\"gateway_id\":\"\",\"sent_on\":1562972496852}}",
	}

	rfidEvent, err := transformRspControllerEventToRfidRoiEvent(reading)
	if err != nil {
		t.Fatalf("[FAIL] Error: %s", err.Error())
	}
	assert.Equal(t, 0, len(rfidEvent))
}

func TestTransformRspControllerEventToRfidRoiEvent_4ReadingsArrivalDepartedEvent(t *testing.T) {
	tearDown := setupTest(t)
	defer tearDown(t)

	reading := models.Reading{
		Id:     "20a8c57b-f353-448c-92c6-5b4eeae213a5",
		Origin: 1564611885582447220,
		Device: "rsp-controller",
		Name:   "inventory_event",
		Value:  "{\"jsonrpc\":\"2.0\",\"lane_id\":\"123\",\"method\":\"inventory_event\",\"params\":{\"data\":[{\"epc_code\":\"300C00000000000000000062\",\"epc_encode_format\":\"tbd\",\"event_type\":\"departed\",\"facility_id\":\"Entrance\",\"location\":\"RSP-958769-0\",\"tid\":null,\"timestamp\":1565132918035},{\"epc_code\":\"300C00000000000000000062\",\"epc_encode_format\":\"tbd\",\"event_type\":\"arrival\",\"facility_id\":\"Bagging\",\"location\":\"RSP-95a996-0\",\"tid\":null,\"timestamp\":1565132919134},{\"epc_code\":\"300C00000000000000000060\",\"epc_encode_format\":\"tbd\",\"event_type\":\"departed\",\"facility_id\":\"Entrance\",\"location\":\"RSP-958769-0\",\"tid\":null,\"timestamp\":1565132909636},{\"epc_code\":\"300C00000000000000000060\",\"epc_encode_format\":\"tbd\",\"event_type\":\"arrival\",\"facility_id\":\"Bagging\",\"location\":\"RSP-95a996-0\",\"tid\":null,\"timestamp\":1565132919176}],\"gateway_id\":\"tpm-NUC7i5DNHE\",\"sent_on\":1565132919460}}",
	}

	rfidEvent, err := transformRspControllerEventToRfidRoiEvent(reading)
	if err != nil {
		t.Fatalf("[FAIL] Error: %s", err.Error())
	}
	assert.Equal(t, 4, len(rfidEvent))
	assert.Equal(t, "Entrance", rfidEvent[0].ROIName)
	assert.Equal(t, "Bagging", rfidEvent[1].ROIName)
}
