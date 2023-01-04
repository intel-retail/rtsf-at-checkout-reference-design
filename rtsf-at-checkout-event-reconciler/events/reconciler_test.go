// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"event-reconciler/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ScaleDrop(delta float64, processor *EventsProcessor) *ScaleEventEntry {

	if len(processor.scaleData) == 0 {
		processor.scaleData = append(processor.scaleData, ScaleEventEntry{EventTime: int64(time.Now().Nanosecond()), Delta: delta, Total: delta})
	} else {
		processor.scaleData = append(processor.scaleData, ScaleEventEntry{EventTime: int64(time.Now().Nanosecond()), Delta: delta, Total: processor.scaleData[len(processor.scaleData)-1].Total + delta})
	}
	return &processor.scaleData[len(processor.scaleData)-1]
}

func RTTLScanItemA(quantity float64, processor *EventsProcessor) {
	previousrttlogData := processor.rttlogData[len(processor.rttlogData)-1]
	if previousrttlogData.ProductId == "123" {
		processor.appendToPreviousPosItem(RTTLogEventEntry{ProductId: "123", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 10, ExpectedMaxWeight: 11}})
	} else {
		processor.rttlogData = append(processor.rttlogData, RTTLogEventEntry{ProductId: "123", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 10, ExpectedMaxWeight: 11}})
	}

}

func RTTLScanItemB(quantity float64, processor *EventsProcessor) {
	previousrttlogData := processor.rttlogData[len(processor.rttlogData)-1]
	if previousrttlogData.ProductId == "999" {
		processor.appendToPreviousPosItem(RTTLogEventEntry{ProductId: "999", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 50, ExpectedMaxWeight: 55}})
	} else {
		processor.rttlogData = append(processor.rttlogData, RTTLogEventEntry{ProductId: "999", Quantity: quantity, QuantityUnit: quantityUnitEA, ProductDetails: ProductDetails{ExpectedMinWeight: 50, ExpectedMaxWeight: 55}})
	}

}

func BasketOpen(processor *EventsProcessor) {
	processor.currentCVData = []CVEventEntry{}
	processor.nextCVData = []CVEventEntry{}
	processor.currentRFIDData = []RFIDEventEntry{}
	processor.nextRFIDData = []RFIDEventEntry{}
	processor.afterPaymentSuccess = false
	processor.firstBasketOpenComplete = false
	processor.rttlogData = []RTTLogEventEntry{}
	processor.scaleData = []ScaleEventEntry{}
	processor.rttlogData = append(processor.rttlogData, RTTLogEventEntry{})
	processor.suspectScaleItems = make(map[int64]*ScaleEventEntry)
	processor.eventOccurred = make(map[string]bool)
}

func TestScaleBasketReconciliationDropDropScan4DropDrop(t *testing.T) {
	expectedTotalWeight := 0.0
	eventsProcessing := EventsProcessor{}
	BasketOpen(&eventsProcessing)

	scaleEvent1 := ScaleDrop(10.1, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 10.1
	eventsProcessing.scaleBasketReconciliation(scaleEvent1)
	assert.Nil(t, scaleEvent1.AssociatedRTTLEntry)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)

	scaleEvent2 := ScaleDrop(10.5, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 10.5
	eventsProcessing.scaleBasketReconciliation(scaleEvent2)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 2)
	assert.Equal(t, len(eventsProcessing.scaleData), 2)
	assert.Equal(t, len(eventsProcessing.rttlogData), 1)

	RTTLScanItemA(4, &eventsProcessing)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 2)

	scaleEvent3 := ScaleDrop(11, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 11
	eventsProcessing.scaleBasketReconciliation(scaleEvent3)
	assert.NotNil(t, eventsProcessing.scaleData[0].AssociatedRTTLEntry)
	assert.NotNil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 0)
	assert.Equal(t, len(eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].AssociatedScaleItems), 3) // last rttl has 3 scale events assoc
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	scaleEvent4 := ScaleDrop(21, &eventsProcessing) //outside of range, should not be consolidated with ItemA
	expectedTotalWeight = expectedTotalWeight + 21
	eventsProcessing.scaleBasketReconciliation(scaleEvent4)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Equal(t, len(eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].AssociatedScaleItems), 3)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.Equal(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].Total, expectedTotalWeight)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	scaleEvent5 := ScaleDrop(10, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 10
	eventsProcessing.scaleBasketReconciliation(scaleEvent5)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Equal(t, len(eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].AssociatedScaleItems), 4)
	assert.NotNil(t, eventsProcessing.scaleData[0].AssociatedRTTLEntry)
	assert.NotNil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.Equal(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].Total, expectedTotalWeight)
	assert.True(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

}

