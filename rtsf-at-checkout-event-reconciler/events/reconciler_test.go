// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ScaleDrop(delta float64) *ScaleEventEntry {

	if len(ScaleData) == 0 {
		ScaleData = append(ScaleData, ScaleEventEntry{EventTime: int64(time.Now().Nanosecond()), Delta: delta, Total: delta})
	} else {
		ScaleData = append(ScaleData, ScaleEventEntry{EventTime: int64(time.Now().Nanosecond()), Delta: delta, Total: ScaleData[len(ScaleData)-1].Total + delta})
	}
	return &ScaleData[len(ScaleData)-1]
}

func RTTLScanItemA(quantity float64) {
	previousRttlogData := RttlogData[len(RttlogData)-1]
	if previousRttlogData.ProductId == "123" {
		appendToPreviousPosItem(RTTLogEventEntry{ProductId: "123", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 10, ExpectedMaxWeight: 11}})
	} else {
		RttlogData = append(RttlogData, RTTLogEventEntry{ProductId: "123", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 10, ExpectedMaxWeight: 11}})
	}

}

func RTTLScanItemB(quantity float64) {
	previousRttlogData := RttlogData[len(RttlogData)-1]
	if previousRttlogData.ProductId == "999" {
		appendToPreviousPosItem(RTTLogEventEntry{ProductId: "999", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 50, ExpectedMaxWeight: 55}})
	} else {
		RttlogData = append(RttlogData, RTTLogEventEntry{ProductId: "999", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 50, ExpectedMaxWeight: 55}})
	}

}

func BasketOpen() {
	CurrentCVData = []CVEventEntry{}
	NextCVData = []CVEventEntry{}
	CurrentRFIDData = []RFIDEventEntry{}
	NextRFIDData = []RFIDEventEntry{}
	afterPaymentSuccess = false
	firstBasketOpenComplete = false
	RttlogData = []RTTLogEventEntry{}
	ScaleData = []ScaleEventEntry{}
	RttlogData = append(RttlogData, RTTLogEventEntry{})
	SuspectScaleItems = make(map[int64]*ScaleEventEntry)
	EventOccurred = make(map[string]bool)
}

func TestScaleBasketReconciliationDropDropScan4DropDrop(t *testing.T) {
	expectedTotalWeight := 0.0

	BasketOpen()

	scaleEvent1 := ScaleDrop(10.1)
	expectedTotalWeight = expectedTotalWeight + 10.1
	scaleBasketReconciliation(scaleEvent1)
	assert.Nil(t, scaleEvent1.AssociatedRTTLEntry)
	assert.Equal(t, len(SuspectScaleItems), 1)

	scaleEvent2 := ScaleDrop(10.5)
	expectedTotalWeight = expectedTotalWeight + 10.5
	scaleBasketReconciliation(scaleEvent2)
	assert.Equal(t, len(SuspectScaleItems), 2)
	assert.Equal(t, len(ScaleData), 2)
	assert.Equal(t, len(RttlogData), 1)

	RTTLScanItemA(4)
	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)
	assert.Equal(t, len(SuspectScaleItems), 2)

	scaleEvent3 := ScaleDrop(11)
	expectedTotalWeight = expectedTotalWeight + 11
	scaleBasketReconciliation(scaleEvent3)
	assert.NotNil(t, ScaleData[0].AssociatedRTTLEntry)
	assert.NotNil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.Equal(t, len(SuspectScaleItems), 0)
	assert.Equal(t, len(RttlogData[len(RttlogData)-1].AssociatedScaleItems), 3)
	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	scaleEvent4 := ScaleDrop(21) //outside of range, should not be consolidated with ItemA
	expectedTotalWeight = expectedTotalWeight + 21
	scaleBasketReconciliation(scaleEvent4)
	assert.Equal(t, len(SuspectScaleItems), 1)
	assert.Equal(t, len(RttlogData[len(RttlogData)-1].AssociatedScaleItems), 3)
	assert.Nil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.Equal(t, ScaleData[len(ScaleData)-1].Total, expectedTotalWeight)
	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	scaleEvent5 := ScaleDrop(10)
	expectedTotalWeight = expectedTotalWeight + 10
	scaleBasketReconciliation(scaleEvent5)
	assert.Equal(t, len(SuspectScaleItems), 1)
	assert.Equal(t, len(RttlogData[len(RttlogData)-1].AssociatedScaleItems), 4)
	assert.NotNil(t, ScaleData[0].AssociatedRTTLEntry)
	assert.NotNil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.Equal(t, ScaleData[len(ScaleData)-1].Total, expectedTotalWeight)
	assert.True(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

}

