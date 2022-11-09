// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"errors"
	"fmt"
	"math"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"event-reconciler/rfidgtin"
)

const (
	quantityUnitEA         = "EA"
	quantityUnitEach       = "Each"
	floatingPointTolerance = .000001
	scalePrecision         = 0.01
	scaleStatusOK          = "OK"
)

func (eventsProcessing *EventsProcessor) ProcessCheckoutEvents(edgexcontext interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := edgexcontext.LoggingClient()

	devicePos := eventsProcessing.processConfig.DevicePos

	deviceScale := eventsProcessing.processConfig.DeviceScale

	deviceCV := eventsProcessing.processConfig.DeviceCV

	deviceRFID := eventsProcessing.processConfig.DeviceRFID

	event, ok := data.(dtos.Event)
	if !ok {
		return false, errors.New("unable to cast event to dtos.Event")
	}
	for _, reading := range event.Readings {
		readingData := reading
		resourceName := readingData.ResourceName
		lc.Debugf("Processing Checkout Event: %s", resourceName)
		eventOk := eventsProcessing.checkEventOrderValid(resourceName, edgexcontext)
		if !eventOk {
			lc.Errorf("Error: event occurred out of order: %v", resourceName)
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

		msg := eventsProcessing.formatWebsocketMessage(resourceName)
		eventsProcessing.sendWebsocketMessage(msg, edgexcontext)
	}

	lc.Tracef("RTTLog: %v", eventsProcessing.rttlogData)
	lc.Tracef("scaleData: %v", eventsProcessing.scaleData)
	lc.Tracef("CvData: %v", eventsProcessing.currentCVData)
	lc.Tracef("RfidData: %v", eventsProcessing.currentRFIDData)

	return false, nil
}

func (eventsProcessing *EventsProcessor) processDeviceCVReading(reading dtos.BaseReading, lc logger.LoggingClient) {
	cvReading := CVEventEntry{
		ROIs: make(map[string]ROILocation),
	}
	err := eventsProcessing.unmarshalObjValue(reading.ObjectReading.ObjectValue, &cvReading)
	if err != nil {
		lc.Errorf("CV unmarshal failure: %v", err)
		return
	}

	cvObject := eventsProcessing.getExistingCVDataByObjectName(cvReading)

	if cvObject == nil {
		//object does not exist in currentCVData
		updateCVObjectLocation(cvReading, &cvReading, lc)
		if eventsProcessing.afterPaymentSuccess {
			eventsProcessing.nextCVData = append(eventsProcessing.nextCVData, cvReading)
		} else {
			eventsProcessing.currentCVData = append(eventsProcessing.currentCVData, cvReading)
		}
	} else {
		updateCVObjectLocation(cvReading, cvObject, lc)
	}

	for rttlIndex, rttl := range eventsProcessing.rttlogData {
		if !rttl.CVConfirmed && rttl.EventType == posItemEvent {
			eventsProcessing.cvBasketReconciliation(&eventsProcessing.rttlogData[rttlIndex])
		}
	}
}

func (eventsProcessing *EventsProcessor) processDeviceRFIDReading(reading dtos.BaseReading, lc logger.LoggingClient) {
	rfidReading := RFIDEventEntry{}
	err := eventsProcessing.unmarshalObjValue(reading.ObjectValue, &rfidReading)
	if err != nil {
		lc.Errorf("RFID unmarshal failure: %v", err)
		return
	}

	rfidReading.ROIs = make(map[string]ROILocation)

	upc, err := rfidgtin.GetGtin14(rfidReading.EPC)
	if err != nil {
		lc.Errorf("Bad EPC value. Not adding RFID tag to buffer: %v", err)
		return
	}

	//check if UPC is in Product lookup database. If not, don't add RFID tag to buffer
	prodDetails, err := productLookup(upc, eventsProcessing.processConfig.ProductLookupEndpoint)
	if err != nil {
		lc.Warnf("Could not find RFID tagged product (%s) in database. Not adding to buffer: %v", upc, err)
		return
	}
	rfidReading.UPC = upc
	rfidReading.ProductName = prodDetails.Name

	rfidObject := eventsProcessing.getExistingRFIDDataByEPC(rfidReading)

	if rfidObject == nil {
		//Add new RFID Entry to currentRFIDData
		updateRFIDObjectLocation(rfidReading, &rfidReading, lc)
		if eventsProcessing.afterPaymentSuccess {
			eventsProcessing.nextRFIDData = append(eventsProcessing.nextRFIDData, rfidReading)
		} else {
			eventsProcessing.currentRFIDData = append(eventsProcessing.currentRFIDData, rfidReading)
		}

	} else {
		//Update existing RFID entry in currentRFIDData
		updateRFIDObjectLocation(rfidReading, rfidObject, lc)
	}
}

func (eventsProcessing *EventsProcessor) processDeviceScaleReading(reading dtos.BaseReading, lc logger.LoggingClient) {
	scaleReading := ScaleEventEntry{}
	err := eventsProcessing.unmarshalObjValue(reading.ObjectReading.ObjectValue, &scaleReading)
	if err != nil {
		lc.Errorf("Scale unmarshal failure: %v", err)
		return
	}

	eventsProcessing.calculateScaleDelta(&scaleReading)
	if math.Abs(scaleReading.Delta) <= scalePrecision { //scale delta should be significant enough to be considered an event
		return
	}

	lc.Debugf("Adding %s to Scale Log", reading.ResourceName)

	eventsProcessing.scaleData = append(eventsProcessing.scaleData, scaleReading)

	eventsProcessing.scaleBasketReconciliation(&eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1])
}