func TestScaleBasketReconciliationGroupDropScanDropMultipleItems(t *testing.T) {
	expectedTotalWeight := 0.0
	eventsProcessing := EventsProcessor{}
	BasketOpen(&eventsProcessing)

	scaleEvent1 := ScaleDrop(21, &eventsProcessing) //group drop item A
	expectedTotalWeight = expectedTotalWeight + 21
	eventsProcessing.scaleBasketReconciliation(scaleEvent1)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	RTTLScanItemA(3, &eventsProcessing)

	scaleEvent2 := ScaleDrop(11, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 11
	eventsProcessing.scaleBasketReconciliation(scaleEvent2)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 0)
	assert.Equal(t, len(eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].AssociatedScaleItems), 2)
	assert.NotNil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.NotNil(t, eventsProcessing.scaleData[0].AssociatedRTTLEntry)
	assert.Equal(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].Total, expectedTotalWeight)
	assert.True(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	scaleEvent3 := ScaleDrop(106, &eventsProcessing) //group drop item B
	expectedTotalWeight = expectedTotalWeight + 106
	eventsProcessing.scaleBasketReconciliation(scaleEvent3)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.True(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	RTTLScanItemB(3, &eventsProcessing)

	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	scaleEvent4 := ScaleDrop(55, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 55
	eventsProcessing.scaleBasketReconciliation(scaleEvent4)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 0)
	assert.Equal(t, len(eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].AssociatedScaleItems), 2)
	assert.NotNil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.NotNil(t, eventsProcessing.scaleData[0].AssociatedRTTLEntry)
	assert.Equal(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].Total, expectedTotalWeight)
	assert.True(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)
}

func TestScaleBasketReconciliationDropDropScanScanScanDrop(t *testing.T) {
	expectedTotalWeight := 0.0
	eventsProcessing := EventsProcessor{}
	BasketOpen(&eventsProcessing)

	scaleEvent1 := ScaleDrop(10.2, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 10.2
	eventsProcessing.scaleBasketReconciliation(scaleEvent1)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	scaleEvent2 := ScaleDrop(10.5, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 10.5
	eventsProcessing.scaleBasketReconciliation(scaleEvent2)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 2)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

	RTTLScanItemA(1, &eventsProcessing)
	RTTLScanItemA(1, &eventsProcessing)
	RTTLScanItemA(1, &eventsProcessing)

	scaleEvent3 := ScaleDrop(10.9, &eventsProcessing)
	expectedTotalWeight = expectedTotalWeight + 10.9
	eventsProcessing.scaleBasketReconciliation(scaleEvent3)
	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 0)
	assert.Equal(t, len(eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].AssociatedScaleItems), 3)
	assert.NotNil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.NotNil(t, eventsProcessing.scaleData[0].AssociatedRTTLEntry)
	assert.Equal(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].Total, expectedTotalWeight)
	assert.True(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)

}

func TestScaleBasketReconciliationScanHeavyItem_DropLightItem(t *testing.T) {
	eventsProcessing := EventsProcessor{}
	BasketOpen(&eventsProcessing)

	RTTLScanItemA(1, &eventsProcessing) // item weighs 10lbs

	scaleEvent1 := ScaleDrop(5.0, &eventsProcessing)

	eventsProcessing.scaleBasketReconciliation(scaleEvent1)

	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)
}

