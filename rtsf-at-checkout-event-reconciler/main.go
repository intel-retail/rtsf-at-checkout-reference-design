// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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
	code := app.CreateAndRunAppService(serviceKey, pkg.NewAppService)
	os.Exit(code)
}

func (app *EventReconcilerAppService) CreateAndRunAppService(serviceKey string, newServiceFactory func(string) (interfaces.ApplicationService, bool)) int {
	var ok bool
	app.service, ok = pkg.NewAppServiceWithTargetType(serviceKey, []byte{})
	if !ok {
		return 1
	}

	app.lc = app.service.LoggingClient()

	events.ResetEventsOccurrence()
	events.InitWebSocketConnection(app.service, app.lc)

	appSettings := app.service.ApplicationSettings()
	if appSettings == nil {
		app.lc.Error("No application settings found")
		return 1
	}

	// retrieve the required configurations
	app.serviceConfig = &config.ServiceConfig{}
	if err := app.service.LoadCustomConfig(app.serviceConfig, "Reconciler"); err != nil {
		app.lc.Errorf("failed load custom Reconciler configuration: %s", err.Error())
		return 1
	}

	if err := app.serviceConfig.Reconciler.Validate(); err != nil {
		app.lc.Errorf("failed to validate Reconciler configuration: %v", err)
		return 1
	}

	events.ScaleToScaleTolerance = app.serviceConfig.Reconciler.ScaleToScaleTolerance

	tempDuration, err := time.ParseDuration(app.serviceConfig.Reconciler.CvTimeAlignment)
	if err != nil {
		app.lc.Errorf("failed to parse CvTimeAlignment duration: %v", err)
		return 1
	}
	events.CvTimeAlignment = tempDuration

	deviceNamesList := strings.TrimSpace(app.serviceConfig.Reconciler.DeviceNames)

	deviceNames := strings.Split(deviceNamesList, ",")
	app.lc.Infof("Running the application functions for %v devices...", deviceNames)

	app.service.AddRoute("/current-state", func(writer http.ResponseWriter, req *http.Request) {
		app.lc.Info("Sending current state message to UI")
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET")
		writer.Write(events.CurrentStateMessage)
		writer.WriteHeader(200)
	}, "GET")

	eventsProcessor := events.EventsProcessor{}
	eventsProcessor.ProcessConfig = &app.serviceConfig.Reconciler
	app.service.SetFunctionsPipeline(
		transforms.NewFilterFor(deviceNames).FilterByDeviceName,
		eventsProcessor.ProcessCheckoutEvents,
	)

	events.Mu = &sync.Mutex{}
	app.service.MakeItRun()

	return 0
}
