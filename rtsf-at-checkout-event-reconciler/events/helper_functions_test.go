// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initRttlogData(p *EventsProcessor) {
	p.rttlogData = []RTTLogEventEntry{
		{
			ProductId:  "12345",
			Quantity:   1,
			Collection: []RTTLogEventEntry{},
		},
	}
}

func TestAppendToPreviousPosItem_Once(t *testing.T) {
	processor := &EventsProcessor{}
	rttlogItem := RTTLogEventEntry{
		ProductId:  "12345",
		Quantity:   2,
		Collection: []RTTLogEventEntry{},
	}

	initRttlogData(processor)
	processor.appendToPreviousPosItem(rttlogItem)
	assert.Equal(t, 3.0, processor.rttlogData[len(processor.rttlogData)-1].Quantity)
	assert.Equal(t, 2, len(processor.rttlogData[len(processor.rttlogData)-1].Collection))
}

func TestAppendToPreviousPosItem_Multiple(t *testing.T) {
	processor := &EventsProcessor{}
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

	initRttlogData(processor)
	processor.appendToPreviousPosItem(rttlogItem)

	processor.appendToPreviousPosItem(rttlogItem2)
	assert.Equal(t, 8.0, processor.rttlogData[len(processor.rttlogData)-1].Quantity)
	assert.Equal(t, 3, len(processor.rttlogData[len(processor.rttlogData)-1].Collection))
}

func initScaleData(p *EventsProcessor) {
	p.scaleData = []ScaleEventEntry{
		{
			Delta: 10,
			Total: 10,
		},
	}
}
func TestCalculateScaleDelta(t *testing.T) {
	processor := &EventsProcessor{}
	scaleItem := ScaleEventEntry{
		Total: 12,
	}

	initScaleData(processor)
	processor.calculateScaleDelta(&scaleItem)
	assert.Equal(t, 2.0, scaleItem.Delta)
}

func initRemoveItem(p *EventsProcessor) {
	p.rttlogData = []RTTLogEventEntry{
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
	processor := &EventsProcessor{}
	initRemoveItem(processor)
	removeCat := RTTLogEventEntry{

		ProductId: "cat",
		Quantity:  5.0,
	}
	processor.removeRTTLItemFromBuffer(removeCat)
	assert.Equal(t, 2, len(processor.rttlogData))

}

func TestRemoveRTTLItemFromBufferNoCollectionExactQuantityRemoved(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItem(processor)
	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  5.0,
	}

	processor.removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 6.0, processor.rttlogData[0].Quantity)
	assert.Equal(t, 1, len(processor.rttlogData))

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  6.0,
	}

	processor.removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 0, len(processor.rttlogData))

}

func TestRemoveRTTLItemFromBufferNoCollectionSmallerQuantityRemoved(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItem(processor)

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  5.0,
	}

	processor.removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 5.0, processor.rttlogData[0].Quantity)
	assert.Equal(t, 1.0, processor.rttlogData[1].Quantity) //should only have 1 apple left

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  1.0,
	}

	processor.removeRTTLItemFromBuffer(removePear)
	assert.Equal(t, 4.0, processor.rttlogData[0].Quantity)

}

