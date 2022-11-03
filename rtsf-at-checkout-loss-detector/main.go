// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"

	"loss-detector/config"
	"loss-detector/functions"
)

const (
	serviceKey = "app-loss-detector"
)

func main() {
	var ok bool
	service, ok := pkg.NewAppServiceWithTargetType(serviceKey, &[]byte{})
	if !ok {
		os.Exit(-1)
	}

	lc := service.LoggingClient()

	// retrieve the required configurations
	serviceConfig := &config.ServiceConfig{}
	if err := service.LoadCustomConfig(serviceConfig, "LossDetector"); err != nil {
		lc.Errorf("failed load custom ControllerBoardStatus configuration: %s", err.Error())
		os.Exit(-1)
	}

	if err := serviceConfig.LossDetector.Validate(); err != nil {
		lc.Errorf("failed to validate ControllerBoardStatus configuration: %v", err)
		os.Exit(-1)
	}

	subscriptionClient := service.SubscriptionClient()
	if subscriptionClient == nil {
		lc.Errorf("error notification service missing from client's configuration")
		os.Exit(-1)
	}

	notificationClient := service.NotificationClient()
	if notificationClient == nil {
		lc.Error("error notification service missing from client's configuration")
		os.Exit(-1)
	}

	err := service.SetFunctionsPipeline(functions.NotifySuspectList)
	if err != nil {
		lc.Errorf("failed to set function pipline: %v", err)
		os.Exit(-1)
	}

	if err := functions.SubscribeToNotificationService(serviceConfig.LossDetector, subscriptionClient, lc); err != nil {
		lc.Info(fmt.Sprintf("Error subscribing to edgex notification service %s", err.Error()))
		os.Exit(-1)
	}

	err = service.MakeItRun()
	if err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(-1)
	}

	os.Exit(0)
}
