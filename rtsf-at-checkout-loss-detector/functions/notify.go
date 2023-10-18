// Copyright © 2022 Intel Corporation. All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause

package functions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v3/pkg/interfaces"
	clientInterfaces "github.com/edgexfoundry/go-mod-core-contracts/v3/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/google/uuid"
)

const (
	notificationsURL       = "NotificationsURL"
	sender                 = "Loss Detector"
	securityCategory       = "SECURITY"
	notificationReceiver   = "SystemAdministrator"
	subscriptionAdminState = "UNLOCKED"
)

// NotifySuspectList sends a notification to edgex-go
func NotifySuspectList(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	lc := ctx.LoggingClient()

	lc.Info("Notifying suspect list")

	contentBytes, _ := data.([]byte)
	contentFormated, err := json.MarshalIndent(json.RawMessage(contentBytes), "", "  ")

	if err != nil {
		lc.Error("Failed to post notification, %s", err.Error())
		return false, nil
	}
	notificationClient := ctx.NotificationClient()
	if notificationClient == nil {
		lc.Error("cannot send notification: NotificationsClient is not configured or missing")
		return false, nil
	}

	content := "Suspicious Items:\n" + string(contentFormated)

	notification := dtos.NewNotification(
		[]string{
			string(securityCategory),
		},
		securityCategory,
		content,
		"Loss Detector",
		models.Critical,
	)

	req := requests.NewAddNotificationRequest(notification)
	_, err = notificationClient.SendNotification(context.Background(), []requests.AddNotificationRequest{req})
	if err != nil {
		lc.Error(err.Error())
	}

	return false, nil
}

// SubscribeToNotificationService configures an email notification with edgex-go
func SubscribeToNotificationService(appService interfaces.ApplicationService, subscriptionClient clientInterfaces.SubscriptionClient, lc logger.LoggingClient) error {

	lc.Info("setting up subscription to edgex notification")

	emailAddresses, err := appService.GetAppSettingStrings("NotificationEmailAddresses")
	if err != nil {
		errorMessage := "NotificationEmailAddresses setting not found"
		lc.Error(errorMessage)
		return errors.New(errorMessage)
	}
	notificationName, err := appService.GetAppSetting("NotificationName")
	if err != nil {
		errorMessage := "NotificationName setting not found"
		lc.Error(errorMessage)
		return errors.New(errorMessage)
	}

	dto := dtos.Subscription{
		Id:   uuid.NewString(),
		Name: notificationName,
		Channels: []dtos.Address{
			{
				Type:         "EMAIL",
				EmailAddress: dtos.EmailAddress{Recipients: emailAddresses},
			},
		},
		Receiver: notificationReceiver,
		Labels: []string{
			securityCategory,
		},
		Categories: []string{
			securityCategory,
		},
		AdminState: subscriptionAdminState,
	}
	reqs := []requests.AddSubscriptionRequest{requests.NewAddSubscriptionRequest(dto)}
	_, err = subscriptionClient.Add(context.Background(), reqs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to the EdgeX notification service: %s", err.Error())
	}

	return nil
}