func initRemoveItemCollection(p *EventsProcessor) {
	p.rttlogData = []RTTLogEventEntry{
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
	processor := &EventsProcessor{}
	initRemoveItemCollection(processor)

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  1.0,
	}

	processor.removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 3.0, processor.rttlogData[0].Collection[0].Quantity)
	assert.Equal(t, 4.0, processor.rttlogData[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionFirstCollectionIndexLessAmount(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItemCollection(processor)

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  4.0,
	}

	processor.removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 1.0, processor.rttlogData[0].Quantity)
	assert.Equal(t, 1.0, processor.rttlogData[0].Collection[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionSecondCollectionIndexGreaterAmount(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItemCollection(processor)

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  3.0,
	}

	processor.removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 2.0, processor.rttlogData[0].Quantity)
	assert.Equal(t, 1.0, processor.rttlogData[0].Collection[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionFirstIndexRemoveAll(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItemCollection(processor)

	removePear := RTTLogEventEntry{

		ProductId: "pear",
		Quantity:  5.0,
	}

	processor.removeRTTLItemFromBuffer(removePear)

	assert.Equal(t, 6.0, processor.rttlogData[0].Quantity)
	assert.Equal(t, "apple", processor.rttlogData[0].ProductId)

}

func TestRemoveRTTLItemFromBufferWithCollectionLastIndexLessAmount(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItemCollection(processor)

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  5.0,
	}

	processor.removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 1.0, processor.rttlogData[1].Quantity)
	assert.Equal(t, 1.0, processor.rttlogData[1].Collection[0].Quantity)

}

func TestRemoveRTTLItemFromBufferWithCollectionLastIndexRemoveExactAmount(t *testing.T) {
	processor := &EventsProcessor{}
	initRemoveItemCollection(processor)

	removeApple := RTTLogEventEntry{

		ProductId: "apple",
		Quantity:  6.0,
	}

	processor.removeRTTLItemFromBuffer(removeApple)

	assert.Equal(t, 1, len(processor.rttlogData))
	assert.Equal(t, 3.0, processor.rttlogData[0].Collection[1].Quantity)

}

func TestDeleteLastScaleItem(t *testing.T) {
	scaleEntry := ScaleEventEntry{Total: 2}
	scaleEntry2 := ScaleEventEntry{Total: 3}
	scaleData := []*ScaleEventEntry{&scaleEntry, &scaleEntry2}

	processor := &EventsProcessor{}

	assert.Equal(t, len(scaleData), 2)
	processor.deleteLastScaleItem(&scaleData)
	assert.Equal(t, len(scaleData), 1)
	processor.deleteLastScaleItem(&scaleData)
	assert.Equal(t, len(scaleData), 0)
}

func TestResetRTTLBasket(t *testing.T) {
	processor := &EventsProcessor{}

	processor.resetRTTLBasket()

	initScaleData(processor)
	initRttlogData(processor)

	exampleSuspect := ScaleEventEntry{}
	processor.suspectScaleItems[3] = &exampleSuspect

	assert.Equal(t, len(processor.scaleData), 1)
	assert.Equal(t, len(processor.rttlogData), 1)
	assert.Equal(t, len(processor.suspectScaleItems), 1)

	processor.resetRTTLBasket()

	assert.Equal(t, len(processor.scaleData), 0)
	assert.Equal(t, len(processor.rttlogData), 0)
	assert.Equal(t, len(processor.suspectScaleItems), 0)

}

func TestCheckRTTLForPOSItems(t *testing.T) {
	processor := &EventsProcessor{}
	processor.resetRTTLBasket()
	assert.False(t, processor.checkRTTLForPOSItems())

	rttlEvent := RTTLogEventEntry{Quantity: 3}
	processor.rttlogData = append(processor.rttlogData, rttlEvent)
	assert.True(t, processor.checkRTTLForPOSItems())

	processor.removeRTTLItemFromBuffer(rttlEvent)
	assert.False(t, processor.checkRTTLForPOSItems())

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
	processor := &EventsProcessor{}
	processor.currentCVData = []CVEventEntry{}

	processor.currentCVData = append(processor.currentCVData, initSuspectCVItems())
	processor.currentCVData = append(processor.currentCVData, initSuspectCVItems())

	suspectItems := processor.getSuspectCVItems()
	assert.Equal(t, len(suspectItems), 2)
}

func TestGetSuspectRFIDItems(t *testing.T) {
	processor := &EventsProcessor{}
	processor.currentRFIDData = []RFIDEventEntry{}

	processor.currentRFIDData = append(processor.currentRFIDData, initSuspectRFIDItems())
	processor.currentRFIDData = append(processor.currentRFIDData, initSuspectRFIDItems())
	processor.currentRFIDData = append(processor.currentRFIDData, initNonSuspectRFIDItems())
	processor.currentRFIDData = append(processor.currentRFIDData, initGoBackRFIDItems())

	suspectItems := processor.getSuspectRFIDItems()
	assert.Equal(t, len(suspectItems), 2)
}

func TestResetCVBasket(t *testing.T) {
	processor := &EventsProcessor{}
	processor.currentCVData = []CVEventEntry{}
	processor.nextCVData = []CVEventEntry{}

	processor.currentCVData = append(processor.currentCVData, initSuspectCVItems())
	processor.currentCVData = append(processor.currentCVData, initNonSuspectCVItems())
	processor.currentCVData = append(processor.currentCVData, initGoBackCVItems())
	processor.currentCVData = append(processor.currentCVData, initGoBackCVItems())

	assert.Equal(t, len(processor.currentCVData), 4)

	processor.resetCVBasket()

	assert.Equal(t, len(processor.currentCVData), 3)
}

func TestResetRFIDBasket(t *testing.T) {
	processor := &EventsProcessor{}
	processor.currentRFIDData = []RFIDEventEntry{}
	processor.nextRFIDData = []RFIDEventEntry{}

	processor.currentRFIDData = append(processor.currentRFIDData, initSuspectRFIDItems())
	processor.currentRFIDData = append(processor.currentRFIDData, initGoBackRFIDItems())
	processor.currentRFIDData = append(processor.currentRFIDData, initNonSuspectRFIDItems())
	processor.currentRFIDData = append(processor.currentRFIDData, initNonSuspectRFIDItems())

	assert.Equal(t, len(processor.currentRFIDData), 4)

	processor.resetRFIDBasket()

	assert.Equal(t, len(processor.currentRFIDData), 2)

}

func TestUpdateSuspectRFIDItems(t *testing.T) {
	processor := &EventsProcessor{}
	processor.currentRFIDData = []RFIDEventEntry{}
	processor.rttlogData = []RTTLogEventEntry{}

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

	processor.currentRFIDData = append(processor.currentRFIDData, rfidRedApplesEntry)
	processor.rttlogData = append(processor.rttlogData, basketOpenEntry)
	processor.rttlogData = append(processor.rttlogData, RedApplesPOSEntry)

	assert.Equal(t, len(processor.rttlogData[1].AssociatedRFIDItems), 0)
	assert.Nil(t, processor.currentRFIDData[0].AssociatedRTTLEntry)

	processor.updateSuspectRFIDItems()
	assert.Equal(t, len(processor.rttlogData[1].AssociatedRFIDItems), 1)
	assert.NotNil(t, processor.currentRFIDData[0].AssociatedRTTLEntry)

}

func TestConvertProductIDTo14Char(t *testing.T) {
	processor := &EventsProcessor{}
	productID := "123456789"
	newProductID := processor.convertProductIDTo14Char(productID)

	assert.Equal(t, len(newProductID), 14)
}

func TestEventsProcessor_unmarshalDtosObj(t *testing.T) {
	tests := []struct {
		name     string
		dataObj  RTTLogEventEntry
		instance interface{}
		wantErr  bool
	}{
		{
			name: "test1",
			dataObj: RTTLogEventEntry{
				ProductId: "12345",
				Quantity:  12,
			},
			instance: &RTTLogEventEntry{},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventsProcessing := &EventsProcessor{}
			object := dtos.NewObjectReading("", "", "", &tt.dataObj)
			testJson, err := json.Marshal(object)
			require.NoError(t, err)
			testStruct := &dtos.BaseReading{}
			err = json.Unmarshal(testJson, testStruct)
			require.NoError(t, err)
			if err := eventsProcessing.unmarshalObjValue(testStruct.ObjectReading.ObjectValue, tt.instance); (err != nil) != tt.wantErr {
				t.Errorf("EventsProcessor.unmarshalDtosObj() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
