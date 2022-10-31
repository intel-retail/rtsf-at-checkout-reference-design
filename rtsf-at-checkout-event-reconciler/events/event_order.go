// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import "github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

const (
	posItemEvent        = "scanned-item"
	basketOpenEvent     = "basket-open"
	basketCloseEvent    = "basket-close"
	paymentStartEvent   = "payment-start"
	paymentSuccessEvent = "payment-success"
	removeItemEvent     = "remove-item"
	scaleItemEvent      = "weight"
	cvRoiEvent          = "cv-roi-event"
	rfidRoiEvent        = "rfid-roi-event"
)

var EventOccurred map[string]bool

func ResetEventsOccurrence() {
	events := []string{posItemEvent, basketOpenEvent, basketCloseEvent, paymentStartEvent, paymentSuccessEvent, removeItemEvent}
	EventOccurred = make(map[string]bool)
	for _, evt := range events {
		EventOccurred[evt] = false
	}
}

func checkEventOrderValid(event string, edgexcontext interfaces.AppFunctionContext) bool {
	eventValid := true
	switch event {
	case basketOpenEvent:
		if EventOccurred[basketOpenEvent] {
			eventValid = false
		} else {
			EventOccurred[basketOpenEvent] = true
			EventOccurred[basketCloseEvent] = false

			// add to clear the UI for the demo
			// sendWebsocketMessage([]byte("{\"positems\": [], \"cvsuspectitems\": [], \"rfidsuspectitems\": [], \"scalesuspectitems\": [] }"), edgexcontext)
		}
	case scaleItemEvent:
		if !EventOccurred[basketOpenEvent] || EventOccurred[basketCloseEvent] || EventOccurred[paymentStartEvent] || EventOccurred[paymentSuccessEvent] {
			eventValid = false
		} else {
			EventOccurred[scaleItemEvent] = true
		}
	case posItemEvent:
		if !EventOccurred[basketOpenEvent] || EventOccurred[basketCloseEvent] || EventOccurred[paymentStartEvent] || EventOccurred[paymentSuccessEvent] {
			eventValid = false
		} else {
			EventOccurred[posItemEvent] = true
		}
	case removeItemEvent:
		if !EventOccurred[basketOpenEvent] || EventOccurred[basketCloseEvent] || !EventOccurred[posItemEvent] || EventOccurred[paymentStartEvent] || EventOccurred[paymentSuccessEvent] {
			eventValid = false
		} else {
			EventOccurred[removeItemEvent] = true
		}
	case paymentStartEvent:
		if !EventOccurred[basketOpenEvent] || !EventOccurred[posItemEvent] || EventOccurred[paymentStartEvent] || EventOccurred[paymentSuccessEvent] || EventOccurred[basketCloseEvent] {
			eventValid = false
		} else {
			EventOccurred[paymentStartEvent] = true
		}
	case paymentSuccessEvent:
		if !EventOccurred[basketOpenEvent] || !EventOccurred[posItemEvent] || !EventOccurred[paymentStartEvent] || EventOccurred[paymentSuccessEvent] || EventOccurred[basketCloseEvent] {
			eventValid = false
		} else {
			EventOccurred[paymentSuccessEvent] = true
			EventOccurred[paymentStartEvent] = false
		}
	case basketCloseEvent:
		if !EventOccurred[basketOpenEvent] || EventOccurred[paymentStartEvent] {
			eventValid = false
		} else {
			ResetEventsOccurrence()
			EventOccurred[basketCloseEvent] = true

			// add to clear the UI for the demo
			// sendWebsocketMessage([]byte("{\"positems\": [], \"cvsuspectitems\": [], \"rfidsuspectitems\": [], \"scalesuspectitems\": [] }"), edgexcontext)
		}
	case cvRoiEvent:
		break
	case rfidRoiEvent:
		break
	default:
		eventValid = false
	}

	return eventValid
}
