// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"loss-detector/config"
	"loss-detector/functions"
)

const (
	serviceKey = "app-loss-detector"
)

type LossDetectorApp struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
}

func main() {
	app := LossDetectorApp{}
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

func (app *LossDetectorApp) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = pkg.NewAppServiceWithTargetType(serviceKey, &[]byte{})
	if !ok {
		return 1
	}

	app.lc = app.service.LoggingClient()

	// retrieve the required configurations
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "LossDetector"); err != nil {
		app.lc.Errorf("failed load custom ControllerBoardStatus configuration: %s", err.Error())
		return 1
	}

	if err := app.serviceConfig.LossDetector.Validate(); err != nil {
		app.lc.Errorf("failed to validate ControllerBoardStatus configuration: %v", err)
		return 1
	}

	subscriptionClient := app.service.SubscriptionClient()
	if subscriptionClient == nil {
		app.lc.Errorf("error notification service missing from client's configuration")
		return 1
	}

	notificationClient := app.service.NotificationClient()
	if notificationClient == nil {
		app.lc.Error("error notification service missing from client's configuration")
		return 1
	}

	err := app.service.SetFunctionsPipeline(functions.NotifySuspectList)
	if err != nil {
		app.lc.Errorf("failed to set function pipline: %v", err)
		return 1
	}

	if err := functions.SubscribeToNotificationService(app.serviceConfig.LossDetector, subscriptionClient, app.lc); err != nil {
		app.lc.Info(fmt.Sprintf("Error subscribing to edgex notification service %s", err.Error()))
		return 1
	}

	err = app.service.MakeItRun()
	if err != nil {
		app.lc.Errorf("MakeItRun returned error: %s", err.Error())
		return 1
	}

	return 0
}
