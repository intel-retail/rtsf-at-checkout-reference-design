// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"

	"loss-detector/functions"
)

const (
	serviceKey = "app-loss-detector"
)

func main() {
	var ok bool
	service, ok := pkg.NewAppServiceWithTargetType(serviceKey,&[]byte{})
	if !ok {
		os.Exit(-1)
	}

	lc := service.LoggingClient()

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

	service.SetFunctionsPipeline(functions.NotifySuspectList)

	if err := functions.SubscribeToNotificationService(service, subscriptionClient, lc); err != nil {
		lc.Info(fmt.Sprintf("Error subscribing to edgex notification service %s", err.Error()))
		os.Exit(-1)
	}

	err := service.MakeItRun()
	if err != nil {
		lc.Errorf("MakeItRun returned error: %s", err.Error())
		os.Exit(-1)
	}

	os.Exit(0)
}
