# Real Time Sensor Fusion for Loss Detection at Checkout Reference Design

## Introduction
This guide helps you build and run the Real Time Sensor Fusion for Loss Detection at Checkout Reference Design.

Upon completing the steps in this guide, you will be ready to integrate sensors and services to build your own complete solution.

> *Important: This guide does not create a complete, ready-to-use solution. Instead, upon completing the steps in this guide, you will be ready to integrate sensors and services to build your own Real Time Sensor Fusion for Loss Detection at Checkout solution.*

### Block Diagram
The high-level diagram below shows the sensors and services used with the Real Time Sensor Fusion for Loss Detection at Checkout Reference Design. The diagram shows the sensors and services, and how they communicate through EdgeX. Intel provides the services outlined in green, and you must obtain EdgeX and the sensors and services in pink.
![Real Time Sensor Fusion for Loss Detection at Checkout Diagram](./rtsf-loss-detection-at-checkout.png)

### Prerequisites

The following items are required to build the Real Time Sensor Fusion for Loss Detection at Checkout Reference Design. You will need additional hardware and software when you are ready to build your own solution. 

- Point-of-Sale (POS) of choice. You must use a POS that integrates with either the EdgeX REST or MQTT Device. See [POS Events](./rtsf_at_checkout_events/checkout_events.md#pos-events) for information about integration with the [EdgeX Device Services](./device_services.md#edgex-rest-and-mqtt-device-services) to send POS Events
- Docker
- Docker-Compose
- Go 1.12+
- GIT
- make

### Recommended Domain Knowledge

- MQTT
- REST
- POS systems, including customizations (Transactions, Real-Time Transaction Log (RTTL))
- Computer Vision concepts, if using CV Components
- RFID concepts, if using RFID Components

## Getting Started

### Step 1: Clone the repository

```
git clone  https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design
```

### Step 2: Build the Reference Design

You must build the provided component services and create local docker images. To do so, run

```
make docker
```

which builds the following docker images

- `rtsf-at-checkout/event-reconciler`
- `rtsf-at-checkout/device-scale`
- `rtsf-at-checkout/product-lookup`
- `rtsf-at-checkout/loss-detector`
- `rtsf-at-checkout/rsp-controller-event-handler`
- `rtsf-at-checkout/cv-region-of-interest`

### Step 3: Build EdgeX REST Device service

The EdgeX REST device service is currently the process of being released. Until a release docker image is available via Edge Foundry, we must build it locally by running the following command:

```
make device-rest
```

which creates the following docker image:

- `edgexfoundry/docker-device-rest-go`

This make target will clone the device-rest-go repo and build the docker image.

### Step 4: Use docker-compose to start the reference design suite. To do so, use

```
make run-base
``` 

This command starts all the EdgeX services and then starts all the Loss Detection services.

### Step 5: Use the Event Simulator

Use the [Event Simulator](./rtsf_at_checkout_events/event_simulation.md) to test and explore the example reference design until you feel comfortable with it.

You have successfully created an example reference design and are ready to integrate your components to create your own Real Time Sensor Fusion for Loss Detection at Checkout solution.

### Step 6: Integrate your POS of choice

Your POS of choice must send **POS Events** to either the Edgex REST or MQTT Device service. See the [POS Events](./rtsf_at_checkout_events/checkout_events.md#pos_ events) section for details about each **POS event** and the [Device Services](./device_services.md) section for details on how to send the events. 

If you don't currently have a POS of choice, we recommend using the **uniCenta ** open source POS to get familiar with the POS integration. The [uniCenta ](./unicenta.md) section has details on installing and integrating **uniCenta **.

### Step 7: Integrate your Security Scale of choice 

TBD

### Step 8: (optional) Build the Intel® Retail Sensor Platform (Intel® RSP) Controller 

If you want to use the Intel® RSP application with your RFID sensors, run the following command:

```
make rsp
```

which will create these docker images:

- `rsp/mqtt-device-service`
- `rsp/sw-toolkit-gw`
- `rsp/avahi`
- `rsp/ntp`

This make target will clone the RSP repositories and run the appropriate scripts. Once complete you will need to stop the that docker stack that the scripts started so you can use the compose file provided by this reference design. This is required because the `rsp-sw-toolkit-installer` build script runs the RSP docker-compose file. To take down the docker stack:

```sh
docker-compose -p rsp -f $PROJECTS_DIR/rsp-sw-toolkit-installer/docker/compose/docker-compose.yml down
```

Run the following command to start the RSP services:

```
make run-rsp
```

### Step 9: (optional) Build Video Analytics Serving (Intel's Computer Vision Reference Design)** 

To use the Intel Video Analytics Serving application, run the following command:

```
make vas
```

which will create the following docker image:

- `video_analytics_serving_gstreamer`

This make target will clone the VAS repo and run the appropriate scripts to create the docker image.

Run the following command to start the VAP services:

```
make run-vap
```

## How to Use the Compose Files

The docker-compose files are divided up to let you bring up or take down individual sensor ingestion components. Visit each individual [GitHub subrepository](#references-and-links-to-the-documentation-of-the-individual-components) to learn about the responsibility of each service. 

| Compose File                    | Purpose                             | Command                                                     |
| ------------------------------- | ----------------------------------- | ----------------------------------------------------------- |
| Portainer                       | Container management                | `docker-compose -f docker-compose.portainer.yml up -d`      |
| EdgeX and its components        |                                     | `docker-compose -f docker-compose.edgex.yml up -d`          |
| Real Time Sensor Fusion for Loss Detection at Checkout Core Services and its components  | | `docker-compose -f docker-compose.loss-detection.yml up -d`|
| RFID Components (Intel® RSP SW Toolkit)  |                            | `docker-compose -f docker-compose.rsp.yml up -d`            |
| Video Analytics Pipeline (VAP)  |                                     | `docker-compose -f docker-compose.vap.yml up -d`            |


## Components to Build a Solution based on this Reference Design

The reference design you created is not a complete solution. It provides the base components for creating a sensor fusion based framework. It is your choice on how many and which sensors to include. This section provides information about components you might want to include or replace.

### Components to Consider Adding

| Component                             | Description                                                                                  |
| ------------------------------------- | -------------------------------------------------------------------------------------------- |
| Security scale in bagging area        |  A Scale Device service is provided for a CAS USB scale. As an alternative, you can have Scale Events sent to either the EdgeX REST or MQTT Device Services. For more information, see  [Scale Events](./rtsf_at_checkout_events/checkout_events.md#scale-events) and integrating with the [EdgeX Device Services](./device_services.md#edgex-rest-and-mqtt-device-services) to send the Scale Events.|
| Computer Vision (CV) Object Detection Model|  The Intel® Distribution of OpenVINO toolkit-based object detection model used by the Video Analytics Pipeline (VAP) service is provided as an example, but is not intended for your final solution. The VAP and CV ROI Enter Exit service in this reference design create the CV ROI Events that are sent to the [EdgeX MQTT Device Service](./device_services.md#edgex-mqtt-device-service). See [CV ROI Events](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events).|
| Intel® RFID Sensors                  | This reference design relies on the Intel® Retail Sensor Platform (Intel® RSP). (https://software.intel.com/en-us/retail/rfid-sensor-platform)** which has its own custom EdgeX Device Service which it sends RFID events. These events are transformed into RFID ROI Events by the provided **RSP Controller Event Handler** service. See below for details on [RFID ROI Events](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events).This RSP has its own custom EdgeX Device Service to which it sends RFID events. These events are transformed into RFID ROI Events by the provided RSP Controller Event Handler service. If you are interested in Intel® RSP RFID Sensors, see https://software.intel.com/en-us/retail/rfid-sensor-platform#buy|
| Reconciler Service                  | This service provided does the analytics of reconciling all the sensor events to identify suspect items. As an option, you can replace the provided service with a more advanced service.                    |
| Detector Service                    | The service provided demonstrates how to send an email notification using the EdgeX Notifications service. The contents of the email is a simple JSON list of suspect items. As an option, you can replace the provided service with a more advanced service. |
| Computer Vision Region of Interest (ROI) Solution | If you chose to create your own complete Computer Vision object detection and CV ROI enter/exit solution. Exclude running the components in the docker-compose.vap.yml compose file and remove the cv-region-of-interest from the docker-compose.loss-detection.yml compose file. Integrate your CV solution with either the EdgeX REST or MQTT Device Services to send the CV ROI events. |


### Components to Consider Replacing or Enhancing

| Component                     | Description                                                                                      |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| Reconciler service            |  The Reconciler Service performs the analytics of reconciling the sensor events to identify suspect items. While the service provided performs adequately, your analytics team might be able to improve on this reference implementation.
| Detector Service|  The Dectector Service demonstrates how to send email notifications using the EdgeX Notifications service. The email message content is a simple JSON list of suspect items.
| Product Lookup Service        | This service provided is a basic implementation of a Product Information Management Lookup service. It uses a JSON file as a database for the product information. It is recommended that you replace the service with an Enterprise Resource Planning (ERP) System.
| RFID services | If you choose to use a different RFID solution, in addition to providing your integration to generate the RFID ROI Events you will need to exclude running the components in the `docker-compose.rsp.yml` compose file and remove the `rsp-controller-event-handler` from the `docker-compose.loss-detection.yml` compose file.  You will also need to integrate your RFID solution with either the EdgeX REST or MQTT Device Services to send the RFID events. See below for details on [RFID ROI events](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events) and integrating with these [EdgeX Device Services](./device_services.md#edgex-rest-and-mqtt-device-services) to send RFID ROI events. |


## How to Send Notifications through EdgeX (optional)

This section provides instructions to help you configure the EdgeX notifications service to send alerts through SMS, email, Rest calls, and others. 

Notifications work as follows: 

1. When the reconciler receives a payment-start event, it sends a message to the loss-detector that contain the suspect items list. 

2. The loss-detector sends these alerts as email messages through the EdgeX notification service. 

3. The loss-detector initiates the connection to the EdgeX notifications service.

To change the message type from email to a different medium, you must update the loss-detector.


### Step 1: Set Environment Variables
Set environment variable overrides for `Smtp_Host` and `Smtp_Port` in 'config-seed', which will inject these variables into the notification service's registry. 

Additional notification service configuration properties are [here](https://docs.edgexfoundry.org/Ch-AlertsNotifications.html#configuration-properties "EdgeX Alerts & Notifications").

### Step 2: Add code to the config-seed Environment Section

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

### Step 3: Add SMTP Server to compose file (optional)

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


## Data Dictionary

The data dictionary table below describes the JSON object fields for the [Real Time Sensor Fusion for Loss Detection at Checkout Events](./rtsf_at_checkout_events/checkout_events.md#rtsf-at-checkout-events).

| Field Name      | Data Type | Events used in                                          | Description                                                  |
| --------------- | --------- | ------------------------------------------------------- | ------------------------------------------------------------ |
| lane_id         | string    | All                                                     | Unique identifier of the self checkout lane                  |
| event_time      | number    | All                                                     | Unix nanosecond timestamp of when the event occured          |
| basket_id       | string    | All [POS Events](./rtsf_at_checkout_events/checkout_events.md#pos-events)                           | Unique identifier of self checkout session basket            |
| customer_id     | string    | All [POS Events](./rtsf_at_checkout_events/checkout_events.md#pos-events)                           | Optional unique identifier of the customer using the self checkout. |
| employee_id     | string    | All [POS Events](./rtsf_at_checkout_events/checkout_events.md#pos-events)                           | Optional unique identifier of the employee overseeing the self checkout. |
| product_id      | string    | POS [Scanned Item](./rtsf_at_checkout_events/checkout_events.md#pos-events)                         | Unique identifier of the product that was scanned. <br />**This reference design expects a 14 digit UPC** |
| product_id_type | string    | POS [Scanned Item](./rtsf_at_checkout_events/checkout_events.md#pos-events)                         | Type of the associated product ID. Value can be `UPC`, `SKU`, or `PLU`. <br />**This reference design expects `UPC`** |
| product_name    | string    | POS [Scanned Item](./rtsf_at_checkout_events/checkout_events.md#pos-events)                         | Name of the product that was scanned                         |
| quantity        | number    | POS [Scanned Item](./rtsf_at_checkout_events/checkout_events.md#pos-events)                         | Quantity of the products that were scanned                   |
| quantity_unit   | string    | POS [Scanned Item](./rtsf_at_checkout_events/checkout_events.md#pos-events)                         | Units for the associated quantity. <br />Values can be `EA`, `Each`, `lbs`, `g`, `kg` or `oz`<br />**This reference design expects `EA`, `Each`, or `lbs`** |
| unit_price      | number    | POS [Scanned Item](./rtsf_at_checkout_events/checkout_events.md#pos-events)                         | Price of the individual products scanned.                    |
| scale_id        | string    | [Scale Item](./rtsf_at_checkout_events/checkout_events.md#scale-events)                             | Unique identifier for the scale sending the events           |
| total           | number    | [Scale Item](./rtsf_at_checkout_events/checkout_events.md#scale-events)                             | Total weight for items on the scale                          |
| units           | string    | [Scale Item](./rtsf_at_checkout_events/checkout_events.md#scale-events)                             | Units for the associated total weight.<br />Values can be `lbs`, `g`, `kg` or `oz` <br />**This reference design expects `lbs`** |
| product_name       | string    | [CV ROI](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events)                                | Unique identifier of the product identified by CV object detection. |
| roi_action      | string    | [CV ROI](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events) & [RFID ROI](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events) | Action of the associate product identified.<br />Value can be either `ENTERED` or `EXITED`. |
| roi_name        | string    | [CV ROI](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events) & [RFID ROI](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events) | Name of the region that the associate product was identified with. See the [Configuration](./configuration.md#configuration) section above for details on how these names after configured. |
| epc             | string    | [RFID ROI](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events)                            | Unique EPC code of the product identified by RFID            |

## References and Links

> TODO: Update Links to individual RTSF at Checkout when open sourced


### Application Services

| Service                       | Link                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| Reconciler                    | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-event-reconciler             |
| Product Lookup                | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-product-lookup               |
| RSP Controller Event Handler  | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-rsp-controller-event-handler |
| Loss Detector                 | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-loss-detector                |

### Device Service

| Service                       | Link                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| Device Scale                  | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-device-scale                 |


### Helper Services

| Service                       | Link                                                                                             |
| ----------------------------- | ------------------------------------------------------------------------------------------------ |
| CV ROI Service                | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-cv-region-of-interest        |
| Event Simulator               | https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-event-simulator              |


### Intel® RSP RFID Solution

| Component                      | Link                                                                                                           |
| ------------------------------ | -------------------------------------------------------------------------------------------------------------- |
| Intel® RSP Getting started            | https://software.intel.com/en-us/getting-started-with-intel-rfid-sensor-platform-on-linux-6-use-the-web-portal |
| Intel® RSP installer                  | https://github.com/intel/rsp-sw-toolkit-installer                                                              |
| Intel® RSP gateway                    | https://github.com/intel/rsp-sw-toolkit-gw                                                                     |
| Intel® RSP user guides                | https://github.com/intel/rsp-sw-toolkit-gw/tree/master/docs                                                    |

### Other 

| Component                                            | Link                                                                     |
| ---------------------------------------------------- | ------------------------------------------------------------------------ |
| Intel® Video Analytics Serving Documentation          | https://github.com/intel/video-analytics-serving                         |
| Intel® Distribution of OpenVINO toolkit Documentation | https://software.intel.com/en-us/openvino-toolkit/documentation/featured |
| EdgeX GitHub Repos                                   | https://github.com/edgexfoundry                                          |



