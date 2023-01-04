// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"fmt"
	"math"
)

func (eventsProcessing *EventsProcessor) scaleBasketReconciliation(scaleReading *ScaleEventEntry) {

	//if scale reading is negative
	if scaleReading.Delta < 0 {
		//attempt to match with item in eventsProcessing.suspectScaleItems
		for eventTime, suspectItem := range eventsProcessing.suspectScaleItems {
			if math.Abs(math.Abs(scaleReading.Delta)-suspectItem.Delta) < scalePrecision {
				//remove from eventsProcessing.suspectScaleItems
				delete(eventsProcessing.suspectScaleItems, eventTime)
				break
			}
		}
		//on negative scale drop, we don't want to add it to suspect list
		return
	}

	eventsProcessing.suspectScaleItems[scaleReading.EventTime] = scaleReading //add to eventsProcessing.suspectScaleItems first
	if len(eventsProcessing.rttlogData) == 0 {                                // if scale event occurs before basketOpen
		return
	}
	if eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ProductId != "" {
		currentRTTLEntry := &eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1]

		if eventsProcessing.rttlQuantityIsEach(*currentRTTLEntry) {
			weightRange := eventsProcessing.calculateCurrentWeightRange(currentRTTLEntry)
			if weightRange.ExpectedMinWeight > floatingPointTolerance { //scale is not confirmed. Compare vs. .000001 and not 0 due to floating point precision

				// divide by zero check, followed by check if scale delta is less than (rttl expected min weight / quantity)
				if currentRTTLEntry.Quantity < 1 || scaleReading.Delta < weightRange.ExpectedMinWeight/currentRTTLEntry.Quantity {
					currentRTTLEntry.ScaleConfirmed = false
					return
				}
			}
		}

		if !eventsProcessing.checkScaleConfirmed(currentRTTLEntry) {
			//reverse iterate through scaleBuffer till associatedRTTLEntry != nil
			for scaleBufferIterator := len(eventsProcessing.scaleData) - 1; eventsProcessing.scaleData[scaleBufferIterator].AssociatedRTTLEntry == nil; scaleBufferIterator-- {
				//if current iteration fits within allowed under the umbrella of allotted weight (due to multiple drops)
				if eventsProcessing.scaleData[scaleBufferIterator].Delta <= currentRTTLEntry.CurrentWeightRange.ExpectedMaxWeight {
					//check if assoc.Buffer is full
					if eventsProcessing.checkScaleConfirmed(currentRTTLEntry) {
						break
					}
					//cross associate, remove from unassoc. buffer
					lastScaleReading := &eventsProcessing.scaleData[scaleBufferIterator]
					lastScaleReading.AssociatedRTTLEntry = currentRTTLEntry
					currentRTTLEntry.AssociatedScaleItems = append(currentRTTLEntry.AssociatedScaleItems, lastScaleReading)
					delete(eventsProcessing.suspectScaleItems, (*lastScaleReading).EventTime)
					eventsProcessing.checkScaleConfirmed(currentRTTLEntry)
					if scaleBufferIterator == 0 {
						break
					}

				} else if scaleBufferIterator == 0 {
					break
				}
			}
		}
	}
}

func (eventsProcessing *EventsProcessor) rttlQuantityIsEach(rttlogEventEntry RTTLogEventEntry) bool {
	if rttlogEventEntry.QuantityUnit == quantityUnitEA || rttlogEventEntry.QuantityUnit == quantityUnitEach {
		return true
	}
	return false
}

func (eventsProcessing *EventsProcessor) calculateCurrentWeightRange(currentRTTLEntry *RTTLogEventEntry) ProductDetails {
	currentAllowedWeight := ProductDetails{"", currentRTTLEntry.ProductDetails.ExpectedMinWeight * currentRTTLEntry.Quantity, currentRTTLEntry.ProductDetails.ExpectedMaxWeight * currentRTTLEntry.Quantity, false}
	for _, scaleItem := range currentRTTLEntry.AssociatedScaleItems {
		currentAllowedWeight.ExpectedMinWeight = currentAllowedWeight.ExpectedMinWeight - scaleItem.Delta
		currentAllowedWeight.ExpectedMaxWeight = currentAllowedWeight.ExpectedMaxWeight - scaleItem.Delta
	}
	currentRTTLEntry.CurrentWeightRange = currentAllowedWeight
	return currentAllowedWeight
}

