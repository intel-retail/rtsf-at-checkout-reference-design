// Copyright © 2019 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/startup"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/notifications"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

const (
	notificationsURL = "NotificationsURL"
	useRegistry      = false
	interval         = -1
)

// NotifySuspectList sends a notification to edgex-go
func NotifySuspectList(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {

	edgexcontext.LoggingClient.Info("Notifying suspect list")

	if len(params) < 1 {
		return false, nil
	}

	contentBytes, _ := params[0].([]byte)
	contentFormated, err := json.MarshalIndent(json.RawMessage(contentBytes), "", "  ")

	if err != nil {
		edgexcontext.LoggingClient.Error("Failed to post notification, %s", err.Error())
		return false, nil
	}

	content := "Suspicious Items:\n" + string(contentFormated)

	url, ok := edgexcontext.Configuration.ApplicationSettings[notificationsURL]
	if !ok {
		edgexcontext.LoggingClient.Error(notificationsURL + " setting not found")
		return false, nil
	}

	endpointParams := types.EndpointParams{
		ServiceKey:  clients.SupportNotificationsServiceKey,
		Path:        clients.ApiNotificationRoute,
		UseRegistry: useRegistry,
		Url:         url + clients.ApiNotificationRoute,
		Interval:    interval,
	}

	notification := notifications.Notification{
		Slug:        "suspect-items-" + time.Now().String(),
		Sender:      "Loss Detector",
		Category:    notifications.SECURITY,
		Severity:    notifications.CRITICAL,
		Content:     content,
		Description: "Suspect lists for CV, RFID, and scale",
		Labels: []string{
			string(notifications.SECURITY),
		},
	}

	nc := notifications.NewNotificationsClient(endpointParams, startup.Endpoint{})

	if err = nc.SendNotification(notification, context.Background()); err != nil {
		edgexcontext.LoggingClient.Error(err.Error())
	}

	return false, nil
}

// SubscribeToNotificationService configures an email notification with edgex-go
func SubscribeToNotificationService(edgexSdk *appsdk.AppFunctionsSDK) error {

	edgexSdk.LoggingClient.Info("setting up subscription to edgex notification")

	appSettings := edgexSdk.ApplicationSettings()
	url, ok := appSettings[notificationsURL]
	if !ok {
		errorMessage := notificationsURL + " setting not found"
		edgexSdk.LoggingClient.Error(errorMessage)
		return errors.New(errorMessage)
	}

	endpoint := url + clients.ApiSubscriptionRoute

	notificationEmailAddresses, ok := appSettings["NotificationEmailAddresses"]
	if !ok {
		errorMessage := "NotificationEmailAddresses setting not found"
		edgexSdk.LoggingClient.Error(errorMessage)
		return errors.New(errorMessage)
	}

	emailAddresses := strings.Split(notificationEmailAddresses, ",")

	slug, ok := appSettings["NotificationSlug"]
	if !ok {
		errorMessage := "NotificationSlug setting not found"
		edgexSdk.LoggingClient.Error(errorMessage)
		return errors.New(errorMessage)
	}

	subscriptionMessage := map[string]interface{}{
		"slug":     slug,
		"receiver": "System Administrator",
		"subscribedCategories": []string{
			string(notifications.SECURITY),
		},
		"subscribedLabels": []string{
			string(notifications.SECURITY),
		},
		"channels": []map[string]interface{}{
			{
				"type":          "EMAIL",
				"mailAddresses": emailAddresses,
			},
		},
	}

	byteMessage, err := json.Marshal(subscriptionMessage)
	if err != nil {
		return err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(byteMessage))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict { // http.StatusConflict means the subscription is already created
		return errors.New("Error subscribing to notification http status code: " + resp.Status)
	}

	return nil
}
