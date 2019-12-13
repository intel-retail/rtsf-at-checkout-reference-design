# Real Time Sensor Fusion for Loss Detection at Checkout Reference Design

## Prerequisites

- Docker 
- Docker-Compose 
- Go 1.12+
- GIT
- make

The following domain knowledge is recommended:

- MQTT and REST
- Point-of-Sale (POS)
  - Knowledge of POS Systems (Transactions, Real-Time Transaction Log (RTTL))
  - You will need your own POS software to integrate
  - Know how to customize it
- Familiarity with CV Concepts if using CV Components
- Familiarity of RFID Concepts if using RFID Components

## Block Diagram

The following diagram below helps to understand, at a high-level, the services involved in the Real Time Sensor Fusion for Loss Detection at Checkout Reference Design (RTSF at Checkout). It is not necessarily a 1:1 mapping to each service, however it does provide a visual of all sensors involved, how EdgeX is situated among the services, and the services responsible for reconciling the sensor data that is consumed from EdgeX. 

![ RTSF at Checkout Architecture Diagram ](./rtsf-loss-detection-at-checkout.png)

## How to use this reference design

### Getting Started

Use the following command to clone the repository.

​	`git clone https://github.impcloud.net/loss-detection/rtsf-at-checkout-reference-design`

**[TODO: update repo URL for open source location]**

### Building the required components

