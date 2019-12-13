// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package eventhandler

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	RFIDEventArrival  = "arrival"
	RFIDEventDeparted = "departed"
	RFIDEventMoved    = "moved"
	ROIActionEntered  = "ENTERED"
	ROIActionExited   = "EXITED"
)

var loggingClient logger.LoggingClient

// ProcessRspControllerEvents transforms RSP Controller events to an RFID ROI (region or interest), as defined by checkout-event-reconciler
func ProcessRspControllerEvents(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, nil
	}

	loggingClient = edgexcontext.LoggingClient

	result, ok := params[0].(models.Event)
	if !ok {
		loggingClient.Error("No event received by RSP Controller event handler")
		return false, nil
	}

	for _, reading := range result.Readings {
		rfidEvents, err := transformRspControllerEventToRfidRoiEvent(reading) //, edgexcontext.LoggingClient)
		if err != nil {
			loggingClient.Error(fmt.Sprintf("Transform RSP Controller Reading To RFIDEventEntry error: %v\n", err))
			continue
		}

		for _, rfidEvent := range rfidEvents {
			eventBytes, err := json.Marshal(&rfidEvent)
			if err != nil {
				loggingClient.Error(fmt.Sprintf("Error marshaling RFID event to push to CoreData: %v\n", err))
				continue
			}

			result, err := edgexcontext.PushToCoreData("device-rfid-roi-rest", "rfid-roi-event", eventBytes)
			if err != nil {
				loggingClient.Error(fmt.Sprintf("Error pushing RFID event entry to CoreData: %v\n", err))
				continue
			}

			if result != nil {
				loggingClient.Debug(fmt.Sprintf("Pushed RFID event entry to CoreData: EPC-%v, ROIName-%v, ROIAction-%v, EventTime-%v\n", rfidEvent.EPC, rfidEvent.ROIName, rfidEvent.ROIAction, rfidEvent.EventTime))
			}
		}
	}

	return false, nil
}

func transformRspControllerEventToRfidRoiEvent(reading models.Reading) ([]RFIDEventEntry, error) {
	rfidEvents := []RFIDEventEntry{}
	rspControllerEvent := RspControllerEvent{}

	err := json.Unmarshal([]byte(reading.Value), &rspControllerEvent)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling RSP Controller event: %v", err)
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
			loggingClient.Debug(fmt.Sprintf("Ignoring RSP Controller moved event: EPC-%v, FacilityId-%v, Location-%v, Timestamp-%v\n", apData.EPCCode, apData.FacilityId, apData.Location, apData.TimeStamp))
			continue
		default:
			loggingClient.Debug(fmt.Sprintf("Unrecognized RSP Controller event: %v", apData.EventType))
			continue
		}

		rfidEvents = append(rfidEvents, rfidReading)
	}

	return rfidEvents, nil
}