func TestScaleBasketReconciliationScanLightItem_DropHeavyItem(t *testing.T) {
	eventsProcessing := EventsProcessor{}
	BasketOpen(&eventsProcessing)

	RTTLScanItemA(1, &eventsProcessing) // item weighs 10lbs

	scaleEvent1 := ScaleDrop(20.0, &eventsProcessing)

	eventsProcessing.scaleBasketReconciliation(scaleEvent1)
	// should produce a suspect item

	assert.Equal(t, len(eventsProcessing.suspectScaleItems), 1)
	assert.Nil(t, eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1].AssociatedRTTLEntry)
	assert.False(t, eventsProcessing.rttlogData[len(eventsProcessing.rttlogData)-1].ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleConfirmedPerfectMatch(t *testing.T) {
	eventsProcessing := EventsProcessor{
		processConfig: &config.ReconcilerConfig{
			ScaleToScaleTolerance: 0.02,
		},
	}
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

	assert.True(t, eventsProcessing.checkScaleConfirmed(&rttlEntry))
	assert.True(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleConfirmedRttlUnderWeight(t *testing.T) {
	eventsProcessing := EventsProcessor{
		processConfig: &config.ReconcilerConfig{
			ScaleToScaleTolerance: 0.02,
		},
	}
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

	assert.True(t, eventsProcessing.checkScaleConfirmed(&rttlEntry))
	assert.True(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleConfirmedRttlOverWeight(t *testing.T) {
	eventsProcessing := EventsProcessor{
		processConfig: &config.ReconcilerConfig{
			ScaleToScaleTolerance: 0.02,
		},
	}
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

	assert.True(t, eventsProcessing.checkScaleConfirmed(&rttlEntry))
	assert.True(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedQuantityLbsScaleNotConfirmed(t *testing.T) {
	eventsProcessing := EventsProcessor{
		processConfig: &config.ReconcilerConfig{
			ScaleToScaleTolerance: 0.02,
		},
	}
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

	assert.False(t, eventsProcessing.checkScaleConfirmed(&rttlEntry))
	assert.False(t, rttlEntry.ScaleConfirmed)
}

func TestCheckScaleConfirmedOverpopulatedRTTL(t *testing.T) {
	eventsProcessing := EventsProcessor{
		suspectScaleItems: map[int64]*ScaleEventEntry{},
	}
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
	assert.True(t, eventsProcessing.checkScaleConfirmed(&rttlEntry))
	assert.Equal(t, len(rttlEntry.AssociatedScaleItems), 3)
}

func TestCVBasketReconciliation(t *testing.T) {
	eventsProcessing := EventsProcessor{
		cvTimeAlignment: time.Second * 5,
	}
	BasketOpen(&eventsProcessing)

	cvEntry := CVEventEntry{
		ObjectName: "123",
		ROIs: map[string]ROILocation{
			ScannerROI: {
				LastAtLocation: 1559679684,
			},
		},
	}
	eventsProcessing.currentCVData = append(eventsProcessing.currentCVData, cvEntry)

	posEntry := RTTLogEventEntry{
		ProductName: "123",
		EventTime:   1559679673,
		Quantity:    1,
	}

	eventsProcessing.cvBasketReconciliation(&posEntry)

	assert.NotNil(t, eventsProcessing.currentCVData[0].AssociatedRTTLEntry)
	assert.Equal(t, len(posEntry.AssociatedCVItems), 1)
	assert.True(t, posEntry.CVConfirmed)
}

func TestRFIDBasketReconciliation(t *testing.T) {
	eventsProcessing := EventsProcessor{}
	BasketOpen(&eventsProcessing)

	eventsProcessing.currentRFIDData = []RFIDEventEntry{
		{
			EPC: "301400000047DAC000003039",
			UPC: "00000000735797",
		},
	}

	rttlEntry := RTTLogEventEntry{
		ProductId: "00000000735797",
		Quantity:  1,
	}
	eventsProcessing.rfidBasketReconciliation(&rttlEntry)

	assert.NotNil(t, eventsProcessing.currentRFIDData[0].AssociatedRTTLEntry)
	assert.Equal(t, len(rttlEntry.AssociatedRFIDItems), 1)
	assert.True(t, rttlEntry.RFIDConfirmed)
}
