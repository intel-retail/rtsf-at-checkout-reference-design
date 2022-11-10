// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/gorilla/websocket"
)

var wsPort string
var upgrader = websocket.Upgrader{} // use default options
var conn *websocket.Conn
var Mu *sync.Mutex

var CurrentStateMessage []byte

func formatWebsocketMessage(eventName string) []byte {
	var sb strings.Builder
	sb.WriteString(`{
		"positems": [`)

	numPosItems := 0

	for _, rttlEntry := range RttlogData {
		if rttlEntry.Quantity > floatingPointTolerance {
			if numPosItems > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(rttlEntry.toJSONString())
			numPosItems++
		}
	}

	sb.WriteString(`]`)
	if len(ScaleData) > 0 {
		lastScaleItem := ScaleData[len(ScaleData)-1]
		sb.WriteString(`,
		"scaleitem":`)
		sb.WriteString(lastScaleItem.toJSONString())
	}

	sb.WriteString(`,
		"scalesuspectitems": [`)
	idx := 0
	for _, suspectItem := range SuspectScaleItems {
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

	suspectCVItems := getSuspectCVItems()
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

	suspectRFIDItems := getSuspectRFIDItems()
	suspectRFIDLastIndex := len(suspectRFIDItems) - 1
	for suspectIndex, suspectItem := range suspectRFIDItems {
		sb.WriteString(suspectItem.toJSONString())
		if suspectIndex != suspectRFIDLastIndex {
			sb.WriteString(",")
		}
	}
	sb.WriteString(`]`)

	// add Stats

	cvCount := len(CurrentCVData) + len(NextCVData)
	rfidCount := len(CurrentRFIDData) + len(NextRFIDData)
	scaleCount := len(ScaleData)

	sb.WriteString(`,"stats": {
		"cv_count": "` + fmt.Sprintf("%v", cvCount) + `",
		"rfid_count": "` + fmt.Sprintf("%v", rfidCount) + `",
		"scale_count": "` + fmt.Sprintf("%v", scaleCount) +
		`"}`)

	sb.WriteString("\n}")

	// set the global suspect list
	CurrentStateMessage = []byte(sb.String())
	return CurrentStateMessage
}

// InitWebSocketConnection initializes the websocket
func InitWebSocketConnection(service interfaces.ApplicationService, lc logger.LoggingClient) {

	var wsAddr string
	appSettings := service.ApplicationSettings()
	if wsPortConfig, ok := appSettings["WebSocketPort"]; !ok {
		defaultPort := "9083"
		lc.Errorf(fmt.Sprintf("WebSocketAddress setting not found defaulting to %v", defaultPort))
		wsAddr = fmt.Sprintf(":%s", defaultPort)
	} else {
		wsAddr = fmt.Sprintf(":%s", wsPortConfig)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			lc.Errorf(fmt.Sprintf("upgrade: %s\n", err))
			return
		}
	})

	go func() {
		lc.Infof(fmt.Sprintf("websocket listening to: %v \n", wsAddr))
		if err := http.ListenAndServe(wsAddr, nil); err != nil {
			lc.Error(err.Error())
		}
	}()
}

func sendWebsocketMessage(message []byte, edgexcontext interfaces.AppFunctionContext) {
	lc := edgexcontext.LoggingClient()
	if conn == nil {
		lc.Trace("websocket not connected")
		return
	}

	Mu.Lock()
	defer Mu.Unlock()
	lc.Trace(fmt.Sprintf("websocket message: %v\n", string(message)))
	err := conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		lc.Info(fmt.Sprintf("write: %s", err))
		return
	}
}
