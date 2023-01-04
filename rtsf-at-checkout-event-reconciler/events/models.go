// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"event-reconciler/config"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type EventsProcessor struct {
	afterPaymentSuccess     bool
	conn                    *websocket.Conn
	currentCVData           []CVEventEntry
	currentRFIDData         []RFIDEventEntry
	currentStateMessage     []byte
	cvTimeAlignment         time.Duration
	eventOccurred           map[string]bool
	firstBasketOpenComplete bool
	mu                      *sync.Mutex
	nextCVData              []CVEventEntry
	nextRFIDData            []RFIDEventEntry
	processConfig           *config.ReconcilerConfig
	rttlogData              []RTTLogEventEntry
	scaleData               []ScaleEventEntry
	suspectScaleItems       map[int64]*ScaleEventEntry
	upgrader                websocket.Upgrader
}

type RTTLogEventEntry struct {
	ProductId            string  `json:"product_id"`
	ProductIdType        string  `json:"product_id_type"`
	ProductName          string  `json:"product_name"`
	LaneId               string  `json:"lane_id"`
	BasketId             string  `json:"basket_id"`
	Quantity             float64 `json:"quantity"`
	QuantityUnit         string  `json:"quantity_unit"`
	UnitPrice            float64 `json:"unit_price"`
	EventTime            int64   `json:"event_time"`
	ScaleConfirmed       bool    `json:"scale_confirmed"`
	RFIDConfirmed        bool    `json:"rfid_confirmed"`
	CVConfirmed          bool    `json:"cv_confirmed"`
	CustomerId           string  `json:"customer_id"`
	EmployeeId           string  `json:"employee_id"`
	EventType            string  `json:"event_type"`
	CurrentWeightRange   ProductDetails
	Collection           []RTTLogEventEntry
	AssociatedScaleItems []*ScaleEventEntry
	AssociatedCVItems    []*CVEventEntry
	AssociatedRFIDItems  []*RFIDEventEntry
	ProductDetails       ProductDetails
}

type ProductDetails struct {
	Name              string  `json:"name"`
	ExpectedMinWeight float64 `json:"min_weight"`
	ExpectedMaxWeight float64 `json:"max_weight"`
	RFIDEligible      bool    `json:"rfid_eligible"`
}

type ScaleEventEntry struct {
	Delta               float64 `json:"delta"`
	Total               float64 `json:"total"`
	MinTolerance        string  `json:"min_tolerance"`
	MaxTolerance        string  `json:"max_tolerance"`
	Units               string  `json:"units"`
	SettlingTime        float64 `json:"settling_time"`
	MaxWeight           float64 `json:"max_weight"`
	LaneId              string  `json:"lane_id"`
	ScaleId             string  `json:"scale_id"`
	EventTime           int64   `json:"event_time"`
	Status              string  `json:"status"`
	AssociatedRTTLEntry *RTTLogEventEntry
}

type CVEventEntry struct {
	LaneId              string `json:"lane_id"`
	ObjectName          string `json:"product_name"`
	ROIName             string `json:"roi_name"`
	ROIAction           string `json:"roi_action"`
	EventTime           int64  `json:"event_time"`
	ROIs                map[string]ROILocation
	AssociatedRTTLEntry *RTTLogEventEntry
}

type RFIDEventEntry struct {
	ProductName         string `json:"product_name"`
	LaneId              string `json:"lane_id"`
	EPC                 string `json:"epc"`
	UPC                 string `json:"upc"`
	ROIName             string `json:"roi_name"`
	ROIAction           string `json:"roi_action"`
	EventTime           int64  `json:"event_time"`
	ROIs                map[string]ROILocation
	AssociatedRTTLEntry *RTTLogEventEntry
}

type ROILocation struct {
	AtLocation     bool
	LastAtLocation int64
}

type SuspectLists struct {
	CVSuspect    []CVEventEntry             `json:"cv_suspect_list"`
	RFIDSuspect  []RFIDEventEntry           `json:"rfid_suspect_list"`
	ScaleSuspect map[int64]*ScaleEventEntry `json:"scale_suspect_list"`
}

func NewEventsProcessor(cvTimeAlignment time.Duration, config *config.ReconcilerConfig) *EventsProcessor {
	processor := &EventsProcessor{
		afterPaymentSuccess:     false,
		currentCVData:           []CVEventEntry{},
		currentRFIDData:         []RFIDEventEntry{},
		cvTimeAlignment:         cvTimeAlignment,
		firstBasketOpenComplete: false,
		mu:                      &sync.Mutex{},
		nextCVData:              []CVEventEntry{},
		nextRFIDData:            []RFIDEventEntry{},
		processConfig:           config,
		suspectScaleItems:       make(map[int64]*ScaleEventEntry),
		upgrader:                websocket.Upgrader{},
	}

	return processor
}

func (eventsProcessing *EventsProcessor) GetScaleToScaleTolerance() float64 {
	return eventsProcessing.processConfig.ScaleToScaleTolerance
}
func (eventsProcessing *EventsProcessor) GetCurrentStateMessage() []byte {
	return eventsProcessing.currentStateMessage
}

//these custom Marshal functions are needed as json.Marshal() results in a stack overflow error from an infinite loop, due to cross-referencing of Associated RTTL/Scale Entries
func (rttl RTTLogEventEntry) toJSONString() string {

	rttlStr :=
		`{
			"product_id": "` + rttl.ProductId + `",
			"product_name": "` + rttl.ProductName + `",
			"quantity":  ` + fmt.Sprintf("%f", rttl.Quantity) + `,
			"quantity_unit": "` + rttl.QuantityUnit + `",
			"unit_price": ` + fmt.Sprintf("%f", rttl.UnitPrice) + `,
			"customer_id": "` + rttl.CustomerId + `",
			"employee_id": "` + rttl.EmployeeId + `",
			"event_time": ` + fmt.Sprintf("%d", rttl.EventTime) + `,
			"rfid_eligible": ` + strconv.FormatBool(rttl.ProductDetails.RFIDEligible) + `,
			"rfid_reconciled": ` + strconv.FormatBool(rttl.RFIDConfirmed) + `,
			"cv_reconciled": ` + strconv.FormatBool(rttl.CVConfirmed) + `,
			"scale_reconciled": ` + strconv.FormatBool(rttl.ScaleConfirmed) + `
		  }`

	return rttlStr
}

//same as comment above
func (scaleItem ScaleEventEntry) toJSONString() string {
	scaleStr :=
		`{
			"scale_id": "` + scaleItem.ScaleId + `",
			"total": ` + fmt.Sprintf("%f", scaleItem.Total) + `,
			"delta": ` + fmt.Sprintf("%f", scaleItem.Delta) + `,
			"event_time": "` + fmt.Sprintf("%d", scaleItem.EventTime) + `",
			"units": "` + scaleItem.Units + `"
		}`

	return scaleStr
}

//same as comment above
func (cvItem CVEventEntry) toJSONString() string {
	cvStr := `{
		"product_name": "` + cvItem.ObjectName + `",
		"event_time": "` + fmt.Sprintf("%d", cvItem.EventTime) +
		`"}`

	return cvStr
}

//same as comment above
func (rfidItem RFIDEventEntry) toJSONString() string {
	rfidStr := `{
		"product_name": "` + rfidItem.ProductName + `",
		"roi_name": "` + rfidItem.ROIName + `",
		"event_time": "` + fmt.Sprintf("%d", rfidItem.EventTime) +
		`"}`

	return rfidStr
}
