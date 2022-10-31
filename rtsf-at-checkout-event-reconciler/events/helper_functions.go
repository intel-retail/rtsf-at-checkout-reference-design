// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"fmt"
	"math"
)

func calculateScaleDelta(scaleReading *ScaleEventEntry) {
	if len(ScaleData) == 0 {
		scaleReading.Delta = scaleReading.Total
	} else {
		previousReading := ScaleData[len(ScaleData)-1]
		scaleReading.Delta = scaleReading.Total - previousReading.Total
	}
}

// consolidate the previous and current POS item into same RTTLogData entry
func appendToPreviousPosItem(newItem RTTLogEventEntry) {
	if len(RttlogData) == 0 {
		return
	}
	previousItem := RttlogData[len(RttlogData)-1]
	if len(previousItem.Collection) == 0 {
		// add the exisiting item to the collection list
		previousItem.Collection = append(previousItem.Collection, RttlogData[len(RttlogData)-1])
	}

	// add the new item to the existing collection
	previousItem.Collection = append(previousItem.Collection, newItem)
	previousItem.Quantity = previousItem.Quantity + newItem.Quantity

	RttlogData[len(RttlogData)-1] = previousItem
}
func deleteRTTLItemAtIndex(list *[]RTTLogEventEntry, index int) {
	if index == len(*list)-1 {
		// last item in list - gets deleted
		*list = (*list)[:len(*list)-1]
	} else if index < len(*list)-1 {
		//not last item
		*list = append((*list)[:index], (*list)[index+1:]...)
	}
}

func removeRTTLItemFromBuffer(rttlogReading RTTLogEventEntry) error {
	quantityToRemove := rttlogReading.Quantity
	for rttlogIndex, item := range RttlogData {
		if item.ProductId == rttlogReading.ProductId {
			if item.Quantity >= (quantityToRemove - floatingPointTolerance) {
				// this means that items were not collapsed into a collection
				if len(item.Collection) == 0 {
					if math.Abs(item.Quantity-quantityToRemove) <= floatingPointTolerance { //checking if quantity and quantitytoRemove are equal
						// remove item at that index from RttlogData
						deleteRTTLItemAtIndex(&RttlogData, rttlogIndex)
					} else {
						item.Quantity = item.Quantity - quantityToRemove
						RttlogData[rttlogIndex] = item
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
						deleteRTTLItemAtIndex(&item.Collection, 0)
						RttlogData[rttlogIndex].Quantity = RttlogData[rttlogIndex].Quantity - collectionItem.Quantity
						continue
					}
					//this means the collection item has more than we want to remove
					collectionItem.Quantity = collectionItem.Quantity - quantityToRemove
					RttlogData[rttlogIndex].Quantity = RttlogData[rttlogIndex].Quantity - quantityToRemove
					quantityToRemove = 0
					item.Collection[0] = collectionItem //store updated quantity back into collection

				}
				if len(item.Collection) == 0 {
					deleteRTTLItemAtIndex(&RttlogData, rttlogIndex)
				}

			}
		}
	}
	if quantityToRemove > floatingPointTolerance {
		return fmt.Errorf("Error: Remove item failed: %s\n", rttlogReading.ProductName)
	}
	return nil
}

func deleteLastScaleItem(list *[]*ScaleEventEntry) {
	if list == nil || len(*list) < 1 {
		return
	}

	*list = (*list)[:len(*list)-1]
}

func wrapSuspectItems() ([]byte, error) {
	suspectList := SuspectLists{
		CVSuspect:    getSuspectCVItems(),
		RFIDSuspect:  getSuspectRFIDItems(),
		ScaleSuspect: getSuspectScaleItems(),
	}

	byteSuspects, err := json.MarshalIndent(suspectList, "", "   ")
	if err != nil {
		return nil, err
	}

	return byteSuspects, nil
}

func resetRTTLBasket() {
	RttlogData = []RTTLogEventEntry{}
	ScaleData = []ScaleEventEntry{}
	SuspectScaleItems = make(map[int64]*ScaleEventEntry)
}

func resetRFIDBasket() {
	persistRFIDGoBack()
	persistRFIDSuspectItems()
	CurrentRFIDData = NextRFIDData
	NextRFIDData = []RFIDEventEntry{}
	afterPaymentSuccess = false
}

func persistRFIDGoBack() {
	for _, rfidItem := range CurrentRFIDData {
		if atROILocation(GoBackROI, rfidItem.ROIs) {
			NextRFIDData = append(NextRFIDData, rfidItem)
		}
	}
}

