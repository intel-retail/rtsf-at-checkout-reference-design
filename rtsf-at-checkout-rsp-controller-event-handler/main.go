// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"

	"rsp-controller-event-handler/eventhandler"
)

const (
	serviceKey = "RspControllerEventHandlerApp"
)

func main() {

	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		edgexSdk.LoggingClient.Error(fmt.Sprintf("SDK initialization failed: %v\n", err))
		os.Exit(-1)
	}

	appSettings := edgexSdk.ApplicationSettings()
	if appSettings == nil {
		edgexSdk.LoggingClient.Error("No application settings found")
		os.Exit(-1)
	}

	deviceNames, ok := appSettings["DeviceNames"]
	if !ok {
		edgexSdk.LoggingClient.Error("DeviceNames application setting not found")
		os.Exit(-1)
	}
	deviceNamesList := []string{deviceNames}
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Running the application functions for %v devices...", deviceNamesList))

	valueDescriptor, ok := appSettings["ValueDescriptorToFilter"]
	if !ok {
		edgexSdk.LoggingClient.Error("ValueDescriptorToFilter application setting not found")
		os.Exit(-1)
	}
	valueDescriptorList := []string{valueDescriptor}

	edgexSdk.SetFunctionsPipeline(
		transforms.NewFilter(deviceNamesList).FilterByDeviceName,
		transforms.NewFilter(valueDescriptorList).FilterByValueDescriptor,
		eventhandler.ProcessRspControllerEvents,
	)
	edgexSdk.MakeItRun()
}
