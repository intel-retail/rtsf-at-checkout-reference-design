// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package events

import (
	"testing"
)

func TestCheckEventOrderValid(t *testing.T) {
	tables := []struct {
		name           string
		eventsList     []string
		expectedResult bool
	}{
		{
			name:           "happy path",
			eventsList:     []string{basketOpenEvent, posItemEvent, scaleItemEvent, removeItemEvent, posItemEvent, paymentStartEvent, paymentSuccessEvent, basketCloseEvent},
			expectedResult: true,
		},
		{
			name:           "payment-start without scanning pos item",
			eventsList:     []string{basketOpenEvent, paymentStartEvent, paymentSuccessEvent, basketCloseEvent},
			expectedResult: false,
		},
		{
			name:           "basket-close with payment-start but without payment-complete",
			eventsList:     []string{basketOpenEvent, posItemEvent, paymentStartEvent, basketCloseEvent},
			expectedResult: false,
		},
		{
			name:           "payment-start without basket-open",
			eventsList:     []string{paymentStartEvent, basketCloseEvent},
			expectedResult: false,
		},
		{
			name:           "paymentsucess without basket-open",
			eventsList:     []string{paymentSuccessEvent},
			expectedResult: false,
		},
		{
			name:           "posItem without basket-open",
			eventsList:     []string{posItemEvent},
			expectedResult: false,
		},
		{
			name:           "basket-open after basket-open",
			eventsList:     []string{basketOpenEvent, basketOpenEvent},
			expectedResult: false,
		},
		{
			name:           "bad event name",
			eventsList:     []string{"hello"},
			expectedResult: false,
		},
		{
			name:           "scaleEvent without basket-open",
			eventsList:     []string{scaleItemEvent},
			expectedResult: false,
		},
		{
			name:           "removeItem after payment",
			eventsList:     []string{basketOpenEvent, posItemEvent, paymentStartEvent, removeItemEvent},
			expectedResult: false,
		},
	}

	for _, table := range tables {
		ResetEventsOccurrence()
		var eventValid bool
		for _, event := range table.eventsList {
			eventValid = checkEventOrderValid(event, nil)
			if !eventValid {
				break
			}
		}
		if eventValid != table.expectedResult {
			t.Errorf("Test failed. Name: %s\n. Expected: %t, Actual: %t\n", table.name, table.expectedResult, eventValid)
		}
	}

}
