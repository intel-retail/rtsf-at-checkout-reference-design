// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"event-reconciler/config"
	"event-reconciler/events"
)

const (
	serviceKey = "app-event-reconciler"
)

type EventReconcilerAppService struct {
	service       interfaces.ApplicationService
	lc            logger.LoggingClient
	serviceConfig *config.ServiceConfig
}

func main() {
	app := EventReconcilerAppService{}
	code := app.CreateAndRunAppService(serviceKey)
	os.Exit(code)
}

func (app *EventReconcilerAppService) CreateAndRunAppService(serviceKey string) int {
	var ok bool
	app.service, ok = pkg.NewAppService(serviceKey)
	if !ok {
		return 1
	}

	app.lc = app.service.LoggingClient()

	// retrieve the required configurations
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "Reconciler"); err != nil {
		app.lc.Errorf("failed load custom Reconciler configuration: %s", err.Error())
		return 1
	}

	cvTimeAlignment, err := app.serviceConfig.Reconciler.Validate()
	if err != nil {
		app.lc.Errorf("failed to validate Reconciler configuration: %v", err)
		return 1
	}

	eventsProcessor := events.NewEventsProcessor(cvTimeAlignment, &app.serviceConfig.Reconciler)
	eventsProcessor.ResetEventsOccurrence()
	eventsProcessor.InitWebSocketConnection(app.service, app.lc)

	deviceNamesList := strings.TrimSpace(app.serviceConfig.Reconciler.DeviceNames)

	deviceNames := strings.Split(deviceNamesList, ",")
	app.lc.Infof("Running the application functions for %v devices...", deviceNames)

	app.service.AddRoute("/current-state", func(writer http.ResponseWriter, req *http.Request) {
		app.lc.Info("Sending current state message to UI")
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET")
		writer.Write(eventsProcessor.GetCurrentStateMessage())
		writer.WriteHeader(200)
	}, "GET")

	app.service.SetFunctionsPipeline(
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		eventsProcessor.ProcessCheckoutEvents,
	)

	app.service.MakeItRun()

	return 0
}