func (eventsProcessing *EventsProcessor) processDevicePosReading(reading dtos.BaseReading, edgexcontext interfaces.AppFunctionContext) {
	lc := edgexcontext.LoggingClient()
	resourceName := reading.ResourceName

	rttLogReading := RTTLogEventEntry{}
	err := eventsProcessing.unmarshalObjValue(reading.ObjectReading.ObjectValue, &rttLogReading)
	if err != nil {
		lc.Errorf("RTTlog unmarshal failure: %v", err)
		return
	}

	rttLogReading.EventType = resourceName

	//if ProductId is given and is not 14 digits, prepend with 0s to convert everything to a GTIN14
	if rttLogReading.ProductId != "" {
		rttLogReading.ProductId = eventsProcessing.convertProductIDTo14Char(rttLogReading.ProductId)
	}

	switch resourceName {
	case basketOpenEvent:
		eventsProcessing.resetRTTLBasket()
		//only reset Baskets after first basketOpen
		if eventsProcessing.firstBasketOpenComplete {
			eventsProcessing.resetCVBasket()
			eventsProcessing.resetRFIDBasket()
		} else {
			eventsProcessing.firstBasketOpenComplete = true
		}

	case basketCloseEvent:
		eventsProcessing.resetRTTLBasket()

		// Adding these two basket resets to clear the 'blacklists' after a payment for the demo
		eventsProcessing.resetCVBasket()
		eventsProcessing.resetRFIDBasket()

	case removeItemEvent:
		err := eventsProcessing.removeRTTLItemFromBuffer(rttLogReading)
		if err != nil {
			lc.Errorf("Remove Item Error: %v", err)
		}

		eventsProcessing.eventOccurred[posItemEvent] = eventsProcessing.checkRTTLForPOSItems()
		return

	case posItemEvent:
		if rttLogReading.QuantityUnit == quantityUnitEA || rttLogReading.QuantityUnit == quantityUnitEach {
			//if QuantityUnit is "EA", there is a expected minimum and maximum weight. Otherwise, you only consider the weight of the purchase
			rttLogReading.ProductDetails, err = productLookup(rttLogReading.ProductId, eventsProcessing.processConfig.ProductLookupEndpoint)
			if err != nil {
				lc.Errorf("Product Lookup failed for product: %s. Not adding to RTTL. Error Message: %s", rttLogReading.ProductId, err.Error())
				return
			}
			lc.Tracef("Found product detail for %s", rttLogReading.ProductId)
		} else {
			rttLogReading.ProductDetails = ProductDetails{"", rttLogReading.Quantity, rttLogReading.Quantity, false}
		}

		eventsProcessing.cvBasketReconciliation(&rttLogReading)

		if eventsProcessing.isRFIDEligible(rttLogReading) {
			err := eventsProcessing.rfidBasketReconciliation(&rttLogReading)
			if err != nil {
				lc.Errorf("EPC to UPC transform failure for RFID Basket Reconciliation: %v", err)
			}
		}

	case paymentStartEvent:
		eventsProcessing.updateSuspectRFIDItems()

		suspectCVItems := eventsProcessing.getSuspectCVItems()
		suspectRFIDItems := eventsProcessing.getSuspectRFIDItems()

		if len(eventsProcessing.suspectScaleItems) > 0 || len(suspectCVItems) > 0 || len(suspectRFIDItems) > 0 {
			outputData, err := eventsProcessing.wrapSuspectItems()
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
		eventsProcessing.afterPaymentSuccess = true

	default:
		lc.Errorf("Unkown POS event: %s", resourceName)
	}

	lc.Tracef("Adding %s to RTT Log", resourceName)

	if len(eventsProcessing.rttlogData) == 0 {
		eventsProcessing.rttlogData = append(eventsProcessing.rttlogData, rttLogReading)
	} else {
		previousrttlogData := eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1]
		if previousrttlogData.ProductId == rttLogReading.ProductId && rttLogReading.ProductId != "" {
			eventsProcessing.appendToPreviousPosItem(rttLogReading)
		} else {
			eventsProcessing.rttlogData = append(eventsProcessing.rttlogData, rttLogReading)
		}
	}

}
