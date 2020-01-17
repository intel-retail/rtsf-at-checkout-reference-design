
# How to Send Notifications through EdgeX (optional)

This section provides instructions to help you configure the EdgeX notifications service to send alerts through SMS, email, Rest calls, and others. 

Notifications work as follows: 

1. When the reconciler receives a payment-start event, it sends a message to the loss-detector that contain the suspect items list. 

2. The loss-detector sends these alerts as email messages through the EdgeX notification service. 

3. The loss-detector initiates the connection to the EdgeX notifications service.

To change the message type from email to a different medium, you must update the loss-detector.


## Step 1: Set Environment Variables
Set environment variable overrides for `Smtp_Host` and `Smtp_Port` in 'config-seed', which will inject these variables into the notification service's registry. 

Additional notification service configuration properties are [here](https://docs.edgexfoundry.org/Ch-AlertsNotifications.html#configuration-properties "EdgeX Alerts & Notifications").

## Step 2: Add code to the config-seed Environment Section

The code snippet below is a docker-compose example that sends an email notification. Add this code to the config-seed environment section in `docker-compose.edgex.yml`, under the config-seed service.

``` yaml
environment:
  <<: *common-variables
  Smtp_Host: <host name>
  Smtp_Port: 25
  Smtp_Password: <password if applicable>
  Smtp_Sender: <some email>
  Smtp_Subject: EdgeX Notification Suspect List
```

## Step 3: Add SMTP Server to compose file (optional)

The snipped below adds a development SMTP server smtp4dev to your `docker-compose.loss-detection.yml`. 
Skip this step if you want to use Gmail or another server.

``` yaml
smtp-server:
  image: rnwood/smtp4dev:linux-amd64-v3
  ports:
    - "3000:80"
    - "2525:25"
  restart: "on-failure:5"
  container_name: smtp-server
  networks:
    - theft-detection-app_edgex-network
```
