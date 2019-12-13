// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"

	"loss-detector/functions"

)

const (
	serviceKey = "LossDetector"
)

func main() {

	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey, TargetType: &[]byte{}}
	if err := edgexSdk.Initialize(); err != nil {
		fmt.Printf("SDK initialization failed: %v\n", err)
		os.Exit(-1)
	}

	edgexSdk.SetFunctionsPipeline(
		functions.NotifySuspectList,
	)

	if err := functions.SubscribeToNotificationService(edgexSdk); err != nil {
		edgexSdk.LoggingClient.Info(fmt.Sprintf("Error subscribing to edgex notification service %s", err.Error()))
		os.Exit(-1)
	}

	edgexSdk.MakeItRun()
}
