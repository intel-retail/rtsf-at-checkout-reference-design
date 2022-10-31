// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"encoding/json"
	"event-reconciler/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
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

	m.Run()
}

func initCVReadingScannerENTER() models.ObjectReading {
	reading := models.ObjectReading{
		// Name:  cvRoiEvent,
		BaseReading: models.BaseReading{
			ResourceName: cvRoiEvent,
		},
		ObjectValue: `{"object_count":1,"product_name":"item 1","roi_action":"ENTERED","event_time":1559679684,"roi_name":"Scanner"}`,
	}

	return reading
}

func initCVReadingScannerEXIT() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: cvRoiEvent,
		},
		ObjectValue: `{"object_count":1,"product_name":"item 1","roi_action":"EXITED","event_time":1559679695,"roi_name":"Scanner"}`,
	}

	return reading
}

func initCVReadingBaggingENTER() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: cvRoiEvent,
		},
		ObjectValue: `{"object_count":1,"product_name":"item 1","roi_action":"ENTERED","event_time":1559679784,"roi_name":"Bagging"}`,
	}

	return reading
}

func initCVReadingNewItemStagingENTER() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: cvRoiEvent,
		},
		ObjectValue: `{"object_count":1,"product_name":"Item 2","roi_action":"ENTERED","event_time":1559679834,"roi_name":"Staging"}`,
	}

	return reading
}

func initCVReadingNewItemEntranceENTER() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: cvRoiEvent,
		},
		ObjectValue: `{"object_count":1,"product_name":"Item 3","roi_action":"ENTERED","event_time":1559679834,"roi_name":"Entrance"}`,
	}

	return reading
}

func initScaleReadingScaleItem() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: scaleItemEvent,
		},
		ObjectValue: `{"event_time":1559679665,"lane_id":"123","scale_id":"123","total":2,"units":"lbs"}`,
	}

	return reading
}

func initPosReadingBasketOpen() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: basketOpenEvent,
		},
		ObjectValue: `{"basket_id":"abc-012345-def","customer_id":"joe5","employee_id":"mary1","event_time":1559679584,"lane_id":"123"}`,
	}
	return reading
}

func initPosReadingBasketClose() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: basketOpenEvent,
		},
		ObjectValue: `{"basket_id":"abc-012345-def","customer_id":"joe5","employee_id":"mary1","event_time":1559679789,"lane_id":"123"}`,
	}
	return reading
}

func initPosReadingRemoveItem() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: removeItemEvent,
		},
		ObjectValue: `{"basket_id":"abc-012345-def","customer_id":"joe5","employee_id":"mary1","event_time":1559679672,"lane_id":"123","product_id":"00000000735797","product_id_type":"UPC","product_name":"Steak","quantity":3,"quantity_unit":"lbs","unit_price":8.99}`,
	}
	return reading
}

func initPosReadingPosItemSteak() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: posItemEvent,
		},
		ObjectValue: `{"basket_id":"abc-012345-def","customer_id":"joe5","employee_id":"mary1","event_time":1559679673,"lane_id":"123","product_id":"00000000735797","product_id_type":"UPC","product_name":"Steak","quantity":3,"quantity_unit":"lbs","unit_price":8.99}`,
	}
	return reading
}

func initPosReadingPosItemApples() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: posItemEvent,
		},
		ObjectValue: `{
			"basket_id": "abc-012345-def",
			"product_id": "00000000324588",
			"product_id_type": "UPC",
			"product_name": "Red Apples",
			"quantity": 1.0,
			"quantity_unit": "EA",
			"unit_price": 0.99,
			"customer_id": "joe5",
			"employee_id": "mary1"
		}`,
	}
	return reading
}

func initPosReadingPaymentStart() models.ObjectReading {
	reading := models.ObjectReading{
		//	Name:  paymentStartEvent,
		BaseReading: models.BaseReading{
			ResourceName: paymentStartEvent,
		},
		ObjectValue: `{"basket_id":"abc-012345-def","customer_id":"joe5","employee_id":"mary1","event_time":1559679588,"lane_id":"123"}`,
	}
	return reading
}

func initRFIDReadingApplesExitedBagging() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: rfidRoiEvent,
		},
		ObjectValue: `{
			"epc": "30140000001FB28000003039",
			"roi_name": "Bagging",
			"roi_action": "EXITED",
			"event_time": 1562972496874
		}`,
	}
	return reading
}

func initRFIDReadingApplesEnterBagging() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: rfidRoiEvent,
		},
		ObjectValue: `{
			"epc": "30140000001FB28000003039",
			"roi_name": "Bagging",
			"roi_action": "ENTERED",
			"event_time": 1562972496854
		}`,
	}

	return reading
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

