// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func initRttlogData() {
	RttlogData = []RTTLogEventEntry{
		{
			ProductId:  "12345",
			Quantity:   1,
			Collection: []RTTLogEventEntry{},
		},
	}
}

func TestAppendToPreviousPosItem_Once(t *testing.T) {
	rttlogItem := RTTLogEventEntry{
		ProductId:  "12345",
		Quantity:   2,
		Collection: []RTTLogEventEntry{},
	}

	initRttlogData()
	appendToPreviousPosItem(rttlogItem)
	assert.Equal(t, 3.0, RttlogData[len(RttlogData)-1].Quantity)
	assert.Equal(t, 2, len(RttlogData[len(RttlogData)-1].Collection))
}

func TestAppendToPreviousPosItem_Multiple(t *testing.T) {
	rttlogItem := RTTLogEventEntry{
		ProductId:  "12345",
		Quantity:   2,
		Collection: []RTTLogEventEntry{},
	}

	rttlogItem2 := RTTLogEventEntry{
		ProductId:  "12345",
		Quantity:   5,
		Collection: []RTTLogEventEntry{},
	}

	initRttlogData()
	appendToPreviousPosItem(rttlogItem)

	appendToPreviousPosItem(rttlogItem2)
	assert.Equal(t, 8.0, RttlogData[len(RttlogData)-1].Quantity)
	assert.Equal(t, 3, len(RttlogData[len(RttlogData)-1].Collection))
}

func initScaleData() {
	ScaleData = []ScaleEventEntry{
		{
			Delta: 10,
			Total: 10,
		},
	}
}
func TestCalculateScaleDelta(t *testing.T) {
	scaleItem := ScaleEventEntry{
		Total: 12,
	}

	initScaleData()
	calculateScaleDelta(&scaleItem)
	assert.Equal(t, 2.0, scaleItem.Delta)
}

func initRemoveItem() {
	RttlogData = []RTTLogEventEntry{
		{
			ProductId: "pear",
			Quantity:  5.0,
		},
		{
			ProductId: "apple",
			Quantity:  6.0,
		},
	}
}
func TestRemoveRTTLItemFromBufferWrongItem(t *testing.T) {
	initRemoveItem()
	removeCat := RTTLogEventEntry{

		ProductId: "cat",
		Quantity:  5.0,
	}
	removeRTTLItemFromBuffer(removeCat)
	assert.Equal(t, 2, len(RttlogData))

}

func TestRemoveRTTLItemFromBufferNoCollectionExactQuantityRemoved(t *testing.T) {
	initRemoveItem()

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  5.0,
	}

	removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 6.0, RttlogData[0].Quantity)
	assert.Equal(t, 1, len(RttlogData))

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  6.0,
	}

	removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 0, len(RttlogData))

}

func TestRemoveRTTLItemFromBufferNoCollectionSmallerQuantityRemoved(t *testing.T) {
	initRemoveItem()

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  5.0,
	}

	removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 5.0, RttlogData[0].Quantity)
	assert.Equal(t, 1.0, RttlogData[1].Quantity) //should only have 1 apple left

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  1.0,
	}

	removeRTTLItemFromBuffer(removePear)
	assert.Equal(t, 4.0, RttlogData[0].Quantity)

}

func initRemoveItemCollection() {
	RttlogData = []RTTLogEventEntry{
		{
			ProductId: "pear",
			Quantity:  5.0,
			Collection: []RTTLogEventEntry{
				{
					ProductId: "pear",
					Quantity:  1.0,
				},
				{
					ProductId: "pear",
					Quantity:  3.0,
				},
				{
					ProductId: "pear",
					Quantity:  1.0,
				},
			},
		},
		{
			ProductId: "apple",
			Quantity:  6.0,
			Collection: []RTTLogEventEntry{
				{
					ProductId: "apple",
					Quantity:  5.0,
				},
				{
					ProductId: "apple",
					Quantity:  1.0,
				},
			},
		},
	}
}

