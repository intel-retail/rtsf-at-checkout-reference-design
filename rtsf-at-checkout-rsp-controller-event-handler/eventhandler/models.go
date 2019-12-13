// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package eventhandler

type RFIDEventEntry struct {
	LaneId         string `json:"lane_id"`
	EPC            string `json:"epc"`
	ROIName        string `json:"roi_name"`
	ROIAction      string `json:"roi_action"`
	AtEntrance     bool   `json:"at_entrance"`
	LastAtEntrance int64  `json:"last_at_entrance"`
	AtGoBack       bool   `json:"at_go_back"`
	LastAtGoBack   int64  `json:"last_at_go_back"`
	AtBagging      bool   `json:"at_bagging"`
	LastAtBagging  int64  `json:"last_at_bagging"`
	InCart         bool   `json:"in_cart"`
	LastInCart     int64  `json:"last_in_cart"`
	EventTime      int64  `json:"event_time"`
	//  AssociatedRTTLEntry *RTTLogEventEntry
	//DetectedAt	AntennaData
}

type RspControllerEvent struct {
	LaneId  string `json:"lane_id"`
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  RspControllerEventParams
}

type RspControllerEventParams struct {
	SentOn    int64  `json:"sent_on"`
	GatewayId string `json:"gateway_id"`
	Data      []RspControllerEventParamsData
}

type RspControllerEventParamsData struct {
	FacilityId      string `json:"facility_id"`
	EPCCode         string `json:"epc_code"`
	EPCEncodeFormat string `json:"epc_encode_format"`
	EventType       string `json:"event_type"`
	TimeStamp       int64  `json:"timestamp"`
	Location        string `json:"location"`
}