func persistRFIDSuspectItems() {
	for _, rfidItem := range CurrentRFIDData {
		if rfidItem.AssociatedRTTLEntry == nil && !atROILocation(GoBackROI, rfidItem.ROIs) && !atROILocation(EntranceROI, rfidItem.ROIs) {
			NextRFIDData = append(NextRFIDData, rfidItem)
		}
	}
}

func resetCVBasket() {
	persistCVGoBack()
	persistCVSuspectItems()
	CurrentCVData = NextCVData
	NextCVData = []CVEventEntry{}
	afterPaymentSuccess = false
}

func persistCVGoBack() {
	for _, cvItem := range CurrentCVData {
		if atROILocation(GoBackROI, cvItem.ROIs) {
			NextCVData = append(NextCVData, cvItem)
		}
	}
}

func persistCVSuspectItems() {
	for _, cvItem := range CurrentCVData {
		if cvItem.AssociatedRTTLEntry == nil && !atROILocation(GoBackROI, cvItem.ROIs) && !atROILocation(EntranceROI, cvItem.ROIs) {
			NextCVData = append(NextCVData, cvItem)
		}
	}
}

func checkRTTLForPOSItems() bool {
	for _, rttlEvent := range RttlogData {
		if rttlEvent.Quantity > floatingPointTolerance {
			return true
		}
	}
	return false
}

func getExistingCVDataByObjectName(cvReading CVEventEntry) *CVEventEntry {
	for cvIndex, cvItem := range CurrentCVData {
		if cvReading.ObjectName == cvItem.ObjectName {
			return &CurrentCVData[cvIndex]
		}
	}
	return nil
}

func getExistingRFIDDataByEPC(rfidReading RFIDEventEntry) *RFIDEventEntry {
	for rfidIndex, rfidItem := range CurrentRFIDData {
		if rfidReading.EPC == rfidItem.EPC {
			return &CurrentRFIDData[rfidIndex]
		}
	}
	return nil
}

func getSuspectScaleItems() map[int64]*ScaleEventEntry {

	scaleEntries := make(map[int64]*ScaleEventEntry)

	for key, scaleItem := range SuspectScaleItems {
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

func getSuspectCVItems() []CVEventEntry {
	suspectItems := []CVEventEntry{}
	for _, cvItem := range CurrentCVData {
		if cvItem.AssociatedRTTLEntry == nil && !atROILocation(GoBackROI, cvItem.ROIs) && !atROILocation(EntranceROI, cvItem.ROIs) {
			suspectItems = append(suspectItems, cvItem)
		}
	}
	return suspectItems
}

func getSuspectRFIDItems() []RFIDEventEntry {
	suspectItems := []RFIDEventEntry{}
	for _, rfidItem := range CurrentRFIDData {
		if rfidItem.AssociatedRTTLEntry == nil && !atROILocation(GoBackROI, rfidItem.ROIs) && !atROILocation(EntranceROI, rfidItem.ROIs) {
			suspectItems = append(suspectItems, rfidItem)
		}
	}
	return suspectItems
}

func updateSuspectRFIDItems() {
	for rfidIndex, rfidItem := range CurrentRFIDData {
		if CurrentRFIDData[rfidIndex].AssociatedRTTLEntry != nil ||
			atROILocation(GoBackROI, CurrentRFIDData[rfidIndex].ROIs) ||
			atROILocation(EntranceROI, CurrentRFIDData[rfidIndex].ROIs) {
			continue
		}

		for rttlIndex, rttlItem := range RttlogData {
			if !isRFIDEligible(rttlItem) {
				continue
			}
			//wont be a fractional quantity if its rfid-eligible - it will be "EA"
			if int(rttlItem.Quantity) > len(rttlItem.AssociatedRFIDItems) && rttlItem.ProductId == rfidItem.UPC {
				//cross-associate
				RttlogData[rttlIndex].AssociatedRFIDItems = append(RttlogData[rttlIndex].AssociatedRFIDItems, &CurrentRFIDData[rfidIndex])
				CurrentRFIDData[rfidIndex].AssociatedRTTLEntry = &RttlogData[rttlIndex]
			}
		}

	}

}

func isRFIDEligible(rttlogReading RTTLogEventEntry) bool {
	return rttlogReading.ProductDetails.RFIDEligible
}

func convertProductIDTo14Char(productID string) string {
	formattedProductID := fmt.Sprintf("%014s", productID)
	return formattedProductID
}

func atROILocation(name string, ROIs map[string]ROILocation) bool {
	location, ok := ROIs[name]
	if !ok {
		return false
	}
	return location.AtLocation
}