func TestRemoveRTTLItemFromBufferWithCollectionFirstCollectionIndexExactAmount(t *testing.T) {
	initRemoveItemCollection()

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  1.0,
	}

	removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 3.0, RttlogData[0].Collection[0].Quantity)
	assert.Equal(t, 4.0, RttlogData[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionFirstCollectionIndexLessAmount(t *testing.T) {
	initRemoveItemCollection()

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  4.0,
	}

	removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 1.0, RttlogData[0].Quantity)
	assert.Equal(t, 1.0, RttlogData[0].Collection[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionSecondCollectionIndexGreaterAmount(t *testing.T) {
	initRemoveItemCollection()

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  3.0,
	}

	removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 2.0, RttlogData[0].Quantity)
	assert.Equal(t, 1.0, RttlogData[0].Collection[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionFirstIndexRemoveAll(t *testing.T) {
	initRemoveItemCollection()

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  5.0,
	}

	removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 6.0, RttlogData[0].Quantity)
	assert.Equal(t, "apple", RttlogData[0].ProductId)

}

func TestRemoveRTTLItemFromBufferWithCollectionLastIndexLessAmount(t *testing.T) {
	initRemoveItemCollection()

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  5.0,
	}

	removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 1.0, RttlogData[1].Quantity)
	assert.Equal(t, 1.0, RttlogData[1].Collection[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionLastIndexRemoveExactAmount(t *testing.T) {
	initRemoveItemCollection()

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  6.0,
	}

	removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 1, len(RttlogData))
	assert.Equal(t, 3.0, RttlogData[0].Collection[1].Quantity)

}

func TestDeleteLastScaleItem(t *testing.T) {
	scaleEntry := ScaleEventEntry{Total: 2}
	scaleEntry2 := ScaleEventEntry{Total: 3}
	scaleData := []*ScaleEventEntry{&scaleEntry, &scaleEntry2}

	assert.Equal(t, len(scaleData), 2)
	deleteLastScaleItem(&scaleData)
	assert.Equal(t, len(scaleData), 1)
	deleteLastScaleItem(&scaleData)
	assert.Equal(t, len(scaleData), 0)
}

func TestResetRTTLBasket(t *testing.T) {
	resetRTTLBasket()

	initScaleData()
	initRttlogData()

	exampleSuspect := ScaleEventEntry{}
	SuspectScaleItems[3] = &exampleSuspect

	assert.Equal(t, len(ScaleData), 1)
	assert.Equal(t, len(RttlogData), 1)
	assert.Equal(t, len(SuspectScaleItems), 1)

	resetRTTLBasket()

	assert.Equal(t, len(ScaleData), 0)
	assert.Equal(t, len(RttlogData), 0)
	assert.Equal(t, len(SuspectScaleItems), 0)

}

func TestCheckRTTLForPOSItems(t *testing.T) {
	resetRTTLBasket()
	assert.False(t, checkRTTLForPOSItems())

	rttlEvent := RTTLogEventEntry{Quantity: 3}
	RttlogData = append(RttlogData, rttlEvent)
	assert.True(t, checkRTTLForPOSItems())

	removeRTTLItemFromBuffer(rttlEvent)
	assert.False(t, checkRTTLForPOSItems())

}

func initSuspectRFIDItems() RFIDEventEntry {
	suspectItem := RFIDEventEntry{
		EPC: "Suspect",
	}
	return suspectItem
}

func initNonSuspectRFIDItems() RFIDEventEntry {
	nonSuspectItem := RFIDEventEntry{
		EPC: "Non-Suspect",
		ROIs: map[string]ROILocation{
			EntranceROI: {
				AtLocation: true,
			},
		},
	}
	return nonSuspectItem
}

func initGoBackRFIDItems() RFIDEventEntry {
	goBackItem := RFIDEventEntry{
		EPC: "Go Back",
		ROIs: map[string]ROILocation{
			GoBackROI: {
				AtLocation: true,
			},
		},
	}

	return goBackItem
}
func initSuspectCVItems() CVEventEntry {
	suspectItem := CVEventEntry{
		ObjectName: "Suspect",
	}
	return suspectItem
}

func initNonSuspectCVItems() CVEventEntry {
	nonSuspectItem := CVEventEntry{
		ObjectName: "Non-Suspect",
		ROIs: map[string]ROILocation{
			EntranceROI: {
				AtLocation: true,
			},
		},
	}
	return nonSuspectItem
}

func initGoBackCVItems() CVEventEntry {
	goBackItem := CVEventEntry{
		ObjectName: GoBackROI,
		ROIs: map[string]ROILocation{
			GoBackROI: {
				AtLocation: true,
			},
		},
	}

	return goBackItem
}

func TestGetSuspectCVItems(t *testing.T) {
	CurrentCVData = []CVEventEntry{}

	CurrentCVData = append(CurrentCVData, initSuspectCVItems())
	CurrentCVData = append(CurrentCVData, initSuspectCVItems())

	suspectItems := getSuspectCVItems()
	assert.Equal(t, len(suspectItems), 2)
}

func TestGetSuspectRFIDItems(t *testing.T) {
	CurrentRFIDData = []RFIDEventEntry{}

	CurrentRFIDData = append(CurrentRFIDData, initSuspectRFIDItems())
	CurrentRFIDData = append(CurrentRFIDData, initSuspectRFIDItems())
	CurrentRFIDData = append(CurrentRFIDData, initNonSuspectRFIDItems())
	CurrentRFIDData = append(CurrentRFIDData, initGoBackRFIDItems())

	suspectItems := getSuspectRFIDItems()
	assert.Equal(t, len(suspectItems), 2)
}

func TestResetCVBasket(t *testing.T) {
	CurrentCVData = []CVEventEntry{}
	NextCVData = []CVEventEntry{}

	CurrentCVData = append(CurrentCVData, initSuspectCVItems())
	CurrentCVData = append(CurrentCVData, initNonSuspectCVItems())
	CurrentCVData = append(CurrentCVData, initGoBackCVItems())
	CurrentCVData = append(CurrentCVData, initGoBackCVItems())

	assert.Equal(t, len(CurrentCVData), 4)

	resetCVBasket()

	assert.Equal(t, len(CurrentCVData), 3)
}

func TestResetRFIDBasket(t *testing.T) {
	CurrentRFIDData = []RFIDEventEntry{}
	NextRFIDData = []RFIDEventEntry{}

	CurrentRFIDData = append(CurrentRFIDData, initSuspectRFIDItems())
	CurrentRFIDData = append(CurrentRFIDData, initGoBackRFIDItems())
	CurrentRFIDData = append(CurrentRFIDData, initNonSuspectRFIDItems())
	CurrentRFIDData = append(CurrentRFIDData, initNonSuspectRFIDItems())

	assert.Equal(t, len(CurrentRFIDData), 4)

	resetRFIDBasket()

	assert.Equal(t, len(CurrentRFIDData), 2)

}

func TestUpdateSuspectRFIDItems(t *testing.T) {
	CurrentRFIDData = []RFIDEventEntry{}
	RttlogData = []RTTLogEventEntry{}

	rfidRedApplesEntry := RFIDEventEntry{
		EPC: "30140000001FB28000003039",
		UPC: "00000000324588",
	}

	basketOpenEntry := RTTLogEventEntry{
		ProductId: "basket open",
	}

	RedApplesPOSEntry := RTTLogEventEntry{
		ProductId: "00000000324588",
		Quantity:  2,
		ProductDetails: ProductDetails{
			RFIDEligible: true,
		},
	}

	CurrentRFIDData = append(CurrentRFIDData, rfidRedApplesEntry)
	RttlogData = append(RttlogData, basketOpenEntry)
	RttlogData = append(RttlogData, RedApplesPOSEntry)

	assert.Equal(t, len(RttlogData[1].AssociatedRFIDItems), 0)
	assert.Nil(t, CurrentRFIDData[0].AssociatedRTTLEntry)

	updateSuspectRFIDItems()
	assert.Equal(t, len(RttlogData[1].AssociatedRFIDItems), 1)
	assert.NotNil(t, CurrentRFIDData[0].AssociatedRTTLEntry)

}

func TestConvertProductIDTo14Char(t *testing.T) {
	productID := "123456789"
	newProductID := convertProductIDTo14Char(productID)

	assert.Equal(t, len(newProductID), 14)
}
