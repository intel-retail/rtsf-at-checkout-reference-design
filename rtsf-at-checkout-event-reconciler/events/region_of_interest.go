// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

const (
	EntranceROI    = "Entrance"
	DepartureROI   = "Departure"
	GoBackROI      = "Go Back"
	StagingROI     = "Staging"
	ScannerROI     = "Scanner"
	BaggingROI     = "Bagging"
	CartROI        = "Cart"
	ROIActionEnter = "ENTERED"
	ROIActionExit  = "EXITED"

	ROIActionErrorMessage = "Could not recognize ROI Action: %s\n"
)

func updateCVObjectLocation(cvEvent CVEventEntry, currentCVItem *CVEventEntry, edgexcontext *appcontext.Context) {
	currentCVItem.ROIName = cvEvent.ROIName
	currentCVItem.ROIAction = cvEvent.ROIAction
	currentCVItem.EventTime = cvEvent.EventTime

	roiLocation, ok := currentCVItem.ROIs[cvEvent.ROIName]
	if !ok {
		//First time visiting ROI, add new ROI
		roiLocation = ROILocation{}
	}

	switch cvEvent.ROIAction {
	case ROIActionEnter:
		roiLocation.AtLocation = true
	case ROIActionExit:
		roiLocation.AtLocation = false
	default:
		edgexcontext.LoggingClient.Error(fmt.Sprintf(ROIActionErrorMessage, cvEvent.ROIAction))
		return //dont update ROI if action is unrecoginzed E.G. not "ENTERED" or "EXITED"
	}

	roiLocation.LastAtLocation = cvEvent.EventTime

	currentCVItem.ROIs[cvEvent.ROIName] = roiLocation

}

func updateRFIDObjectLocation(rfidRoiEvent RFIDEventEntry, currentRFIDItem *RFIDEventEntry, edgexcontext *appcontext.Context) {
	currentRFIDItem.ROIName = rfidRoiEvent.ROIName
	currentRFIDItem.ROIAction = rfidRoiEvent.ROIAction
	currentRFIDItem.EventTime = rfidRoiEvent.EventTime

	roiLocation, ok := currentRFIDItem.ROIs[rfidRoiEvent.ROIName]
	if !ok {
		roiLocation = ROILocation{}
	}

	switch rfidRoiEvent.ROIAction {
	case ROIActionEnter:
		roiLocation.AtLocation = true
	case ROIActionExit:
		roiLocation.AtLocation = false
	default:
		edgexcontext.LoggingClient.Error(fmt.Sprintf(ROIActionErrorMessage, rfidRoiEvent.ROIAction))
		return

	}

	roiLocation.LastAtLocation = rfidRoiEvent.EventTime
	currentRFIDItem.ROIs[rfidRoiEvent.ROIName] = roiLocation
}
