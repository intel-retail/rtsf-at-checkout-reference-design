package scale

import (
	"io"
	"os"

	"github.com/jacobsa/go-serial/serial"
)

var ReadingTbl = map[int]Reading{
	0: Reading{Status: "OK", Value: "02.494", Unit: "LB"},
	1: Reading{Status: "Scale at Zero", Value: "000.00", Unit: "LB"},
	2: Reading{Status: "OK", Value: "0.854", Unit: "OZ"},
	3: Reading{Status: "Over Capacity", Value: "000.00", Unit: "OZ"},
	4: Reading{Status: "Under Capacity", Value: "000.00", Unit: "LB"},
	5: Reading{Status: "OK", Value: "a###", Unit: "LB"},
}

type MockDevice struct {
	SerialPort io.ReadWriteCloser
	Config     *Config
	TestCase   int
}

func InitializeMockDevice(config *Config) *MockDevice {

	options := &serial.OpenOptions{BaudRate: config.BaudRate,
		DataBits:        config.DataBits,
		MinimumReadSize: config.MinimumReadSize,
		ParityMode:      serial.ParityMode(config.ParityMode),
		PortName:        config.PortName,
		StopBits:        config.StopBits}

	config.Options = options
	if config.TimeOutMilli <= 0 {
		config.TimeOutMilli = 500
	}

	return &MockDevice{Config: config}
}

func (device *MockDevice) openSerialPort() error {
	device.SerialPort = &os.File{}
	return nil
}

func (device *MockDevice) getSerialPort() io.ReadWriteCloser {
	return device.SerialPort
}

func (device *MockDevice) getConfig() *Config {
	return device.Config
}

func (device *MockDevice) getReading(scaleReading chan Reading, readingErr chan error) {

	reading := ReadingTbl[device.TestCase]

	scaleReading <- reading
}

func (device *MockDevice) sendBytes(bytes []byte) (int, error) {
	// no op
	return 0, nil
}

func (device *MockDevice) readBytes() ([]byte, error) {
	// no op
	return nil, nil
}
