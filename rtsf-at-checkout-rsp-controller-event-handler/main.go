// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"

	"rsp-controller-event-handler/eventhandler"
)

const (
	serviceKey = "app-rsp-controller-event-handler"
)

func main() {

	var ok bool
	service, ok := pkg.NewAppService(serviceKey)
	if !ok {
		os.Exit(1)
	}

	lc := service.LoggingClient()

	deviceNames, err := service.GetAppSettingStrings("DeviceNames")
	if err != nil {
		lc.Errorf("DeviceNames application setting not found: %v",err)
		os.Exit(1)
	}

	lc.Infof("Running the application functions for %v devices...", deviceNames)

	valueDescriptor, err := service.GetAppSettingStrings("ValueDescriptorToFilter")
	if err != nil {
		lc.Error("ValueDescriptorToFilter application setting not found: %v", err)
		os.Exit(1)
	}

	err = service.SetFunctionsPipeline(
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		transforms.NewFilterFor(valueDescriptor).FilterByResourceName,
		eventhandler.ProcessRspControllerEvents,
	)

	if err != nil {
		lc.Errorf("faield to SetFunctionsPipeline: %v", err)
		os.Exit(1)
	}

	err = service.MakeItRun()
	if err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