1. **Building the RTSF Reference Design** (required)

    The provided component services must be built and local docker images created. The docker images can be built by running `make docker`.  See [Getting Started](#getting-started) above for details on cloning this repo and it's sub modules.

2. **Building RSP Controller (Intel's RFID Reference Design)** (optional)

    To use the Intel RSP RFID application with your RFID sensors, you will need to build the following docker images `rsp-gw`, `avahi`, `ntp` this can be done by running the installer located [here](https://github.com/intel/rsp-sw-toolkit-installer#docker-environment "rsp-sw-toolkit-installer") inside the docker folder.  You will also need to build the docker image for the `rsp-mqtt-device-service` this service can be built by following the install instructions [here](https://github.com/intel/rsp-sw-toolkit-im-suite-mqtt-device-service "RSP MQTT Device Service"). Once these components have been installed you can continue.  

    > *Note that for both the installers you only need to build the docker images, you don't need to create or copy any docker-compose files, the compose file for rsp is already created in this project `rtsf-at-checkout-reference-design/loss-detection-app/docker-compose.rfid.yml`.*  

    The rsp-sw-toolkit-installer's build script runs the docker-compose file for you, you **must** take down that docker stack and use the compose file provided by the reference design.  This is done by running the following command.

    ```sh
    docker-compose -p rsp -f $PROJECTS_DIR/rsp-sw-toolkit-installer/docker/compose/docker-compose.yml down
    ```

3. **Building Video Analytics Serving (Intel's CV Reference Design)** (optional)

    To use the Intel Video Analytics Serving application follow the instructions [here](https://github.com/intel/video-analytics-serving "RSP MQTT Device Service") to build the `video-analytic` docker image.  The compose file for Video Analytics Serving is `rtsf-at-checkout-reference-design/loss-detection-app/docker-compose.vap.yml`

First the provided component services must be built and local docker images created. This is done by first cloning this repository and then executing  `make docker`.  See  [Getting Started](#getting-started) above for details on cloning this repo and it's sub modules.

### Running the solution

Once the above steps are complete the simplest way to get everything up and running is to execute the `make run` command to startup the reference design suite using docker-compose. Now that this reference design is running, you can use the [Event Simulator](./rtsf_at_checkout_events/checkout_events.md) to test it out and explore how it all works. Once you that is complete you are ready to integrate your components to create a complete `RTSF at Checkout` solution.

> *Note: ALL of the services in these reference design are meant to be reference services and are intended to be examples to be built/improved upon to create your complete solution.*

### Compose Files

The docker-compose files are broken out in a way that allows bringing up or down individual sensor ingestion components as needed. Be sure to visit each individual GitHub sub-repositories listed [here](#references-and-links-to-the-documentation-of-the-individual-components) for an understanding of the responsibility of each service. 

Portainer is used for management of the various containers.
`docker-compose -f docker-compose.portainer.yml up -d`

EdgeX and its required components can be run with: 
`docker-compose -f docker-compose.edgex.yml up -d`

RTSF at Checkout Core Services and its required components  can be run with:
`docker-compose -f docker-compose.loss-detection.yml up -d`

RFID Components (Intel® RSP SW Toolkit) can be run with:
`docker-compose -f docker-compose.rsp.yml up -d`

The Video Analytics Pipeline (VAP) can be run with: 
`docker-compose -f docker-compose.vap.yml up -d`

## How to build a solution using this reference design

This reference design does not provide a complete solution. It provides the base components for creating a sensor fusion based framework.  It is your choice on how many and which sensors to include. Further more, you will need to provide the follow components to complete your RTSF at Checkout solution.

1. **Point of Sale (POS)** (required)
   The POS system of your choice will need to integrate with either the EdgeX REST or MQTT Device Services to send the POS Events. See below for details on [POS Events](./rtsf_at_checkout_events/checkout_events.md#pos-events) and integrating with the [Edgex Device Services](./device_services/index.md#edgex-rest-and-mqtt-device-services) to send the POS Events.

2. **Security (bagging area) Scale** (optional)

   The Scale Device Service provided is specifically for a CAS USB scale. It is a good starting point for integrating other USB scales. 

   Alternatively, the Scale Events can be sent to either the EdgeX REST or MQTT Device Services. See below for details on [Scale Events](./rtsf_at_checkout_events/checkout_events.md#scale-events) and integrating with the [Edgex Device Services](./device_services/index.md#edgex-rest-and-mqtt-device-services)  to send the Scale Events.

3. **Computer Vision Object Detection Model** (optional)
   
   The OPENVINO™ based object detection model used by the Video Analytics Pipeline (VAP) service is  provided as an example and is not of production quality. For a robust solution you will need to provide an improved object detection model. 
   
   The VAP and CV ROI Enter Exit service used in this reference design create the CV ROI Events which are sent to the [EdgeX MQTT Device Service](./device_services/index.md#edgex-mqtt-device-service). See below for details on [CV ROI Events](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events).

4. **Intel RFID Sensors** (optional)
   This reference design relies on **[Intel Retail Sensor Platform (RSP)](https://software.intel.com/en-us/retail/rfid-sensor-platform)** which has its own custom EdgeX Device Service which it sends RFID events. These events are transformed into RFID ROI Events by the provided **RSP Controller Event Handler** service. See below for details on [RFID ROI Events](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events). 

   Available Intel RSP RFID Sensors:  https://software.intel.com/en-us/retail/rfid-sensor-platform#buy . 

In addition you may want to replace or enhance the follow components.

1. **Reconciler service**
   This service does the analytics of reconciling all the sensor events to identify suspect items. While it does a solid job, you may have your own analytics team which can improve on this reference implementation.
2. **Detector Service**
   This service simply demonstrates how to send an email notification using the EdgeX Notifications service. The contents of the email is a simple JSON list of suspect items.
3. **Product Lookup Service**
    This service is a very basic implementation of a Product Information Management Lookup service. It simply uses a JSON file as a database for the product information. This service should be replaced with an interface to a real Product Information Management system; commonly referred to as an Enterprise Resource Planning (ERP) System.
4. **CV Region of Interest (ROI) Solution**
   If you chose to create your own complete CV object detection and CV ROI enter/exit solution you will need to exclude running the components in the `docker-compose.vap.yml` compose file and remove the `cv-region-of-interest` from the `docker-compose.loss-detection.yml` compose file. You will also need to integrate your CV solution with either the EdgeX REST or MQTT Device Services to send the CV ROI events. See below for details on [CV ROI events](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events) and integrating with the [Edgex Device Services](./device_services/index.md#edgex-rest-and-mqtt-device-services)  to send CV ROI events.
5. **RFID services**
   If you choose to use a different RFID solution, in addition to providing your integration to generate the RFID ROI Events you will need to exclude running the components in the `docker-compose.rsp.yml` compose file and remove the `rsp-controller-event-handler` from the `docker-compose.loss-detection.yml` compose file.  You will also need to integrate your RFID solution with either the EdgeX REST or MQTT Device Services to send the RFID events. See below for details on [RFID ROI events](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events) and integrating with these [Edgex Device Services](./device_services/index.md#edgex-rest-and-mqtt-device-services) to send RFID ROI events.

6. **Sending Notifications Through EdgeX** (optional)

    The EdgeX notifications service can be configured to send alerts via SMS, Email, Rest call and various other means. First, you'll need to set environment variable overrides for `Smtp_Host` and `Smtp_Port` in config-seed, config-seed will inject these variables into the notification service's registry. Additional notification service configuration properties can be found [here](https://docs.edgexfoundry.org/Ch-AlertsNotifications.html#configuration-properties "EdgeX Alerts & Notifications").

    Below is the example docker-compose snippets for sending an Email notification.

    This is the config-seed environment section add this to `docker-compose.edgex.yml` under the config-seed service.

    ```toml
    environment:
      <<: *common-variables
      Smtp_Host: <host name>
      Smtp_Port: 25
      Smtp_Password: <password if applicable>
      Smtp_Sender: <some email>
      Smtp_Subject: EdgeX Notification Suspect List
    ```

    this snippet adds a development SMTP server smtp4dev to your `docker-compose.loss-detection.yml`, if you want to use Gmail or another server this step can be skipped.

    ```toml
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

    When the reconciler receives a payment-start event it will send a message to the loss-detector containing the suspect items list.  The loss-detector sends these alerts as emails through the EdgeX notification service. The loss-detector initiates the connection to the EdgeX notifications service, so to change the message type from email to something else you would need to update the loss-detector.

## Data Dictionary

The following is the data dictionary for the fields found in the JSON object for all the above [RTSF at Checkout Events](./rtsf_at_checkout_events/checkout_events.md#rtsf-at-checkout-events).

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
| roi_name        | string    | [CV ROI](./rtsf_at_checkout_events/checkout_events.md#cv-roi-events) & [RFID ROI](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events) | Name of the region that the associate product was identified with. See the [Configuration](./configuration/configuration.md#configuration) section above for details on how these names after configured. |
| epc             | string    | [RFID ROI](./rtsf_at_checkout_events/checkout_events.md#rfid-roi-events)                            | Unique EPC code of the product identified by RFID            |

## References and links to the documentation of the individual components

> TODO: Update Links to individual RTSF at Checkout when open sourced


#### Application Services

Reconciler:  https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-event-reconciler 

Product Lookup:  https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-product-lookup 

RSP Controller Event Handler: https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-rsp-controller-event-handler

Loss Detector:  https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-loss-detector 

#### Device Services

Device Scale:  https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-device-scale 

#### Helper Services

CV ROI Service:  https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-cv-region-of-interest 

Event Simulator:  https://github.impcloud.net/loss-prevention-at-pos/rtsf-at-checkout-event-simulator 

#### Intel RSP RFID Solution

RSP Getting started: https://software.intel.com/en-us/getting-started-with-intel-rfid-sensor-platform-on-linux-6-use-the-web-portal  

RSP installer: https://github.com/intel/rsp-sw-toolkit-installer  

RSP gateway: https://github.com/intel/rsp-sw-toolkit-gw 

RSP user guides: https://github.com/intel/rsp-sw-toolkit-gw/tree/master/docs 

#### Other 

Documentation for Intel Video Analytics Serving: 
https://github.com/intel/video-analytics-serving  

Documentation for OPENVINO™: 
https://software.intel.com/en-us/openvino-toolkit/documentation/featured 

GitHub Repos for EdgeX: 
https://github.com/edgexfoundry