func (eventsProcessing *EventsProcessor) checkScaleConfirmed(rttlogEventEntry *RTTLogEventEntry) bool {
	if eventsProcessing.rttlQuantityIsEach(*rttlogEventEntry) {
		weightRange := eventsProcessing.calculateCurrentWeightRange(rttlogEventEntry)
		if weightRange.ExpectedMinWeight > floatingPointTolerance { //scale is not confirmed. Compare vs. .000001 and not 0 due to floating point precision
			rttlogEventEntry.ScaleConfirmed = false
			return rttlogEventEntry.ScaleConfirmed
		}

		// scale is confirmed i.e. the weight matches the rttls
		for weightRange.ExpectedMinWeight <= ((rttlogEventEntry.ProductDetails.ExpectedMinWeight * -1) + floatingPointTolerance) { //if overpopulated RTTL due to UPDATE in Quantity
			//pop latest out of assoc.ScaleBuffer, re-add to suspectItems
			lastAssociatedScaleItem := rttlogEventEntry.AssociatedScaleItems[len(rttlogEventEntry.AssociatedScaleItems)-1]
			eventsProcessing.suspectScaleItems[lastAssociatedScaleItem.EventTime] = lastAssociatedScaleItem
			eventsProcessing.deleteLastScaleItem(&(rttlogEventEntry.AssociatedScaleItems))
			weightRange = eventsProcessing.calculateCurrentWeightRange(rttlogEventEntry)
		}
		rttlogEventEntry.ScaleConfirmed = true
		return rttlogEventEntry.ScaleConfirmed
	}

	// Quantity unit is not "EA"/"EACH"

	totalAvailableWeight := rttlogEventEntry.Quantity
	var scaleWeight float64
	for _, associatedScaleItem := range rttlogEventEntry.AssociatedScaleItems {
		scaleWeight = scaleWeight + associatedScaleItem.Delta
	}
	var percentChange float64
	if scaleWeight != 0 {
		percentChange = math.Abs(((totalAvailableWeight - scaleWeight) / scaleWeight))
	} else {
		percentChange = 0
	}

	tolerance := eventsProcessing.GetScaleToScaleTolerance()
	fmt.Printf("Quantity unit %v, weight diff: %v (%v%%), scale-to-scale tolerance: %v (%v%%)",
		rttlogEventEntry.QuantityUnit, percentChange, percentChange*100, tolerance, tolerance*100)

	if percentChange < tolerance || percentChange == 0 {
		rttlogEventEntry.ScaleConfirmed = true
		return rttlogEventEntry.ScaleConfirmed
	}
	rttlogEventEntry.ScaleConfirmed = false
	return rttlogEventEntry.ScaleConfirmed
}

func (eventsProcessing *EventsProcessor) cvBasketReconciliation(rttlReading *RTTLogEventEntry) {

	for cvIndex, cvItem := range eventsProcessing.currentCVData {
		if rttlReading.ProductName == cvItem.ObjectName {
			// check that the cvItem was at the scanner when the rttl was scanned
			// if CvTimeAlignment is negative ignore time alignment entirely
			if (math.Abs(float64(rttlReading.EventTime-cvItem.ROIs[ScannerROI].LastAtLocation)) < float64(eventsProcessing.cvTimeAlignment)) || eventsProcessing.cvTimeAlignment < 0 {
				//cross-associate
				rttlReading.AssociatedCVItems = append(rttlReading.AssociatedCVItems, &eventsProcessing.currentCVData[cvIndex])
				eventsProcessing.currentCVData[cvIndex].AssociatedRTTLEntry = rttlReading

				if math.Abs(float64(len(rttlReading.AssociatedCVItems))-rttlReading.Quantity) <= floatingPointTolerance {
					rttlReading.CVConfirmed = true
				}
			}
		}
	}
}

func (eventsProcessing *EventsProcessor) rfidBasketReconciliation(rttlReading *RTTLogEventEntry) error {
	rttlQuantity := rttlReading.Quantity
	for rfidIndex, rfidItem := range eventsProcessing.currentRFIDData {
		if rttlQuantity == 0 {
			break
		}

		//todo - && !AtGoBack && !AtEntrance
		//todo - priority of removing suspect RFID items (Bagging area first, then scanner, etc.)
		if rfidItem.AssociatedRTTLEntry == nil && rfidItem.UPC == rttlReading.ProductId {
			//cross associate
			rttlReading.AssociatedRFIDItems = append(rttlReading.AssociatedRFIDItems, &eventsProcessing.currentRFIDData[rfidIndex])
			eventsProcessing.currentRFIDData[rfidIndex].AssociatedRTTLEntry = rttlReading
			rttlQuantity--
		}
	}

	if math.Abs(float64(len(rttlReading.AssociatedRFIDItems))-rttlReading.Quantity) <= floatingPointTolerance {
		rttlReading.RFIDConfirmed = true
	}

	return nil
}
