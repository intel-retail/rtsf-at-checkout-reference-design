// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package scale

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/jacobsa/go-serial/serial"
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

	testDevice := initializeMockDevice(&options)

	for index, expectedReading := range readingTbl {

		testDevice.testCase = index

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

var readingTbl = map[int]Reading{
	0: Reading{Status: "OK", Value: "02.494", Unit: "LB"},
	1: Reading{Status: "Scale at Zero", Value: "000.00", Unit: "LB"},
	2: Reading{Status: "OK", Value: "0.854", Unit: "OZ"},
	3: Reading{Status: "Over Capacity", Value: "000.00", Unit: "OZ"},
	4: Reading{Status: "Under Capacity", Value: "000.00", Unit: "LB"},
}

type mockDevice struct {
	serialPort io.ReadWriteCloser
	config     *Config
	testCase   int
}

func initializeMockDevice(config *Config) *mockDevice {

	options := &serial.OpenOptions{BaudRate: config.BaudRate,
		DataBits:        config.DataBits,
		MinimumReadSize: config.MinimumReadSize,
		ParityMode:      serial.ParityMode(config.ParityMode),
		PortName:        config.PortName,
		StopBits:        config.StopBits}

	config.options = options
	if config.TimeOutMilli <= 0 {
		config.TimeOutMilli = 500
	}

	return &mockDevice{config: config}
}

func (device *mockDevice) openSerialPort() error {
	device.serialPort = &os.File{}
	return nil
}

func (device *mockDevice) getSerialPort() io.ReadWriteCloser {
	return device.serialPort
}

func (device *mockDevice) getConfig() *Config {
	return device.config
}

func (device *mockDevice) getReading(scaleReading chan Reading, readingErr chan error) {

	reading := readingTbl[device.testCase]

	scaleReading <- reading
}

func (device *mockDevice) sendBytes(bytes []byte) (int, error) {
	// no op
	return 0, nil
}

func (device *mockDevice) readBytes() ([]byte, error) {
	// no op
	return nil, nil
}
