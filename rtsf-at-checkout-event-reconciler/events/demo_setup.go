// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/gorilla/websocket"
)

func (eventsProcessing *EventsProcessor) formatWebsocketMessage(eventName string) []byte {
	var sb strings.Builder
	sb.WriteString(`{
		"positems": [`)

	numPosItems := 0

	for _, rttlEntry := range eventsProcessing.rttlogData {
		if rttlEntry.Quantity > floatingPointTolerance {
			if numPosItems > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(rttlEntry.toJSONString())
			numPosItems++
		}
	}

	sb.WriteString(`]`)
	if len(eventsProcessing.scaleData) > 0 {
		lastScaleItem := eventsProcessing.scaleData[len(eventsProcessing.scaleData)-1]
		sb.WriteString(`,
		"scaleitem":`)
		sb.WriteString(lastScaleItem.toJSONString())
	}

	sb.WriteString(`,
		"scalesuspectitems": [`)
	idx := 0
	for _, suspectItem := range eventsProcessing.suspectScaleItems {
		if suspectItem.Delta > 0 {
			if idx > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(suspectItem.toJSONString())
			idx++
		}
	}
	sb.WriteString(`]`)

	// add CV
	sb.WriteString(`,
		"cvsuspectitems": [`)

	suspectCVItems := eventsProcessing.getSuspectCVItems()
	suspectCVLastIndex := len(suspectCVItems) - 1
	for suspectIndex, suspectItem := range suspectCVItems {
		sb.WriteString(suspectItem.toJSONString())
		if suspectIndex != suspectCVLastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(`]`)

	// add RFID
	sb.WriteString(`,
	"rfidsuspectitems": [`)

	suspectRFIDItems := eventsProcessing.getSuspectRFIDItems()
	suspectRFIDLastIndex := len(suspectRFIDItems) - 1
	for suspectIndex, suspectItem := range suspectRFIDItems {
		sb.WriteString(suspectItem.toJSONString())
		if suspectIndex != suspectRFIDLastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(`]`)

	// add Stats

	cvCount := len(eventsProcessing.currentCVData) + len(eventsProcessing.nextCVData)
	rfidCount := len(eventsProcessing.currentRFIDData) + len(eventsProcessing.nextRFIDData)
	scaleCount := len(eventsProcessing.scaleData)

	sb.WriteString(`,"stats": {
		"cv_count": "` + fmt.Sprintf("%v", cvCount) + `",
		"rfid_count": "` + fmt.Sprintf("%v", rfidCount) + `",
		"scale_count": "` + fmt.Sprintf("%v", scaleCount) +
		`"}`)

	sb.WriteString("\n}")

	// set the global suspect list
	eventsProcessing.currentStateMessage = []byte(sb.String())
	return eventsProcessing.currentStateMessage
}

// InitWebSocketConnection initializes the websocket
func (eventsProcessing *EventsProcessor) InitWebSocketConnection(service interfaces.ApplicationService, lc logger.LoggingClient) {
	wsAddr := eventsProcessing.processConfig.WebSocketPort

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		eventsProcessing.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		eventsProcessing.conn, err = eventsProcessing.upgrader.Upgrade(w, r, nil)
		if err != nil {
			lc.Errorf("upgrade: %s", err)
			return
		}
	})

	go func() {
		lc.Infof("websocket listening to: %v", wsAddr)
		if err := http.ListenAndServe(wsAddr, nil); err != nil {
			lc.Error(err.Error())
		}
	}()
}

func (eventsProcessing *EventsProcessor) sendWebsocketMessage(message []byte, edgexcontext interfaces.AppFunctionContext) {
	lc := edgexcontext.LoggingClient()
	if eventsProcessing.conn == nil {
		lc.Trace("websocket not connected")
		return
	}

	eventsProcessing.mu.Lock()
	defer eventsProcessing.mu.Unlock()
	lc.Tracef("websocket message: %v", string(message))
	err := eventsProcessing.conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		lc.Infof("write: %s", err)
		return
	}
}
