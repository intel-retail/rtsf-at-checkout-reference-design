// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"fmt"
	"math"
)

func (eventsProcessing *EventsProcessor) calculateScaleDelta(scaleReading *ScaleEventEntry) {
	if len(eventsProcessing.scaleData) == 0 {
		scaleReading.Delta = scaleReading.Total
	} else {
		previousReading := eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1]
		scaleReading.Delta = scaleReading.Total - previousReading.Total
	}
}

// consolidate the previous and current POS item into same RTTLogData entry
func (eventsProcessing *EventsProcessor) appendToPreviousPosItem(newItem RTTLogEventEntry) {
	if len(eventsProcessing.rttlogData) == 0 {
		return
	}
	previousItem := eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1]
	if len(previousItem.Collection) == 0 {
		// add the exisiting item to the collection list
		previousItem.Collection = append(previousItem.Collection, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1])
	}

	// add the new item to the existing collection
	previousItem.Collection = append(previousItem.Collection, newItem)
	previousItem.Quantity = previousItem.Quantity + newItem.Quantity

	eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1] = previousItem
}
func (eventsProcessing *EventsProcessor) deleteRTTLItemAtIndex(list *[]RTTLogEventEntry, index int) {
	if index == len(*list)-1 {
		// last item in list - gets deleted
		*list = (*list)[:len(*list)-1]
	} else if index < len(*list)-1 {
		//not last item
		*list = append((*list)[:index], (*list)[index+1:]...)
	}
}

func (eventsProcessing *EventsProcessor) removeRTTLItemFromBuffer(rttlogReading RTTLogEventEntry) error {
	quantityToRemove := rttlogReading.Quantity
	for rttlogIndex, item := range eventsProcessing.rttlogData {
		if item.ProductId == rttlogReading.ProductId {
			if item.Quantity >= (quantityToRemove - floatingPointTolerance) {
				// this means that items were not collapsed into a collection
				if len(item.Collection) == 0 {
					if math.Abs(item.Quantity-quantityToRemove) <= floatingPointTolerance { //checking if quantity and quantitytoRemove are equal
						// remove item at that index from RttlogData
						eventsProcessing.deleteRTTLItemAtIndex(&eventsProcessing.rttlogData, rttlogIndex)
					} else {
						item.Quantity = item.Quantity - quantityToRemove
						eventsProcessing.rttlogData[rttlogIndex] = item
					}
					quantityToRemove = 0
					break
				}
				// items were consolidated into a collection
				for quantityToRemove >= floatingPointTolerance { //while quantityToRemove is not 0
					collectionItem := item.Collection[0]
					if collectionItem.Quantity <= quantityToRemove {
						quantityToRemove = quantityToRemove - collectionItem.Quantity
						// remove item from collection
						eventsProcessing.deleteRTTLItemAtIndex(&item.Collection, 0)
						eventsProcessing.rttlogData[rttlogIndex].Quantity = eventsProcessing.rttlogData[rttlogIndex].Quantity - collectionItem.Quantity
						continue
					}
					//this means the collection item has more than we want to remove
					collectionItem.Quantity = collectionItem.Quantity - quantityToRemove
					eventsProcessing.rttlogData[rttlogIndex].Quantity = eventsProcessing.rttlogData[rttlogIndex].Quantity - quantityToRemove
					quantityToRemove = 0
					item.Collection[0] = collectionItem //store updated quantity back into collection

				}
				if len(item.Collection) == 0 {
					eventsProcessing.deleteRTTLItemAtIndex(&eventsProcessing.rttlogData, rttlogIndex)
				}

			}
		}
	}
	if quantityToRemove > floatingPointTolerance {
		return fmt.Errorf("error: Remove item failed: %s", rttlogReading.ProductName)
	}
	return nil
}

func (eventsProcessing *EventsProcessor) deleteLastScaleItem(list *[]*ScaleEventEntry) {
	if list == nil || len(*list) < 1 {
		return
	}

	*list = (*list)[:len(*list)-1]
}

