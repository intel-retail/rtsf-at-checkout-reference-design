// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package scale

import (
	"fmt"
	"testing"
)

func TestGetScaleReading(t *testing.T) {

	options := Config{
		PortName:        "/dev/tty.usbserial-test",
		BaudRate:        9600,
		DataBits:        7,
		StopBits:        1,
		MinimumReadSize: 1,
		ParityMode:      2,
		TimeOutMilli:    500,
	}

	scaleReading := make(chan Reading)
	readingErr := make(chan error)

	testDevice := InitializeMockDevice(&options)

	for index, expectedReading := range ReadingTbl {

		testDevice.TestCase = index

		GetScaleReading(testDevice, scaleReading, readingErr)
		select {
		case err := <-readingErr:
			t.Fatalf("Error reading from test device %v", err)

		case reading := <-scaleReading:
			fmt.Println(reading)
			if expectedReading.Value != reading.Value || expectedReading.Status != reading.Status || expectedReading.Unit != reading.Unit {
				t.Fatalf("Expected reading %v does not match real reading %v", expectedReading, reading)
			}
		}
	}
}
