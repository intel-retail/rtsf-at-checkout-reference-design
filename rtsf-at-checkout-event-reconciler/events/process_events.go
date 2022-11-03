// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"event-reconciler/config"
	"event-reconciler/rfidgtin"
)

type EventsProcessor struct {
	ProcessConfig *config.ReconcilerConfig
}

const (
	quantityUnitEA         = "EA"
	quantityUnitEach       = "Each"
	floatingPointTolerance = .000001
	scalePrecision         = 0.01
	scaleStatusOK          = "OK"
)

var CvTimeAlignment time.Duration = 1 * time.Second
var RttlogData []RTTLogEventEntry
var ScaleData []ScaleEventEntry
var SuspectScaleItems = make(map[int64]*ScaleEventEntry)
var CurrentCVData = []CVEventEntry{}
var NextCVData = []CVEventEntry{}
var CurrentRFIDData = []RFIDEventEntry{}
var NextRFIDData = []RFIDEventEntry{}
var firstBasketOpenComplete = false
var afterPaymentSuccess = false

func (eventsProcessing *EventsProcessor) ProcessCheckoutEvents(edgexcontext interfaces.AppFunctionContext, param interface{}) (bool, interface{}) {
	lc := edgexcontext.LoggingClient()

	devicePos := eventsProcessing.ProcessConfig.DevicePos

	deviceScale := eventsProcessing.ProcessConfig.DeviceScale

	deviceCV := eventsProcessing.ProcessConfig.DeviceCV

	deviceRFID := eventsProcessing.ProcessConfig.DeviceRFID

	result, _ := param.(models.Event)
	for _, reading := range result.Readings {
		readingData := reading.(models.ObjectReading)
		eventName := readingData.ResourceName
		lc.Debugf("Processing Checkout Event: %s", eventName)
		eventOk := checkEventOrderValid(eventName, edgexcontext)
		if !eventOk {
			lc.Errorf("Error: event occurred out of order: %v", eventName)
			continue
		}

		switch readingData.DeviceName {
		case devicePos + "-rest", devicePos + "-mqtt":
			eventsProcessing.processDevicePosReading(readingData, edgexcontext)

		case deviceScale, deviceScale + "-rest", deviceScale + "-mqtt":
			eventsProcessing.processDeviceScaleReading(readingData, lc)

		case deviceCV + "-rest", deviceCV + "-mqtt":
			eventsProcessing.processDeviceCVReading(readingData, lc)

		case deviceRFID + "-rest", deviceRFID + "-mqtt":
			eventsProcessing.processDeviceRFIDReading(readingData, lc)

		default:
			lc.Errorf("Did not recognize Device: %s", readingData.DeviceName)
			continue
		}

		msg := formatWebsocketMessage(eventName)
		sendWebsocketMessage(msg, edgexcontext)
	}

	lc.Tracef("RTTLog: %v", RttlogData)
	lc.Tracef("ScaleData: %v", ScaleData)
	lc.Tracef("CvData: %v", CurrentCVData)
	lc.Tracef("RfidData: %v", CurrentRFIDData)

	return false, nil
}

func (eventsProcessing *EventsProcessor) processDeviceCVReading(reading models.ObjectReading, lc logger.LoggingClient) {
	cvReading := CVEventEntry{
		ROIs: make(map[string]ROILocation),
	}
	err := json.Unmarshal([]byte(reading.ObjectValue.(string)), &cvReading)
	if err != nil {
		lc.Errorf("CV unmarshal failure: %v", err)
		return
	}

	cvObject := getExistingCVDataByObjectName(cvReading)

	if cvObject == nil {
		//object does not exist in CurrentCVData
		updateCVObjectLocation(cvReading, &cvReading, lc)
		if afterPaymentSuccess {
			NextCVData = append(NextCVData, cvReading)
		} else {
			CurrentCVData = append(CurrentCVData, cvReading)
		}
	} else {
		updateCVObjectLocation(cvReading, cvObject, lc)
	}

	for rttlIndex, rttl := range RttlogData {
		if !rttl.CVConfirmed && rttl.EventType == posItemEvent {
			cvBasketReconciliation(&RttlogData[rttlIndex])
		}
	}
}

func (eventsProcessing *EventsProcessor) processDeviceRFIDReading(reading models.ObjectReading, lc logger.LoggingClient) {
	rfidReading := RFIDEventEntry{
		ROIs: make(map[string]ROILocation),
	}
	err := json.Unmarshal([]byte(reading.ObjectValue.(string)), &rfidReading)
	if err != nil {
		lc.Errorf("RFID unmarshal failure: %v", err)
		return
	}

	upc, err := rfidgtin.GetGtin14(rfidReading.EPC)
	if err != nil {
		lc.Errorf("Bad EPC value. Not adding RFID tag to buffer: %v", err)
		return
	}

	//check if UPC is in Product lookup database. If not, don't add RFID tag to buffer
	prodDetails, err := productLookup(upc, lc, eventsProcessing.ProcessConfig.ProductLookupEndpoint)
	if err != nil {
		lc.Warnf("Could not find RFID tagged product (%s) in database. Not adding to buffer: %v", upc, err)
		return
	}
	rfidReading.UPC = upc
	rfidReading.ProductName = prodDetails.Name

	rfidObject := getExistingRFIDDataByEPC(rfidReading)

	if rfidObject == nil {
		//Add new RFID Entry to CurrentRFIDData
		updateRFIDObjectLocation(rfidReading, &rfidReading, lc)
		if afterPaymentSuccess {
			NextRFIDData = append(NextRFIDData, rfidReading)
		} else {
			CurrentRFIDData = append(CurrentRFIDData, rfidReading)
		}

	} else {
		//Update existing RFID entry in CurrentRFIDData
		updateRFIDObjectLocation(rfidReading, rfidObject, lc)
	}
}