func (eventsProcessing *EventsProcessor) wrapSuspectItems() ([]byte, error) {
	suspectList := SuspectLists{
		CVSuspect:    eventsProcessing.getSuspectCVItems(),
		RFIDSuspect:  eventsProcessing.getSuspectRFIDItems(),
		ScaleSuspect: eventsProcessing.getSuspectScaleItems(),
	}

	byteSuspects, err := json.MarshalIndent(suspectList, "", "   ")
	if err != nil {
		return nil, err
	}

	return byteSuspects, nil
}

func (eventsProcessing *EventsProcessor) resetRTTLBasket() {
	eventsProcessing.rttlogData = []RTTLogEventEntry{}
	eventsProcessing.scaleData = []ScaleEventEntry{}
	eventsProcessing.suspectScaleItems = make(map[int64]*ScaleEventEntry)
}

func (eventsProcessing *EventsProcessor) resetRFIDBasket() {
	eventsProcessing.persistRFIDGoBack()
	eventsProcessing.persistRFIDSuspectItems()
	eventsProcessing.currentRFIDData = eventsProcessing.nextRFIDData
	eventsProcessing.nextRFIDData = []RFIDEventEntry{}
	eventsProcessing.afterPaymentSuccess = false
}

func (eventsProcessing *EventsProcessor) persistRFIDGoBack() {
	for _, rfidItem := range eventsProcessing.currentRFIDData {
		if eventsProcessing.atROILocation(GoBackROI, rfidItem.ROIs) {
			eventsProcessing.nextRFIDData = append(eventsProcessing.nextRFIDData, rfidItem)
		}
	}
}

func (eventsProcessing *EventsProcessor) persistRFIDSuspectItems() {
	for _, rfidItem := range eventsProcessing.currentRFIDData {
		if rfidItem.AssociatedRTTLEntry == nil && !eventsProcessing.atROILocation(GoBackROI, rfidItem.ROIs) && !eventsProcessing.atROILocation(EntranceROI, rfidItem.ROIs) {
			eventsProcessing.nextRFIDData = append(eventsProcessing.nextRFIDData, rfidItem)
		}
	}
}

func (eventsProcessing *EventsProcessor) resetCVBasket() {
	eventsProcessing.persistCVGoBack()
	eventsProcessing.persistCVSuspectItems()
	eventsProcessing.currentCVData = eventsProcessing.nextCVData
	eventsProcessing.nextCVData = []CVEventEntry{}
	eventsProcessing.afterPaymentSuccess = false
}

func (eventsProcessing *EventsProcessor) persistCVGoBack() {
	for _, cvItem := range eventsProcessing.currentCVData {
		if eventsProcessing.atROILocation(GoBackROI, cvItem.ROIs) {
			eventsProcessing.nextCVData = append(eventsProcessing.nextCVData, cvItem)
		}
	}
}

func (eventsProcessing *EventsProcessor) persistCVSuspectItems() {
	for _, cvItem := range eventsProcessing.currentCVData {
		if cvItem.AssociatedRTTLEntry == nil && !eventsProcessing.atROILocation(GoBackROI, cvItem.ROIs) && !eventsProcessing.atROILocation(EntranceROI, cvItem.ROIs) {
			eventsProcessing.nextCVData = append(eventsProcessing.nextCVData, cvItem)
		}
	}
}

func (eventsProcessing *EventsProcessor) checkRTTLForPOSItems() bool {
	for _, rttlEvent := range eventsProcessing.rttlogData {
		if rttlEvent.Quantity > floatingPointTolerance {
			return true
		}
	}
	return false
}

func (eventsProcessing *EventsProcessor) getExistingCVDataByObjectName(cvReading CVEventEntry) *CVEventEntry {
	for cvIndex, cvItem := range eventsProcessing.currentCVData {
		if cvReading.ObjectName == cvItem.ObjectName {
			return &eventsProcessing.currentCVData[cvIndex]
		}
	}
	return nil
}

func (eventsProcessing *EventsProcessor) getExistingRFIDDataByEPC(rfidReading RFIDEventEntry) *RFIDEventEntry {
	for rfidIndex, rfidItem := range eventsProcessing.currentRFIDData {
		if rfidReading.EPC == rfidItem.EPC {
			return &eventsProcessing.currentRFIDData[rfidIndex]
		}
	}
	return nil
}

