// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"fmt"
	"math"
)

var ScaleToScaleTolerance float64 = 0.02

func scaleBasketReconciliation(scaleReading *ScaleEventEntry) {

	//if scale reading is negative
	if scaleReading.Delta < 0 {
		//attempt to match with item in suspectScaleItems
		for eventTime, suspectItem := range SuspectScaleItems {
			if math.Abs(math.Abs(scaleReading.Delta)-suspectItem.Delta) < scalePrecision {
				//remove from suspectScaleItems
				delete(SuspectScaleItems, eventTime)
				break
			}
		}
		//on negative scale drop, we don't want to add it to suspect list
		return
	}

	SuspectScaleItems[scaleReading.EventTime] = scaleReading //add to suspectScaleItems first
	if len(RttlogData) == 0 {                                // if scale event occurs before basketOpen
		return
	}
	if RttlogData[len(RttlogData)-1].ProductId != "" {
		currentRTTLEntry := &RttlogData[len(RttlogData)-1]
		if checkScaleConfirmed(currentRTTLEntry) == false {
			//reverse iterate through scaleBuffer till associatedRTTLEntry != nil
			for scaleBufferIterator := len(ScaleData) - 1; ScaleData[scaleBufferIterator].AssociatedRTTLEntry == nil; scaleBufferIterator-- {
				//if current iteration fits within allowed under the umbrella of allotted weight (due to multiple drops)
				if ScaleData[scaleBufferIterator].Delta <= currentRTTLEntry.CurrentWeightRange.ExpectedMaxWeight {
					//check if assoc.Buffer is full
					if checkScaleConfirmed(currentRTTLEntry) {
						break
					}
					//cross associate, remove from unassoc. buffer
					lastScaleReading := &ScaleData[scaleBufferIterator]
					lastScaleReading.AssociatedRTTLEntry = currentRTTLEntry
					currentRTTLEntry.AssociatedScaleItems = append(currentRTTLEntry.AssociatedScaleItems, lastScaleReading)
					delete(SuspectScaleItems, (*lastScaleReading).EventTime)
					checkScaleConfirmed(currentRTTLEntry)
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

func calculateCurrentWeightRange(currentRTTLEntry *RTTLogEventEntry) ProductDetails {
	currentAllowedWeight := ProductDetails{"", currentRTTLEntry.ProductDetails.ExpectedMinWeight * currentRTTLEntry.Quantity, currentRTTLEntry.ProductDetails.ExpectedMaxWeight * currentRTTLEntry.Quantity, false}
	for _, scaleItem := range currentRTTLEntry.AssociatedScaleItems {
		currentAllowedWeight.ExpectedMinWeight = currentAllowedWeight.ExpectedMinWeight - scaleItem.Delta
		currentAllowedWeight.ExpectedMaxWeight = currentAllowedWeight.ExpectedMaxWeight - scaleItem.Delta
	}
	currentRTTLEntry.CurrentWeightRange = currentAllowedWeight
	return currentAllowedWeight
}

func checkScaleConfirmed(rttlogEventEntry *RTTLogEventEntry) bool {
	if rttlogEventEntry.QuantityUnit == quantityUnitEA || rttlogEventEntry.QuantityUnit == quantityUnitEach {
		weightRange := calculateCurrentWeightRange(rttlogEventEntry)
		if weightRange.ExpectedMinWeight > floatingPointTolerance { //scale is not confirmed. Compare vs. .000001 and not 0 due to floating point precision
			rttlogEventEntry.ScaleConfirmed = false
			return rttlogEventEntry.ScaleConfirmed
		}
		for weightRange.ExpectedMinWeight <= ((rttlogEventEntry.ProductDetails.ExpectedMinWeight * -1) + floatingPointTolerance) { //if overpopulated RTTL due to UPDATE in Quantity
			//pop latest out of assoc.ScaleBuffer, re-add to suspectItems
			lastAssociatedScaleItem := rttlogEventEntry.AssociatedScaleItems[len(rttlogEventEntry.AssociatedScaleItems)-1]
			SuspectScaleItems[lastAssociatedScaleItem.EventTime] = lastAssociatedScaleItem
			deleteLastScaleItem(&(rttlogEventEntry.AssociatedScaleItems))
			weightRange = calculateCurrentWeightRange(rttlogEventEntry)
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

	fmt.Printf("Quantity unit %v, weight diff: %v (%v%%), scale-to-scale tolerance: %v (%v%%)",
		rttlogEventEntry.QuantityUnit, percentChange, percentChange*100, ScaleToScaleTolerance, ScaleToScaleTolerance*100)

	if percentChange < ScaleToScaleTolerance || percentChange == 0 {
		rttlogEventEntry.ScaleConfirmed = true
		return rttlogEventEntry.ScaleConfirmed
	}
	rttlogEventEntry.ScaleConfirmed = false
	return rttlogEventEntry.ScaleConfirmed
}

func cvBasketReconciliation(rttlReading *RTTLogEventEntry) {

	for cvIndex, cvItem := range CurrentCVData {
		if rttlReading.ProductName == cvItem.ObjectName {
			// check that the cvItem was at the scanner when the rttl was scanned
			// if CvTimeAlignment is negative ignore time alignment entirely
			if (math.Abs(float64(rttlReading.EventTime-cvItem.ROIs[ScannerROI].LastAtLocation)) < float64(CvTimeAlignment)) || CvTimeAlignment < 0 {
				//cross-associate
				rttlReading.AssociatedCVItems = append(rttlReading.AssociatedCVItems, &CurrentCVData[cvIndex])
				CurrentCVData[cvIndex].AssociatedRTTLEntry = rttlReading

				if math.Abs(float64(len(rttlReading.AssociatedCVItems))-rttlReading.Quantity) <= floatingPointTolerance {
					rttlReading.CVConfirmed = true
				}
			}
		}
	}
}

func rfidBasketReconciliation(rttlReading *RTTLogEventEntry) error {
	rttlQuantity := rttlReading.Quantity
	for rfidIndex, rfidItem := range CurrentRFIDData {
		if rttlQuantity == 0 {
			break
		}

		//todo - && !AtGoBack && !AtEntrance
		//todo - priority of removing suspect RFID items (Bagging area first, then scanner, etc.)
		if rfidItem.AssociatedRTTLEntry == nil && rfidItem.UPC == rttlReading.ProductId {
			//cross associate
			rttlReading.AssociatedRFIDItems = append(rttlReading.AssociatedRFIDItems, &CurrentRFIDData[rfidIndex])
			CurrentRFIDData[rfidIndex].AssociatedRTTLEntry = rttlReading
			rttlQuantity--
		}
	}

	if math.Abs(float64(len(rttlReading.AssociatedRFIDItems))-rttlReading.Quantity) <= floatingPointTolerance {
		rttlReading.RFIDConfirmed = true
	}

	return nil
}
