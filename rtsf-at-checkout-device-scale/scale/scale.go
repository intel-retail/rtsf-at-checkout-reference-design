// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package scale

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// Config configures how to connect to serial scale
type Config struct {
	// The name of the port, e.g. "/dev/tty.usbserial-A8008HlV".
	PortName string
	// The baud rate for the port.
	BaudRate uint
	// The number of data bits per frame. Legal values are 5, 6, 7, and 8.
	DataBits uint
	// The number of stop bits per frame. Legal values are 1 and 2.
	StopBits uint
	// The type of parity bits to use for the connection. none = 0, odd = 1, even = 2
	ParityMode int
	// The minimum buffer read size 1, or 2
	MinimumReadSize uint
	// The time out in milliseconds for reading from the scale
	TimeOutMilli int64

	options *serial.OpenOptions
}

// Reading the weight and status from the scale
type Reading struct {
	Status string
	Value  string
	Unit   string
}

// SerialDevice a device used to connect to a weight scale
type SerialDevice interface {
	openSerialPort() error
	getReading(scaleReading chan Reading, readingErr chan error)
	sendBytes(bytes []byte) (int, error)
	readBytes() ([]byte, error)
	getSerialPort() io.ReadWriteCloser
	getConfig() *Config
}

// CasPD2 serial weight scale
type CasPD2 struct {
	serialPort io.ReadWriteCloser
	config     *Config
}

// NewScale creates a new instance of the SerialDevice
func NewScale(config Config) SerialDevice {

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

	return newCasPD2(&config)
}

// GetScaleReading gets the current reading and status from the scale (async call), config timeOutMilli specifies the maximum time before the request times out
func GetScaleReading(device SerialDevice, scaleReading chan Reading, readingErr chan error) {

	broadcastReading := make(chan Reading)
	broadcastErr := make(chan error)
	go func() {

		if err := device.openSerialPort(); err != nil {
			readingErr <- err
			return
		}

		defer device.getSerialPort().Close()

		weightRequest := []byte{0x57, 0x0D}
		timeOut := time.NewTimer(time.Duration(device.getConfig().TimeOutMilli) * time.Millisecond)

		go device.getReading(broadcastReading, readingErr)

		_, err := device.sendBytes(weightRequest)
		if err != nil {
			readingErr <- err
		}

		select {
		case err = <-broadcastErr:
			readingErr <- err
			return
		case reading := <-broadcastReading:
			scaleReading <- reading
			return
		case <-timeOut.C:
			timeOutMsg := "time out connecting to scale"
			readingErr <- errors.New(timeOutMsg)
		}
	}()
}

func newCasPD2(config *Config) *CasPD2 {
	return &CasPD2{config: config}
}

func (device *CasPD2) openSerialPort() error {
	serialPort, err := serial.Open(*device.config.options)
	if err != nil {
		return err
	}
	device.serialPort = serialPort
	return nil
}

func (device *CasPD2) getSerialPort() io.ReadWriteCloser {
	return device.serialPort
}

func (device *CasPD2) getConfig() *Config {
	return device.config
}

func (device *CasPD2) getReading(scaleReading chan Reading, readingErr chan error) {
	var weightBytes []byte
	var err error
	bufferBytes := []byte{}
	statusEnding := "0D03"
	weightEnding := "0D0A"
	weightValueStart := "0A3"
	periodByte := "2E"
	expectedPeriodIndex := 6

	statusTbl := map[string]string{
		"00": "OK",
		"10": "Motion",
		"20": "Scale at Zero",
		"01": "Under Capacity",
		"02": "Over Capacity",
	}

	for {
		weightBytes, err = device.readBytes()
		if err != nil {
			readingErr <- err
			return
		}

		bufferBytes = append(bufferBytes, weightBytes...)
		bufferStr := strings.ToUpper(hex.EncodeToString(bufferBytes))
		fmt.Println(bufferStr)

		startIndex := strings.Index(bufferStr, weightValueStart)

		fmt.Printf("start index: %d\n", startIndex)

		if startIndex != 0 || len(bufferStr) > 30 {
			return
		}

		if strings.Contains(bufferStr, statusEnding) {

			reading := Reading{}
			weightLen := 16
			statusLen := 4
			statusIndex := strings.Index(bufferStr, statusEnding)
			weightIndex := strings.Index(bufferStr, weightEnding)
			periodIndex := strings.Index(bufferStr, periodByte)

			if statusIndex >= 0 && len(bufferStr) > statusIndex {
				status := bufferStr[statusIndex-statusLen : statusIndex]
				s, _ := hex.DecodeString(status)
				statusDef, ok := statusTbl[fmt.Sprintf("%s", s)]
				if !ok {
					statusDef = "N/A"
				}

				reading.Status = statusDef
			}

			if weightIndex >= 0 && len(bufferStr) > weightIndex && periodIndex == expectedPeriodIndex {

				weightStr := bufferStr[weightIndex-weightLen : weightIndex]
				weight, _ := hex.DecodeString(weightStr)

				unitLen := 2
				unitIndex := len(weight) - unitLen
				reading.Value = string(weight[0:unitIndex])
				reading.Unit = string(weight[unitIndex:len(weight)])
			}
			scaleReading <- reading
			break
		}
	}
}

func (device *CasPD2) sendBytes(bytes []byte) (int, error) {

	n, err := device.serialPort.Write(bytes)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (device *CasPD2) readBytes() ([]byte, error) {

	bufSize := 16
	buf := make([]byte, bufSize)

	n, err := device.serialPort.Read(buf)

	if err != nil {
		return nil, err
	}

	readBytes := buf[:n]
	return readBytes, nil
}