func (eventsProcessing *EventsProcessor) getSuspectScaleItems() map[int64]*ScaleEventEntry {

	scaleEntries := make(map[int64]*ScaleEventEntry)

	for key, scaleItem := range eventsProcessing.suspectScaleItems {
		newEntry := ScaleEventEntry{
			Delta:        scaleItem.Delta,
			Total:        scaleItem.Total,
			MinTolerance: scaleItem.MinTolerance,
			MaxTolerance: scaleItem.MaxTolerance,
			Units:        scaleItem.Units,
			SettlingTime: scaleItem.SettlingTime,
			MaxWeight:    scaleItem.MaxWeight,
			LaneId:       scaleItem.LaneId,
			ScaleId:      scaleItem.ScaleId,
			EventTime:    scaleItem.EventTime,
			Status:       scaleItem.Status,
		}
		scaleEntries[key] = &newEntry
	}
	return scaleEntries
}

func (eventsProcessing *EventsProcessor) getSuspectCVItems() []CVEventEntry {
	suspectItems := []CVEventEntry{}
	for _, cvItem := range eventsProcessing.currentCVData {
		if cvItem.AssociatedRTTLEntry == nil && !eventsProcessing.atROILocation(GoBackROI, cvItem.ROIs) && !eventsProcessing.atROILocation(EntranceROI, cvItem.ROIs) {
			suspectItems = append(suspectItems, cvItem)
		}
	}
	return suspectItems
}

func (eventsProcessing *EventsProcessor) getSuspectRFIDItems() []RFIDEventEntry {
	suspectItems := []RFIDEventEntry{}
	for _, rfidItem := range eventsProcessing.currentRFIDData {
		if rfidItem.AssociatedRTTLEntry == nil && !eventsProcessing.atROILocation(GoBackROI, rfidItem.ROIs) && !eventsProcessing.atROILocation(EntranceROI, rfidItem.ROIs) {
			suspectItems = append(suspectItems, rfidItem)
		}
	}
	return suspectItems
}

func (eventsProcessing *EventsProcessor) updateSuspectRFIDItems() {
	for rfidIndex, rfidItem := range eventsProcessing.currentRFIDData {
		if eventsProcessing.currentRFIDData[rfidIndex].AssociatedRTTLEntry != nil ||
			eventsProcessing.atROILocation(GoBackROI, eventsProcessing.currentRFIDData[rfidIndex].ROIs) ||
			eventsProcessing.atROILocation(EntranceROI, eventsProcessing.currentRFIDData[rfidIndex].ROIs) {
			continue
		}

		for rttlIndex, rttlItem := range eventsProcessing.rttlogData {
			if !eventsProcessing.isRFIDEligible(rttlItem) {
				continue
			}
			//wont be a fractional quantity if its rfid-eligible - it will be "EA"
			if int(rttlItem.Quantity) > len(rttlItem.AssociatedRFIDItems) && rttlItem.ProductId == rfidItem.UPC {
				//cross-associate
				eventsProcessing.rttlogData[rttlIndex].AssociatedRFIDItems = append(eventsProcessing.rttlogData[rttlIndex].AssociatedRFIDItems, &eventsProcessing.currentRFIDData[rfidIndex])
				eventsProcessing.currentRFIDData[rfidIndex].AssociatedRTTLEntry = &eventsProcessing.rttlogData[rttlIndex]
			}
		}

	}

}

func (eventsProcessing *EventsProcessor) isRFIDEligible(rttlogReading RTTLogEventEntry) bool {
	return rttlogReading.ProductDetails.RFIDEligible
}

func (eventsProcessing *EventsProcessor) convertProductIDTo14Char(productID string) string {
	formattedProductID := fmt.Sprintf("%014s", productID)
	return formattedProductID
}

func (eventsProcessing *EventsProcessor) atROILocation(name string, ROIs map[string]ROILocation) bool {
	location, ok := ROIs[name]
	if !ok {
		return false
	}
	return location.AtLocation
}

func (eventsProcessing *EventsProcessor) unmarshalObjValue(object interface{}, instance interface{}) error {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, instance)
	if err != nil {
		return err
	}
	return nil
}
