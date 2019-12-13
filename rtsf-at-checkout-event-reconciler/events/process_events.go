// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"event-reconciler/rfidgtin"
)

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

func ProcessCheckoutEvents(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {


	if len(params) < 1 {
		edgexcontext.LoggingClient.Error("Didn't receive an event")
		return false, nil
	}

	appsettings := edgexcontext.Configuration.ApplicationSettings
	devicePos, ok := appsettings["DevicePos"]
	if !ok {
		edgexcontext.LoggingClient.Error("DevicePos setting not found")
		return false, nil
	}
	deviceScale, ok := appsettings["DeviceScale"]
	if !ok {
		edgexcontext.LoggingClient.Error("DeviceScale setting not found")
		return false, nil
	}
	deviceCV, ok := appsettings["DeviceCV"]
	if !ok {
		edgexcontext.LoggingClient.Error("DeviceCV setting not found")
		return false, nil
	}
	deviceRFID, ok := appsettings["DeviceRFID"]
	if !ok {
		edgexcontext.LoggingClient.Error("DeviceRFID setting not found")
		return false, nil
	}

	result, _ := params[0].(models.Event)
	for _, reading := range result.Readings {
		eventName := reading.Name
		edgexcontext.LoggingClient.Debug(fmt.Sprintf("Processing Checkout Event: %s", eventName))
		eventOk := checkEventOrderValid(eventName, edgexcontext)
		if !eventOk {
			edgexcontext.LoggingClient.Error(fmt.Sprintf("Error: event occurred out of order: %v", eventName))
			continue
		}

		switch reading.Device {
		case devicePos+"-rest", devicePos+"-mqtt":
			processDevicePosReading(reading, edgexcontext)

		case deviceScale, deviceScale+"-rest", deviceScale+"-mqtt":
			processDeviceScaleReading(reading, edgexcontext)

		case deviceCV+"-rest", deviceCV+"-mqtt":
			processDeviceCVReading(reading, edgexcontext)

		case deviceRFID+"-rest", deviceRFID+"-mqtt":
			processDeviceRFIDReading(reading, edgexcontext)

		default:
			edgexcontext.LoggingClient.Error(fmt.Sprintf("Did not recognize Device: %s", reading.Device))
			continue
		}

		msg := formatWebsocketMessage(eventName)
		sendWebsocketMessage(msg, edgexcontext)
	}

	edgexcontext.LoggingClient.Debug(fmt.Sprintf("RTTLog: %v", RttlogData))
	edgexcontext.LoggingClient.Debug(fmt.Sprintf("ScaleData: %v", ScaleData))
	edgexcontext.LoggingClient.Debug(fmt.Sprintf("CvData: %v", CurrentCVData))
	edgexcontext.LoggingClient.Debug(fmt.Sprintf("RfidData: %v", CurrentRFIDData))

	return false, nil
}

func processDeviceCVReading(reading models.Reading, edgexcontext *appcontext.Context) {
	cvReading := CVEventEntry{
		ROIs: make(map[string]ROILocation),
	}
	err := json.Unmarshal([]byte(reading.Value), &cvReading)
	if err != nil {
		edgexcontext.LoggingClient.Error(fmt.Sprintf("CV unmarshal failure: %v", err))
		return
	}

	cvObject := getExistingCVDataByObjectName(cvReading)

	if cvObject == nil {
		//object does not exist in CurrentCVData
		updateCVObjectLocation(cvReading, &cvReading, edgexcontext)
		if afterPaymentSuccess {
			NextCVData = append(NextCVData, cvReading)
		} else {
			CurrentCVData = append(CurrentCVData, cvReading)
		}
	} else {
		updateCVObjectLocation(cvReading, cvObject, edgexcontext)
	}

	for rttlIndex, rttl := range RttlogData {
		if !rttl.CVConfirmed && rttl.EventType == posItemEvent {
			cvBasketReconciliation(&RttlogData[rttlIndex])
		}
	}
}

func processDeviceRFIDReading(reading models.Reading, edgexcontext *appcontext.Context) {
	rfidReading := RFIDEventEntry{
		ROIs: make(map[string]ROILocation),
	}
	err := json.Unmarshal([]byte(reading.Value), &rfidReading)
	if err != nil {
		edgexcontext.LoggingClient.Error(fmt.Sprintf("RFID unmarshal failure: %v", err))
		return
	}

	upc, err := rfidgtin.GetGtin14(rfidReading.EPC)
	if err != nil {
		edgexcontext.LoggingClient.Error(fmt.Sprintf("Bad EPC value. Not adding RFID tag to buffer: %v", err))
		return
	}

	//check if UPC is in Product lookup database. If not, don't add RFID tag to buffer
	prodDetails, err := productLookup(upc, edgexcontext)
	if err != nil {
		edgexcontext.LoggingClient.Warn(fmt.Sprintf("Could not find RFID tagged product (%s) in database. Not adding to buffer: %v", upc, err))
		return
	}
	rfidReading.UPC = upc
	rfidReading.ProductName = prodDetails.Name

	rfidObject := getExistingRFIDDataByEPC(rfidReading)

	if rfidObject == nil {
		//Add new RFID Entry to CurrentRFIDData
		updateRFIDObjectLocation(rfidReading, &rfidReading, edgexcontext)
		if afterPaymentSuccess {
			NextRFIDData = append(NextRFIDData, rfidReading)
		} else {
			CurrentRFIDData = append(CurrentRFIDData, rfidReading)
		}

	} else {
		//Update existing RFID entry in CurrentRFIDData
		updateRFIDObjectLocation(rfidReading, rfidObject, edgexcontext)
	}
}

func processDeviceScaleReading(reading models.Reading, edgexcontext *appcontext.Context) {
	scaleReading := ScaleEventEntry{}
	err := json.Unmarshal([]byte(reading.Value), &scaleReading)
	if err != nil {
		edgexcontext.LoggingClient.Error(fmt.Sprintf("Scale unmarshal failure: %v", err))
		return
	}

	calculateScaleDelta(&scaleReading)
	if math.Abs(scaleReading.Delta) <= scalePrecision { //scale delta should be significant enough to be considered an event
		return
	}

	edgexcontext.LoggingClient.Debug(fmt.Sprintf("Adding %s to Scale Log", reading.Name))

	ScaleData = append(ScaleData, scaleReading)

	scaleBasketReconciliation(&ScaleData[len(ScaleData)-1])
}

func processDevicePosReading(reading models.Reading, edgexcontext *appcontext.Context) {
	eventName := reading.Name
	rttLogReading := RTTLogEventEntry{EventType: eventName}
	err := json.Unmarshal([]byte(reading.Value), &rttLogReading)
	if err != nil {
		edgexcontext.LoggingClient.Error(fmt.Sprintf("RTTlog unmarshal failure: %v", err))
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
			edgexcontext.LoggingClient.Error(fmt.Sprintf("Remove Item Error: %v", err))
		}

		EventOccurred[posItemEvent] = checkRTTLForPOSItems()
		return

	case posItemEvent:
		if rttLogReading.QuantityUnit == quantityUnitEA || rttLogReading.QuantityUnit == quantityUnitEach {
			//if QuantityUnit is "EA", there is a expected minimum and maximum weight. Otherwise, you only consider the weight of the purchase
			rttLogReading.ProductDetails, err = productLookup(rttLogReading.ProductId, edgexcontext)
			if err != nil {
				edgexcontext.LoggingClient.Error(fmt.Sprintf("Product Lookup failed for product: %s. Not adding to RTTL. Error Message: %s", rttLogReading.ProductId, err.Error()))
				return
			}
			edgexcontext.LoggingClient.Debug(fmt.Sprintf("Found product detail for %s", rttLogReading.ProductId))
		} else {
			rttLogReading.ProductDetails = ProductDetails{"", rttLogReading.Quantity, rttLogReading.Quantity, false}
		}

		cvBasketReconciliation(&rttLogReading)

		if isRFIDEligible(rttLogReading) {
			err := rfidBasketReconciliation(&rttLogReading)
			if err != nil {
				edgexcontext.LoggingClient.Error(fmt.Sprintf("EPC to UPC transform failure for RFID Basket Reconciliation: %v", err))
			}
		}

	case paymentStartEvent:
		updateSuspectRFIDItems()

		suspectCVItems := getSuspectCVItems()
		suspectRFIDItems := getSuspectRFIDItems()

		if len(SuspectScaleItems) > 0 || len(suspectCVItems) > 0 || len(suspectRFIDItems) > 0 {
			outputData, err := wrapSuspectItems()
			if err != nil {
				edgexcontext.LoggingClient.Error("Failed to marshal suspect items for output")
			}
			edgexcontext.LoggingClient.Info("Pushing suspect items to message bus")
			//export suspect  items
			fmt.Println(string(outputData))
			edgexcontext.Complete(outputData)
		}

	case paymentSuccessEvent:
		afterPaymentSuccess = true

	default:
		edgexcontext.LoggingClient.Error(fmt.Sprintf("Unkown POS event: %s", eventName))
	}

	edgexcontext.LoggingClient.Debug(fmt.Sprintf("Adding %s to RTT Log", eventName))

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
