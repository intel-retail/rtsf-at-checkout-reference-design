// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"event-reconciler/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	"github.com/stretchr/testify/assert"
)

var context interfaces.AppFunctionContext

type productInfoTest struct {
	Barcode      string  `json:"barcode"`
	Name         string  `json:"name"`
	MinWeight    float64 `json:"min_weight"`
	MaxWeight    float64 `json:"max_weight"`
	RfidEligible bool    `json:"rfid_eligible"`
}

func TestMain(m *testing.M) {

	lc := logger.NewMockClient()
	context = pkg.NewAppFuncContextForTest("app_functions_sdk_go", lc)

	os.Exit(m.Run())
}

func initCVReadingScannerENTER() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		cvRoiEvent,
		CVEventEntry{
			ObjectName: "item 1",
			ROIAction:  "ENTERED",
			EventTime:  1559679684,
			ROIName:    "Scanner",
			ROIs:       map[string]ROILocation{},
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initCVReadingScannerEXIT() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		cvRoiEvent,
		CVEventEntry{
			ObjectName: "item 1",
			ROIAction:  "EXITED",
			EventTime:  1559679684,
			ROIName:    "Scanner",
			ROIs:       map[string]ROILocation{},
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initCVReadingBaggingENTER() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		cvRoiEvent,
		CVEventEntry{
			ObjectName: "item 1",
			ROIAction:  "ENTERED",
			EventTime:  1559679784,
			ROIName:    "Bagging",
			ROIs:       map[string]ROILocation{},
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initCVReadingNewItemStagingENTER() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		cvRoiEvent,
		CVEventEntry{
			ObjectName: "Item 2",
			ROIAction:  "ENTERED",
			EventTime:  1559679834,
			ROIName:    "Staging",
			ROIs:       map[string]ROILocation{},
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initCVReadingNewItemEntranceENTER() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		cvRoiEvent,
		CVEventEntry{
			ObjectName: "Item 3",
			ROIAction:  "ENTERED",
			EventTime:  1559679834,
			ROIName:    "Entrance",
			ROIs:       map[string]ROILocation{},
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initScaleReadingScaleItem() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		scaleItemEvent,
		ScaleEventEntry{
			EventTime: 1559679665,
			LaneId:    "123",
			ScaleId:   "123",
			Total:     2,
			Units:     "lbs",
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initPosReadingBasketOpen() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		basketOpenEvent,
		RTTLogEventEntry{
			BasketId:   "abc-012345-def",
			CustomerId: "joe5",
			EmployeeId: "mary1",
			EventTime:  1559679584,
			LaneId:     "123",
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initPosReadingBasketClose() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		basketOpenEvent,
		RTTLogEventEntry{
			BasketId:   "abc-012345-def",
			CustomerId: "joe5",
			EmployeeId: "mary1",
			EventTime:  1559679789,
			LaneId:     "123",
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initPosReadingRemoveItem() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		removeItemEvent,
		RTTLogEventEntry{
			BasketId:      "abc-012345-def",
			CustomerId:    "joe5",
			EmployeeId:    "mary1",
			EventTime:     1559679672,
			LaneId:        "123",
			ProductId:     "00000000735797",
			ProductIdType: "UPC",
			ProductName:   "Steak",
			Quantity:      3,
			QuantityUnit:  "lbs",
			UnitPrice:     8.99,
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initPosReadingPosItemSteak() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		posItemEvent,
		RTTLogEventEntry{
			BasketId:      "abc-012345-def",
			CustomerId:    "joe5",
			EmployeeId:    "mary1",
			EventTime:     1559679673,
			LaneId:        "123",
			ProductId:     "00000000735797",
			ProductIdType: "UPC",
			ProductName:   "Steak",
			Quantity:      3,
			QuantityUnit:  "lbs",
			UnitPrice:     8.99,
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initPosReadingPaymentStart() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		paymentStartEvent,
		RTTLogEventEntry{
			BasketId:   "abc-012345-def",
			CustomerId: "joe5",
			EmployeeId: "mary1",
			EventTime:  1559679588,
			LaneId:     "123",
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initRFIDReadingApplesExitedBagging() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		rfidRoiEvent,
		RFIDEventEntry{
			EPC:       "30140000001FB28000003039",
			ROIName:   "Bagging",
			ROIAction: "EXITED",
			EventTime: 1562972496854,
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initRFIDReadingApplesEnterBagging() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		rfidRoiEvent,
		RFIDEventEntry{
			EPC:       "30140000001FB28000003039",
			ROIName:   "Bagging",
			ROIAction: "ENTERED",
			EventTime: 1562972496854,
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initMockRFIDItem(name string, barcode string) productInfoTest {

	pim := productInfoTest{
		Barcode:      barcode,
		Name:         name,
		MinWeight:    1.3,
		MaxWeight:    1.4,
		RfidEligible: true,
	}

	return pim
}

func initRFIDReadingSteakEnterBagging() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		rfidRoiEvent,
		RFIDEventEntry{
			EPC:       "301400000047DAC000003039",
			ROIName:   "Bagging",
			ROIAction: "ENTERED",
			EventTime: 1562972496854,
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func initRFIDReadingSalsaEnterBagging() dtos.BaseReading {
	reading := dtos.NewObjectReading(
		"",
		"",
		rfidRoiEvent,
		RFIDEventEntry{
			EPC:       "301400000051240000003039",
			ROIName:   "Bagging",
			ROIAction: "ENTERED",
			EventTime: 1562972496854,
		},
	)
	simulateJson, _ := json.Marshal(reading)
	simulateStruct := dtos.BaseReading{}
	_ = json.Unmarshal(simulateJson, &simulateStruct)

	return simulateStruct
}

func TestProcessDeviceRFIDReading(t *testing.T) {
	processor := &EventsProcessor{}
	processor.currentRFIDData = []RFIDEventEntry{}
	processor.nextRFIDData = []RFIDEventEntry{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var pim productInfoTest
		if strings.Contains("00000000324588", r.URL.EscapedPath()) {
			pim = initMockRFIDItem("Apples", "00000000324588")
		} else if strings.Contains("00000000735797", r.URL.EscapedPath()) {
			pim = initMockRFIDItem("Steak", "00000000735797")
		}
		bytes, err := json.Marshal(pim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	}))
	defer ts.Close()

	tsURL, _ := url.Parse(ts.URL)
	eventsProcessor := &EventsProcessor{
		processConfig: &config.ReconcilerConfig{
			ProductLookupEndpoint: tsURL.Hostname() + ":" + tsURL.Port(),
		},
	}
	lc := logger.MockLogger{}

	reading := initRFIDReadingApplesEnterBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)
	assert.Equal(t, len(eventsProcessor.currentRFIDData), 1)
	assert.Contains(t, eventsProcessor.currentRFIDData[0].ROIs, BaggingROI)
	assert.True(t, eventsProcessor.currentRFIDData[0].ROIs[BaggingROI].AtLocation)

	reading = initRFIDReadingApplesExitedBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)
	assert.Equal(t, len(eventsProcessor.currentRFIDData), 1)
	assert.False(t, eventsProcessor.currentRFIDData[0].ROIs[BaggingROI].AtLocation)

	reading = initRFIDReadingSteakEnterBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)
	assert.Equal(t, len(eventsProcessor.currentRFIDData), 2)
	assert.Equal(t, len(eventsProcessor.nextRFIDData), 0)
	assert.True(t, eventsProcessor.currentRFIDData[1].ROIs[BaggingROI].AtLocation)

	eventsProcessor.afterPaymentSuccess = true

	reading = initRFIDReadingSalsaEnterBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)

	assert.Equal(t, len(eventsProcessor.currentRFIDData), 2)
	assert.Equal(t, len(eventsProcessor.nextRFIDData), 1)

	eventsProcessor.afterPaymentSuccess = false
}

func TestProcessDeviceCVReading(t *testing.T) {
	lc := logger.NewMockClient()
	eventsProcessor := &EventsProcessor{}

	eventsProcessor.currentCVData = []CVEventEntry{}
	eventsProcessor.nextCVData = []CVEventEntry{}

	reading := initCVReadingScannerENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(eventsProcessor.currentCVData), 1)

	reading = initCVReadingScannerEXIT()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(eventsProcessor.currentCVData), 1)

	reading = initCVReadingBaggingENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(eventsProcessor.currentCVData), 1)

	reading = initCVReadingNewItemStagingENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(eventsProcessor.currentCVData), 2)

	eventsProcessor.afterPaymentSuccess = true

	reading = initCVReadingNewItemEntranceENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)
	assert.Equal(t, len(eventsProcessor.currentCVData), 2)
	assert.Equal(t, len(eventsProcessor.nextCVData), 1)

	eventsProcessor.afterPaymentSuccess = false
}

func TestProcessDeviceScaleReading(t *testing.T) {
	eventsProcessor := &EventsProcessor{}
	eventsProcessor.resetRTTLBasket()

	lc := logger.NewMockClient()

	reading := initScaleReadingScaleItem()
	eventsProcessor.processDeviceScaleReading(reading, lc)
	assert.Equal(t, len(eventsProcessor.scaleData), 1)
	assert.Equal(t, len(eventsProcessor.suspectScaleItems), 1)
	assert.Equal(t, eventsProcessor.scaleData[0].Total, 2.0)
	assert.Equal(t, eventsProcessor.scaleData[0].Units, "lbs")
}

func TestProcessDevicePosReading(t *testing.T) {
	eventsProcessor := EventsProcessor{}
	BasketOpen(&eventsProcessor)
	eventsProcessor.resetRTTLBasket()

	lc := logger.NewMockClient()
	context := pkg.NewAppFuncContextForTest("test", lc)

	reading := initPosReadingBasketOpen()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(eventsProcessor.rttlogData), 1)
	assert.Equal(t, eventsProcessor.rttlogData[len(eventsProcessor.rttlogData)-1].EventTime, int64(1559679584))

	reading = initPosReadingPosItemSteak()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(eventsProcessor.rttlogData), 2)
	assert.Equal(t, eventsProcessor.rttlogData[len(eventsProcessor.rttlogData)-1].EventTime, int64(1559679673))

	reading = initPosReadingRemoveItem()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(eventsProcessor.rttlogData), 1)
	assert.Equal(t, eventsProcessor.rttlogData[len(eventsProcessor.rttlogData)-1].EventTime, int64(1559679584))

	reading = initPosReadingPaymentStart()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(eventsProcessor.rttlogData), 2)
	assert.Equal(t, eventsProcessor.rttlogData[len(eventsProcessor.rttlogData)-1].EventTime, int64(1559679588))

	reading = initPosReadingBasketClose()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(eventsProcessor.rttlogData), 1)
	assert.Equal(t, eventsProcessor.rttlogData[len(eventsProcessor.rttlogData)-1].EventTime, int64(1559679789))

}