func TestScaleBasketReconciliationGroupDropScanDropMultipleItems(t *testing.T) {
	expectedTotalWeight := 0.0

	BasketOpen()

	scaleEvent1 := ScaleDrop(21) //group drop item A
	expectedTotalWeight = expectedTotalWeight + 21
	scaleBasketReconciliation(scaleEvent1)
	assert.Equal(t, len(SuspectScaleItems), 1)
	assert.Nil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	RTTLScanItemA(3)

	scaleEvent2 := ScaleDrop(11)
	expectedTotalWeight = expectedTotalWeight + 11
	scaleBasketReconciliation(scaleEvent2)
	assert.Equal(t, len(SuspectScaleItems), 0)
	assert.Equal(t, len(RttlogData[len(RttlogData)-1].AssociatedScaleItems), 2)
	assert.NotNil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.NotNil(t, ScaleData[0].AssociatedRTTLEntry)
	assert.Equal(t, ScaleData[len(ScaleData)-1].Total, expectedTotalWeight)
	assert.True(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	scaleEvent3 := ScaleDrop(106) //group drop item B
	expectedTotalWeight = expectedTotalWeight + 106
	scaleBasketReconciliation(scaleEvent3)
	assert.Equal(t, len(SuspectScaleItems), 1)
	assert.Nil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.True(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	RTTLScanItemB(3)

	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	scaleEvent4 := ScaleDrop(55)
	expectedTotalWeight = expectedTotalWeight + 55
	scaleBasketReconciliation(scaleEvent4)
	assert.Equal(t, len(SuspectScaleItems), 0)
	assert.Equal(t, len(RttlogData[len(RttlogData)-1].AssociatedScaleItems), 2)
	assert.NotNil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.NotNil(t, ScaleData[0].AssociatedRTTLEntry)
	assert.Equal(t, ScaleData[len(ScaleData)-1].Total, expectedTotalWeight)
	assert.True(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)
}

func TestScaleBasketReconciliationDropDropScanScanScanDrop(t *testing.T) {
	expectedTotalWeight := 0.0

	BasketOpen()

	scaleEvent1 := ScaleDrop(10.2)
	expectedTotalWeight = expectedTotalWeight + 10.2
	scaleBasketReconciliation(scaleEvent1)
	assert.Equal(t, len(SuspectScaleItems), 1)
	assert.Nil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	scaleEvent2 := ScaleDrop(10.5)
	expectedTotalWeight = expectedTotalWeight + 10.5
	scaleBasketReconciliation(scaleEvent2)
	assert.Equal(t, len(SuspectScaleItems), 2)
	assert.Nil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.False(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

	RTTLScanItemA(1)
	RTTLScanItemA(1)
	RTTLScanItemA(1)

	scaleEvent3 := ScaleDrop(10.9)
	expectedTotalWeight = expectedTotalWeight + 10.9
	scaleBasketReconciliation(scaleEvent3)
	assert.Equal(t, len(SuspectScaleItems), 0)
	assert.Equal(t, len(RttlogData[len(RttlogData)-1].AssociatedScaleItems), 3)
	assert.NotNil(t, ScaleData[len(ScaleData)-1].AssociatedRTTLEntry)
	assert.NotNil(t, ScaleData[0].AssociatedRTTLEntry)
	assert.Equal(t, ScaleData[len(ScaleData)-1].Total, expectedTotalWeight)
	assert.True(t, RttlogData[len(RttlogData)-1].ScaleConfirmed)

}

func TestCheckScaleConfirmedQuantityLbsScaleConfirmedPerfectMatch(t *testing.T) {

	associatedScaleItems := []*ScaleEventEntry{}
	scaleItem1 := ScaleEventEntry{Total: 1, Delta: 1}
	scaleItem2 := ScaleEventEntry{Total: 5, Delta: 4}
	scaleItem3 := ScaleEventEntry{Total: 8, Delta: 3}
	associatedScaleItems = append(associatedScaleItems, &scaleItem1)
	associatedScaleItems = append(associatedScaleItems, &scaleItem2)
	associatedScaleItems = append(associatedScaleItems, &scaleItem3)

	rttlEntry := RTTLogEventEntry{
		ScaleConfirmed:       false,
		Quantity:             8,
		QuantityUnit:         "lbs",
		AssociatedScaleItems: associatedScaleItems,
	}

	assert.True(t, checkScaleConfirmed(&rttlEntry))
	assert.True(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleConfirmedRttlUnderWeight(t *testing.T) {
	associatedScaleItems := []*ScaleEventEntry{}
	scaleItem1 := ScaleEventEntry{Total: 1, Delta: 1}
	scaleItem2 := ScaleEventEntry{Total: 5, Delta: 4}
	scaleItem3 := ScaleEventEntry{Total: 8.15, Delta: 3.15} // ~ 1.9% diff of 8
	associatedScaleItems = append(associatedScaleItems, &scaleItem1)
	associatedScaleItems = append(associatedScaleItems, &scaleItem2)
	associatedScaleItems = append(associatedScaleItems, &scaleItem3)

	rttlEntry := RTTLogEventEntry{
		ScaleConfirmed:       false,
		Quantity:             8,
		QuantityUnit:         "lbs",
		AssociatedScaleItems: associatedScaleItems,
	}

	assert.True(t, checkScaleConfirmed(&rttlEntry))
	assert.True(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleConfirmedRttlOverWeight(t *testing.T) {
	associatedScaleItems := []*ScaleEventEntry{}
	scaleItem1 := ScaleEventEntry{Total: 1, Delta: 1}
	scaleItem2 := ScaleEventEntry{Total: 5, Delta: 4}
	scaleItem3 := ScaleEventEntry{Total: 8, Delta: 3} // ~ 1.9% diff of 8
	associatedScaleItems = append(associatedScaleItems, &scaleItem1)
	associatedScaleItems = append(associatedScaleItems, &scaleItem2)
	associatedScaleItems = append(associatedScaleItems, &scaleItem3)

	rttlEntry := RTTLogEventEntry{
		ScaleConfirmed:       false,
		Quantity:             8.15,
		QuantityUnit:         "lbs",
		AssociatedScaleItems: associatedScaleItems,
	}

	assert.True(t, checkScaleConfirmed(&rttlEntry))
	assert.True(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleNotConfirmed(t *testing.T) {
	associatedScaleItems := []*ScaleEventEntry{}
	scaleItem1 := ScaleEventEntry{Total: 1, Delta: 1}
	scaleItem2 := ScaleEventEntry{Total: 5, Delta: 4}
	scaleItem3 := ScaleEventEntry{Total: 8, Delta: 3}
	associatedScaleItems = append(associatedScaleItems, &scaleItem1)
	associatedScaleItems = append(associatedScaleItems, &scaleItem2)
	associatedScaleItems = append(associatedScaleItems, &scaleItem3)

	rttlEntry := RTTLogEventEntry{
		ScaleConfirmed:       false,
		Quantity:             15,
		QuantityUnit:         "lbs",
		AssociatedScaleItems: associatedScaleItems,
	}

	assert.False(t, checkScaleConfirmed(&rttlEntry))
	assert.False(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedOverpopulatedRTTL(t *testing.T) {
	associatedScaleItems := []*ScaleEventEntry{}
	scaleItem1 := ScaleEventEntry{Total: 10, Delta: 10}
	scaleItem2 := ScaleEventEntry{Total: 20.5, Delta: 10.5}
	scaleItem3 := ScaleEventEntry{Total: 31.5, Delta: 11}
	scaleItem4 := ScaleEventEntry{Total: 41.6, Delta: 10.1}
	associatedScaleItems = append(associatedScaleItems, &scaleItem1)
	associatedScaleItems = append(associatedScaleItems, &scaleItem2)
	associatedScaleItems = append(associatedScaleItems, &scaleItem3)
	associatedScaleItems = append(associatedScaleItems, &scaleItem4)

	rttlEntry := RTTLogEventEntry{
		ScaleConfirmed:       false,
		Quantity:             3,
		QuantityUnit:         "EA",
		AssociatedScaleItems: associatedScaleItems,
		ProductDetails:       ProductDetails{ExpectedMinWeight: 10, ExpectedMaxWeight: 11},
	}

	//4 associated scale drops, RTTL entry only contains 3 units. Pop one out and then set true for scaleConfirmed
	assert.True(t, checkScaleConfirmed(&rttlEntry))
	assert.Equal(t, len(rttlEntry.AssociatedScaleItems), 3)
}

func TestCVBasketReconciliation(t *testing.T) {

	BasketOpen()
	CurrentCVData = []CVEventEntry{}

	cvEntry := CVEventEntry{
		ObjectName: "123",
		ROIs: map[string]ROILocation{
			ScannerROI: {
				LastAtLocation: 1559679684,
			},
		},
	}
	CurrentCVData = append(CurrentCVData, cvEntry)

	posEntry := RTTLogEventEntry{
		ProductName: "123",
		EventTime:   1559679673,
		Quantity:    1,
	}

	cvBasketReconciliation(&posEntry)

	assert.NotNil(t, CurrentCVData[0].AssociatedRTTLEntry)
	assert.Equal(t, len(posEntry.AssociatedCVItems), 1)
	assert.True(t, posEntry.CVConfirmed)
}

func TestRFIDBasketReconciliation(t *testing.T) {

	BasketOpen()

	CurrentRFIDData = []RFIDEventEntry{RFIDEventEntry{
		EPC: "301400000047DAC000003039",
		UPC: "00000000735797",
	}}

	rttlEntry := RTTLogEventEntry{
		ProductId: "00000000735797",
		Quantity:  1,
	}
	rfidBasketReconciliation(&rttlEntry)

	assert.NotNil(t, CurrentRFIDData[0].AssociatedRTTLEntry)
	assert.Equal(t, len(rttlEntry.AssociatedRFIDItems), 1)
	assert.True(t, rttlEntry.RFIDConfirmed)
}