func (eventsProcessing *EventsProcessor) processDeviceScaleReading(reading models.ObjectReading, lc logger.LoggingClient) {

	scaleReading := ScaleEventEntry{}
	err := json.Unmarshal([]byte(reading.ObjectValue.(string)), &scaleReading)
	if err != nil {
		lc.Errorf("Scale unmarshal failure: %v", err)
		return
	}

	calculateScaleDelta(&scaleReading)
	if math.Abs(scaleReading.Delta) <= scalePrecision { //scale delta should be significant enough to be considered an event
		return
	}

	lc.Debugf("Adding %s to Scale Log", reading.ResourceName)

	ScaleData = append(ScaleData, scaleReading)

	scaleBasketReconciliation(&ScaleData[len(ScaleData)-1])
}

func (eventsProcessing *EventsProcessor) processDevicePosReading(reading models.ObjectReading, edgexcontext interfaces.AppFunctionContext) {
	lc := edgexcontext.LoggingClient()
	eventName := reading.ResourceName

	rttLogReading := RTTLogEventEntry{EventType: eventName}
	err := json.Unmarshal([]byte(reading.ObjectValue.(string)), &rttLogReading)
	if err != nil {
		lc.Errorf("RTTlog unmarshal failure: %v", err)
		return
	}

	//if ProductId is given and is not 14 digits, prepend with 0s to convert everything to a GTIN14
	if rttLogReading.ProductId != "" {
		rttLogReading.ProductId = convertProductIDTo14Char(rttLogReading.ProductId)
	}

	switch eventName {
	case basketOpenEvent:
		resetRTTLBasket()
		//only reset Baskets after first basketOpen
		if firstBasketOpenComplete {
			resetCVBasket()
			resetRFIDBasket()
		} else {
			firstBasketOpenComplete = true
		}

	case basketCloseEvent:
		resetRTTLBasket()

		// Adding these two basket resets to clear the 'blacklists' after a payment for the demo
		resetCVBasket()
		resetRFIDBasket()

	case removeItemEvent:
		err := removeRTTLItemFromBuffer(rttLogReading)
		if err != nil {
			lc.Errorf("Remove Item Error: %v", err)
		}

		EventOccurred[posItemEvent] = checkRTTLForPOSItems()
		return

	case posItemEvent:
		if rttLogReading.QuantityUnit == quantityUnitEA || rttLogReading.QuantityUnit == quantityUnitEach {
			//if QuantityUnit is "EA", there is a expected minimum and maximum weight. Otherwise, you only consider the weight of the purchase
			rttLogReading.ProductDetails, err = productLookup(rttLogReading.ProductId, lc, eventsProcessing.ProcessConfig.ProductLookupEndpoint)
			if err != nil {
				lc.Errorf("Product Lookup failed for product: %s. Not adding to RTTL. Error Message: %s", rttLogReading.ProductId, err.Error())
				return
			}
			lc.Tracef("Found product detail for %s", rttLogReading.ProductId)
		} else {
			rttLogReading.ProductDetails = ProductDetails{"", rttLogReading.Quantity, rttLogReading.Quantity, false}
		}

		cvBasketReconciliation(&rttLogReading)

		if isRFIDEligible(rttLogReading) {
			err := rfidBasketReconciliation(&rttLogReading)
			if err != nil {
				lc.Errorf("EPC to UPC transform failure for RFID Basket Reconciliation: %v", err)
			}
		}

	case paymentStartEvent:
		updateSuspectRFIDItems()

		suspectCVItems := getSuspectCVItems()
		suspectRFIDItems := getSuspectRFIDItems()

		if len(SuspectScaleItems) > 0 || len(suspectCVItems) > 0 || len(suspectRFIDItems) > 0 {
			outputData, err := wrapSuspectItems()
			if err != nil {
				lc.Error("Failed to marshal suspect items for output")
			}
			lc.Info("Suspect items detected, sending to message bus")
			//export suspect  items
			// Not using logger so that it pretty prints
			fmt.Println(string(outputData))
			edgexcontext.SetResponseData(outputData)
		} else {
			// Not using logger so it stands out in docker log
			fmt.Println("No suspect items detected")
		}

	case paymentSuccessEvent:
		afterPaymentSuccess = true

	default:
		lc.Errorf("Unkown POS event: %s", eventName)
	}

	lc.Tracef("Adding %s to RTT Log", eventName)

	if len(RttlogData) == 0 {
		RttlogData = append(RttlogData, rttLogReading)
	} else {
		previousRTTLogData := RttlogData[len(RttlogData)-1]
		if previousRTTLogData.ProductId == rttLogReading.ProductId && rttLogReading.ProductId != "" {
			appendToPreviousPosItem(rttLogReading)
		} else {
			RttlogData = append(RttlogData, rttLogReading)
		}
	}

}