func initRFIDReadingSteakEnterBagging() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: rfidRoiEvent,
		},
		ObjectValue: `{
			"epc": "301400000047DAC000003039",
			"roi_name": "Bagging",
			"roi_action": "ENTERED",
			"event_time": 1562972496854
		}`,
	}
	return reading
}

func initRFIDReadingSalsaEnterBagging() models.ObjectReading {
	reading := models.ObjectReading{
		BaseReading: models.BaseReading{
			ResourceName: rfidRoiEvent,
		},
		ObjectValue: `{
			"epc": "301400000051240000003039",
			"roi_name": "Bagging",
			"roi_action": "ENTERED",
			"event_time": 1562972496854
		}`,
	}
	return reading
}

func TestProcessDeviceRFIDReading(t *testing.T) {

	CurrentRFIDData = []RFIDEventEntry{}
	NextRFIDData = []RFIDEventEntry{}

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
	eventsProcessor := EventsProcessor{
		ProcessConfig: &config.ReconcilerConfig{
			ProductLookupEndpoint: tsURL.Hostname() + ":" + tsURL.Port(),
		},
	}
	lc := logger.MockLogger{}

	reading := initRFIDReadingApplesEnterBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)
	assert.Equal(t, len(CurrentRFIDData), 1)
	assert.Contains(t, CurrentRFIDData[0].ROIs, BaggingROI)
	assert.True(t, CurrentRFIDData[0].ROIs[BaggingROI].AtLocation)

	reading = initRFIDReadingApplesExitedBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)
	assert.Equal(t, len(CurrentRFIDData), 1)
	assert.False(t, CurrentRFIDData[0].ROIs[BaggingROI].AtLocation)

	reading = initRFIDReadingSteakEnterBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)
	assert.Equal(t, len(CurrentRFIDData), 2)
	assert.Equal(t, len(NextRFIDData), 0)
	assert.True(t, CurrentRFIDData[1].ROIs[BaggingROI].AtLocation)

	afterPaymentSuccess = true

	reading = initRFIDReadingSalsaEnterBagging()
	eventsProcessor.processDeviceRFIDReading(reading, lc)

	assert.Equal(t, len(CurrentRFIDData), 2)
	assert.Equal(t, len(NextRFIDData), 1)

	afterPaymentSuccess = false
}

func TestProcessDeviceCVReading(t *testing.T) {
	lc := logger.NewMockClient()
	eventsProcessor := EventsProcessor{}

	CurrentCVData = []CVEventEntry{}
	NextCVData = []CVEventEntry{}

	reading := initCVReadingScannerENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(CurrentCVData), 1)

	reading = initCVReadingScannerEXIT()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(CurrentCVData), 1)

	reading = initCVReadingBaggingENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(CurrentCVData), 1)

	reading = initCVReadingNewItemStagingENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)

	assert.Equal(t, len(CurrentCVData), 2)

	afterPaymentSuccess = true

	reading = initCVReadingNewItemEntranceENTER()
	eventsProcessor.processDeviceCVReading(reading, lc)
	assert.Equal(t, len(CurrentCVData), 2)
	assert.Equal(t, len(NextCVData), 1)

	afterPaymentSuccess = false
}

func TestProcessDeviceScaleReading(t *testing.T) {
	resetRTTLBasket()

	lc := logger.NewMockClient()
	eventsProcessor := EventsProcessor{}

	reading := initScaleReadingScaleItem()
	eventsProcessor.processDeviceScaleReading(reading, lc)
	assert.Equal(t, len(ScaleData), 1)
	assert.Equal(t, len(SuspectScaleItems), 1)
	assert.Equal(t, ScaleData[0].Total, 2.0)
	assert.Equal(t, ScaleData[0].Units, "lbs")
}

func TestProcessDevicePosReading(t *testing.T) {

	BasketOpen()
	resetRTTLBasket()

	lc := logger.NewMockClient()
	eventsProcessor := EventsProcessor{}
	context := pkg.NewAppFuncContextForTest("test", lc)

	reading := initPosReadingBasketOpen()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(RttlogData), 1)
	assert.Equal(t, RttlogData[len(RttlogData)-1].EventTime, int64(1559679584))

	reading = initPosReadingPosItemSteak()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(RttlogData), 2)
	assert.Equal(t, RttlogData[len(RttlogData)-1].EventTime, int64(1559679673))

	reading = initPosReadingRemoveItem()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(RttlogData), 1)
	assert.Equal(t, RttlogData[len(RttlogData)-1].EventTime, int64(1559679584))

	reading = initPosReadingPaymentStart()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(RttlogData), 2)
	assert.Equal(t, RttlogData[len(RttlogData)-1].EventTime, int64(1559679588))

	reading = initPosReadingBasketClose()
	eventsProcessor.processDevicePosReading(reading, context)
	assert.Equal(t, len(RttlogData), 1)
	assert.Equal(t, RttlogData[len(RttlogData)-1].EventTime, int64(1559679789))

}
