// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"

	"event-reconciler/events"
)

const (
	serviceKey = "EventReconciler"
)

func main() {

	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	events.ResetEventsOccurrence()
	events.InitWebSocketConnection(edgexSdk)

	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	defaultTolerance := 0.02
	if toleranceStr, ok := appSettings["ScaleToScaleTolerance"]; !ok {
		events.ScaleToScaleTolerance = defaultTolerance
		edgexSdk.LoggingClient.Error(fmt.Sprintf("ScaleToScaleTolerance setting not found defaulting to %v", defaultTolerance))
	} else {
		var err error
		events.ScaleToScaleTolerance, err = strconv.ParseFloat(toleranceStr, 64)
		if err != nil {
			edgexSdk.LoggingClient.Error(fmt.Sprintf("ScaleToScaleTolerance setting failed to parse defaulting to %v", defaultTolerance))
			events.ScaleToScaleTolerance = defaultTolerance
		}
	}

	var defaultCvTimeAlignment = 5 * time.Second
	cvTimeAlignmentStr, ok := appSettings["CvTimeAlignment"]
	if !ok {
		events.CvTimeAlignment = defaultCvTimeAlignment
		edgexSdk.LoggingClient.Error(fmt.Sprintf("CvTimeAlignment setting not found defaulting to %v", defaultTolerance))
	} else {
		var err error
		if events.CvTimeAlignment, err = time.ParseDuration(cvTimeAlignmentStr); err != nil {
			events.CvTimeAlignment = defaultCvTimeAlignment
			edgexSdk.LoggingClient.Error(fmt.Sprintf("failed to parse CvTimeAlignment: %v", err))
		}
	}

	deviceNamesList, ok := appSettings["DeviceNames"]
	if !ok {
		edgexSdk.LoggingClient.Error("DeviceNames application setting not found")
		os.Exit(-1)
	}

	deviceNamesList = strings.Replace(deviceNamesList, " ", "", -1)
	deviceNames := strings.Split(deviceNamesList, ",")
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Running the application functions for %v devices...", deviceNames))

	edgexSdk.AddRoute("/current-state", func(writer http.ResponseWriter, req *http.Request) {
		context := req.Context().Value(appsdk.SDKKey).(*appsdk.AppFunctionsSDK)
		context.LoggingClient.Info("Sending current state message to UI")
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET")
		writer.Write(events.CurrentStateMessage)
		writer.WriteHeader(200)
	}, "GET")

	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceNames).FilterByDeviceName,
		events.ProcessCheckoutEvents,
	)

	events.Mu = &sync.Mutex{}
	edgexSdk.MakeItRun()
}
