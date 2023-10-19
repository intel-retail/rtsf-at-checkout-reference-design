// Copyright © 2023 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

const (
	posName           = "POS"
	scaleName         = "Scale"
	cvName            = "CV-ROI"
	rfidName          = "RFID-ROI"
	configFilename    = "config.json"
	eventTimeLayout   = "2006-01-02 15:04:05 MST"
	jsonFileExtension = ".json"
)

var configuration Configuration

// Configuration is the struct for all the peripheral settings/configuration
type Configuration struct {
	PosEndpoint     string `json:"pos_endpoint"`
	ScaleEndpoint   string `json:"scale_endpoint"`
	CvRoiEndpoint   string `json:"cv_roi_endpoint"`
	RfidRoiEndpoint string `json:"rfid_roi_endpoint"`

	httpEndpoints map[string]url.URL
}

// CheckoutEvents is the struct for the collection of CheckoutEvents
type CheckoutEvents struct {
	Events []CheckoutEvent `json:"checkout_events"`
}

// CheckoutEvent is the struct for a checkout event at point of sale system
type CheckoutEvent struct {
	Device    string      `json:"device"`
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	WaitTime  string      `json:"wait_time"`
	EventTime *time.Time
}

type commandlineFlags struct {
	eventJSONFileFlag *string
}

func init() {
	content, err := os.ReadFile(configFilename)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	err = json.Unmarshal(content, &configuration)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	configuration.httpEndpoints = make(map[string]url.URL)

	endpoint, parseErr := url.Parse(configuration.PosEndpoint)
	if parseErr != nil {
		fmt.Println(parseErr)
		os.Exit(-1)
	}
	configuration.httpEndpoints[posName] = *endpoint

	endpoint, parseErr = url.Parse(configuration.ScaleEndpoint)
	if parseErr != nil {
		fmt.Println(parseErr)
		os.Exit(-1)
	}
	configuration.httpEndpoints[scaleName] = *endpoint

	endpoint, parseErr = url.Parse(configuration.CvRoiEndpoint)
	if parseErr != nil {
		fmt.Println(parseErr)
		os.Exit(-1)
	}
	configuration.httpEndpoints[cvName] = *endpoint

	endpoint, parseErr = url.Parse(configuration.RfidRoiEndpoint)
	if parseErr != nil {
		fmt.Println(parseErr)
		os.Exit(-1)
	}
	configuration.httpEndpoints[rfidName] = *endpoint
}

func (chkoutEvt *CheckoutEvent) wait() time.Duration {
	duration, err := time.ParseDuration(chkoutEvt.WaitTime)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	time.Sleep(duration)
	return duration
}

func (chkoutEvt *CheckoutEvent) send(eventTime time.Time) (waitTime time.Duration, err error) {
	fmt.Printf("Checkout event: %v\n", chkoutEvt)
	fmt.Printf("Event time: [%v]\n", eventTime.Format(eventTimeLayout))

	chkoutEvt.EventTime = &eventTime

	var chkoutEvtData map[string]interface{}
	payload, err := json.Marshal(chkoutEvt.Data)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	err = json.Unmarshal(payload, &chkoutEvtData)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	chkoutEvtData["event_time"] = eventTime.UnixNano() / int64(time.Millisecond)

	payload, err = json.Marshal(chkoutEvtData)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	go func() {
		resp, postErr := http.Post(chkoutEvt.buildHTTPEndpoint(), "application/json", bytes.NewBuffer(payload))

		if postErr != nil {
			fmt.Printf("Warning - %s data not sent: %v\n", chkoutEvt.Device, postErr)
			// ignore the error for now to keep going
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Post returns status of %d\n", resp.StatusCode)
			os.Exit(-1)
		}

		fmt.Printf("Event %s - %s sent\n", chkoutEvt.Device, chkoutEvt.Event)
	}()

	waitTime = chkoutEvt.wait()
	return waitTime, nil
}

func (chkoutEvt *CheckoutEvent) buildHTTPEndpoint() string {
	url := configuration.httpEndpoints[chkoutEvt.Device]
	url.Path = path.Join(url.Path, chkoutEvt.Event)

	fmt.Println(url.String())

	return url.String()
}

func loadCheckoutEvents(filePath string) (CheckoutEvents, error) {
	var events CheckoutEvents
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return events, err
	}

	fmt.Printf("Loading Events from [%s]\n", filePath)

	err = json.Unmarshal(contents, &events)
	if err != nil {
		return events, err
	}
	return events, nil
}

func processCommandLineFlags() (cmdlineFlags commandlineFlags) {
	cmdlineFlags.eventJSONFileFlag = flag.String("f", "tests/all_events.json", "Specify the JSON script file path for events; it will use the default value if omitted.")
	flag.Bool("h", false, "Print the usage of flags")
	flag.Bool("help", false, "Print the usage of flags")
	flag.Parse()

	usagePrint := false
	if len(os.Args) > 1 {
		flags := os.Args[1]
		switch strings.ToLower(flags) {
		case "-f":
			if !strings.HasSuffix(strings.ToLower(*cmdlineFlags.eventJSONFileFlag), jsonFileExtension) {
				fmt.Printf("Events script file should end with %s as the file extension\n", jsonFileExtension)
				os.Exit(-1)
			}
		case "-h":
			usagePrint = true
		case "--help":
			usagePrint = true
		default:
			usagePrint = true
		}

		if usagePrint {
			flag.PrintDefaults()
			os.Exit(0)
		}
	}
	return cmdlineFlags
}

func main() {
	cmdlineFlags := processCommandLineFlags()
	checkoutEvents, err := loadCheckoutEvents(*cmdlineFlags.eventJSONFileFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	// set the start time clock and use that as base to calculate event time
	eventTime := time.Now()
	for _, checkoutEvent := range checkoutEvents.Events {
		waitTime, err := checkoutEvent.send(eventTime)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		eventTime = eventTime.Add(waitTime)
	}

}
