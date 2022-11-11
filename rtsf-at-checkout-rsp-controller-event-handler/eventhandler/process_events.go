// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package eventhandler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
)

const (
	RFIDEventArrival  = "arrival"
	RFIDEventDeparted = "departed"
	RFIDEventMoved    = "moved"
	ROIActionEntered  = "ENTERED"
	ROIActionExited   = "EXITED"

	resourceName = "rfid-roi"
	sourceName   = "rfid-roi"
	deviceName   = "rfid-roi-rest"
	profileName  = "rfid-roi"
)

// ProcessRspControllerEvents transforms RSP Controller events to an RFID ROI (region or interest), as defined by checkout-event-reconciler
func ProcessRspControllerEvents(edgexcontext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := edgexcontext.LoggingClient()

	if data == nil {
		return false, errors.New("no data received by rsp controller event handler")
	}

	event, ok := data.(dtos.Event)
	if !ok {
		return false, nil
	}

	for _, reading := range event.Readings {
		rfidEvents, err := transformRspControllerEventToRfidRoiEvent(reading, lc)
		if err != nil {
			lc.Errorf("Transform RSP Controller Reading To RFIDEventEntry error: %v", err)
			continue
		}

		for _, rfidEvent := range rfidEvents {
			newEvent := dtos.NewEvent(profileName, deviceName, sourceName)
			newEvent.AddObjectReading(resourceName, &rfidEvent)

			responseData, err := json.Marshal(newEvent)
			if err != nil {
				lc.Errorf("failed to marshal rfid event: %v", err)
				continue
			}

			edgexcontext.SetResponseData(responseData)
			edgexcontext.AddValue(interfaces.PROFILENAME, newEvent.ProfileName)
			edgexcontext.AddValue(interfaces.DEVICENAME, newEvent.DeviceName)
			edgexcontext.AddValue(interfaces.SOURCENAME, newEvent.SourceName)

		}
	}

	return false, nil
}

func transformRspControllerEventToRfidRoiEvent(reading dtos.BaseReading, lc logger.LoggingClient) ([]RFIDEventEntry, error) {
	rfidEvents := []RFIDEventEntry{}
	rspControllerEvent := RspControllerEvent{}

	err := unmarshalObjValue(reading.ObjectReading.ObjectValue, &rspControllerEvent)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling RSP Controller event: %v", err)
	}

	for _, apData := range rspControllerEvent.Params.Data {

		rfidReading := RFIDEventEntry{}
		rfidReading.LaneId = rspControllerEvent.LaneId
		rfidReading.EPC = apData.EPCCode
		rfidReading.ROIName = apData.FacilityId
		rfidReading.ROIAction = apData.EventType
		rfidReading.EventTime = apData.TimeStamp

		switch apData.EventType {
		case RFIDEventArrival:
			rfidReading.ROIAction = ROIActionEntered
		case RFIDEventDeparted:
			rfidReading.ROIAction = ROIActionExited
		case RFIDEventMoved: // ignore moved events - only interested in arrival or departed events
			lc.Debugf("Ignoring RSP Controller moved event: EPC-%v, FacilityId-%v, Location-%v, Timestamp-%v", apData.EPCCode, apData.FacilityId, apData.Location, apData.TimeStamp)
			continue
		default:
			lc.Debugf("Unrecognized RSP Controller event: %v", apData.EventType)
			continue
		}

		rfidEvents = append(rfidEvents, rfidReading)
	}

	return rfidEvents, nil
}
