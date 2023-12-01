// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import "github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"

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

func (eventsProcessing *EventsProcessor) ResetEventsOccurrence() {
	events := []string{posItemEvent, basketOpenEvent, basketCloseEvent, paymentStartEvent, paymentSuccessEvent, removeItemEvent}
	eventsProcessing.eventOccurred = make(map[string]bool)
	for _, evt := range events {
		eventsProcessing.eventOccurred[evt] = false
	}
}

func (eventsProcessing *EventsProcessor) checkEventOrderValid(event string, edgexcontext interfaces.AppFunctionContext) bool {
	eventValid := true
	switch event {
	case basketOpenEvent:
		if eventsProcessing.eventOccurred[basketOpenEvent] {
			eventValid = false
		} else {
			eventsProcessing.eventOccurred[basketOpenEvent] = true
			eventsProcessing.eventOccurred[basketCloseEvent] = false

			// add to clear the UI for the demo
			// sendWebsocketMessage([]byte("{\"positems\": [], \"cvsuspectitems\": [], \"rfidsuspectitems\": [], \"scalesuspectitems\": [] }"), edgexcontext)
		}
	case scaleItemEvent:
		if !eventsProcessing.eventOccurred[basketOpenEvent] || eventsProcessing.eventOccurred[basketCloseEvent] || eventsProcessing.eventOccurred[paymentStartEvent] || eventsProcessing.eventOccurred[paymentSuccessEvent] {
			eventValid = false
		} else {
			eventsProcessing.eventOccurred[scaleItemEvent] = true
		}
	case posItemEvent:
		if !eventsProcessing.eventOccurred[basketOpenEvent] || eventsProcessing.eventOccurred[basketCloseEvent] || eventsProcessing.eventOccurred[paymentStartEvent] || eventsProcessing.eventOccurred[paymentSuccessEvent] {
			eventValid = false
		} else {
			eventsProcessing.eventOccurred[posItemEvent] = true
		}
	case removeItemEvent:
		if !eventsProcessing.eventOccurred[basketOpenEvent] || eventsProcessing.eventOccurred[basketCloseEvent] || !eventsProcessing.eventOccurred[posItemEvent] || eventsProcessing.eventOccurred[paymentStartEvent] || eventsProcessing.eventOccurred[paymentSuccessEvent] {
			eventValid = false
		} else {
			eventsProcessing.eventOccurred[removeItemEvent] = true
		}
	case paymentStartEvent:
		if !eventsProcessing.eventOccurred[basketOpenEvent] || !eventsProcessing.eventOccurred[posItemEvent] || eventsProcessing.eventOccurred[paymentStartEvent] || eventsProcessing.eventOccurred[paymentSuccessEvent] || eventsProcessing.eventOccurred[basketCloseEvent] {
			eventValid = false
		} else {
			eventsProcessing.eventOccurred[paymentStartEvent] = true
		}
	case paymentSuccessEvent:
		if !eventsProcessing.eventOccurred[basketOpenEvent] || !eventsProcessing.eventOccurred[posItemEvent] || !eventsProcessing.eventOccurred[paymentStartEvent] || eventsProcessing.eventOccurred[paymentSuccessEvent] || eventsProcessing.eventOccurred[basketCloseEvent] {
			eventValid = false
		} else {
			eventsProcessing.eventOccurred[paymentSuccessEvent] = true
			eventsProcessing.eventOccurred[paymentStartEvent] = false
		}
	case basketCloseEvent:
		if !eventsProcessing.eventOccurred[basketOpenEvent] || eventsProcessing.eventOccurred[paymentStartEvent] {
			eventValid = false
		} else {
			eventsProcessing.ResetEventsOccurrence()
			eventsProcessing.eventOccurred[basketCloseEvent] = true

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
