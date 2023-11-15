## Scale Device Service

The security scale used in this reference design is the [CAS PD-2 POS/Checkout Scale](http://www.cas-usa.com/products/pd-2) which communicates over a serial connection. The vid:pid values of the scale are `VID = "0403"` and `PID = "6001"`. These values change depending on the model of scale you are using.  If you use a different scale, the logic to read and write to the scale will most likely be different. In this case, simply use this service as a reference for creating your custom Scale Device Service. 

The following items can be configured via the `Driver` section of the service's [configuration.toml](https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/blob/master/rtsf-at-checkout-device-scale/cmd/res/configuration.toml) file. All values are strings.  

- ScaleVID - VID value for the scale
- ScalePID - PID value for the scale
- ScaleID - ID of the scale
- LaneID - ID of the checkout lane the where the scale is being used
- TimeOutMilli - Time out for when reading from the scale in milliseconds 

## EdgeX MQTT Device Service

This reference design uses the [MQTT Device Service](https://github.com/edgexfoundry/device-mqtt-go) from EdgeX with custom device profiles. These device profiles YAML files are located at [https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/tree/master/loss-detection-app/res/device-mqtt/profiles](https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/tree/master/loss-detection-app/res/device-mqtt/profiles) and are volume mounted into the device service's running Docker container.

## EdgeX REST Device Service

This reference design uses the [REST Device Service](https://github.com/edgexfoundry/device-rest-go) from EdgeX Foundry with custom device profiles. These device profiles YAML files are located at [https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/rtsf-at-checkout-reference-design/tree/master/loss-detection-app/res/device-rest/profiles](https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/rtsf-at-checkout-reference-design/tree/master/loss-detection-app/res/device-rest/profiles]) and are volume mounted into the device service's running Docker container.

## CV ROI Service & Video Analytics Pipeline Server 

CV consists of two components `pipeline-server` and `cv-region-of-interest`. `pipeline-server` runs and maintains the CV pipelines.  `cv-region-of-interest` configures the pipelines, collects the frame by frame CV data and generates the CV ROI Events to send to EdgeX. 

To configure the VAP pipeline you need to set the following environment variables in the docker-compose file under `cv-region-of-interest`. All the values other than the MQTT_DESTINATION_HOST are in the form CAMERA{N} where N is the camera number starting with 0. 

In the `docker-compose.vap.yml` file under the `environment` section of `cv-region-of-interest`

- MQTT_DESTINATION_HOST - the destination of the MQTT broker, to use the MQTT broker running in the compose file simply just use `edgex-mqtt-broker:1883` 
- CAMERA{i}_ENDPOINT - the destination of the pipeline resource you wish to run the analytics on. Use [http://pipeline-server:8080](http://pipeline-server:8080/) to hit the container running the VAP software then just add the route to the pipeline you are targeting. 
- CAMERA{i}_SRC - the path to the camera either RTSP or the Linux path to the camera resource 
- CAMERA{i}_ROI_NAME - the name of the ROI location this camera represents, this is a configurable parameter of the reconciler also and these ROI locations will need to match what is defined there as well. 
- CAMERA{i}_CROP_TBLR - the camera’s crop number of pixels to cut from the top, bottom, left, right. Try and crop the camera to just include the ROI of interest.
- CAMERA{i}_FRAME_STORE - the storage for image frames

Here is an example configuration for one camera 

``` yaml
environment:
   - no_proxy=pipeline-server,edgex-mqtt-broker
   - MQTT_DESTINATION_HOST=edgex-mqtt-broker:1883
   - CAMERA0_ENDPOINT=http://pipeline-server:8080/pipelines/product_detection/default
   # - CAMERA0_ENDPOINT=http://pipeline-server:8080/pipelines/product_detection/frame_store
   # Fill in camera source info below
   # for RTSP cameras
   - CAMERA0_SRC=file:///home/pipeline-server/video-samples/EnterExitEvent_2second.mp4
   # or for USB cameras
   # - CAMERA0_SRC=/dev/video0
   - CAMERA0_ROI_NAME=Staging
   - CAMERA0_CROP_TBLR=0,0,0,0
   - CAMERA0_FRAME_STORE=/tmp/my-frame-store
```

For the `pipeline-server` service the following configuration is needed. Adding the device `/dev/dri:/dev/dri` which mounts the GPU is optional. 

``` yaml
environment:
  - ENABLE_RTSP=true
  - EMIT_SOURCE_AND_DESTINATION=true
devices: 
  - /dev/dri:/dev/dri 
volumes:
  - /tmp:/tmp
  - ./pipelines:/home/pipeline-server/pipelines
  - ./models:/home/pipeline-server/models
  - ./extensions:/home/pipeline-server/extensions
  - ./video-samples:/home/pipeline-server/video-samples
```

For additional support see the Video Analytics Pipeline Server repo [https://github.com/dlstreamer/pipeline-server](https://github.com/dlstreamer/pipeline-server).  

## Product Lookup

Product information values are stored in a JSON file. This inventory is used to lookup product information such as the product name, barcode, maximum and minimum weights and whether a product is RFID eligible.  

Example product lookup inventory is shown below:  

``` json
[{ 
    "barcode": "00022000008916", 
    "name": "Extra Peppermint Gum", 
    "min_weight": 0.102, 
    "max_weight": 0.109, 
    "rfid_eligible": true 
}, 
{ 
    "barcode": "00051700988235", 
    "name": "Finish Dishwasher Tablet", 
    "min_weight": 0.610, 
    "max_weight": 0.620, 
    "rfid_eligible": false 
}, 
{ 
    "barcode": "00012000163173", 
    "name": "Mountain Dew 6 Pack", 
    "min_weight": 3.200, 
    "max_weight": 3.255, 
    "rfid_eligible": true 
}, 
{ 
    "barcode": "00024000566670", 
    "name": "Canned Green Beans", 
    "min_weight": 1.025, 
    "max_weight": 1.050, 
    "rfid_eligible": false 
}]
```
## Checkout Event Reconciler

The following Checkout Event Reconciler service settings can be configured. All these settings are contained in the service’s `Reconciler` configuration section. All values are strings. 

- DeviceNames - Comma separated list of device names that the service filters incoming data for. These correspond to the device names defined by the device services. 

- DevicePos - Name of the device that the POS events are received from. Must be one of the device names listed in DeviceNames above. 

- DeviceScale - Name of the device that the Scale events are received from. Must be one of the device names listed in DeviceNames above. 

- DeviceCV - Name of the device that the CV ROI events are received from. Must be one of the device names listed in DeviceNames above. 

- DeviceRFID - Name of the device that the RFID ROI events are received from. Must be one of the device names listed in DeviceNames above. 

- CvTimeAlignment - The time period between when a product was scanned at the POS and when it was seen by CV. This is to allow for lag between CV and the time of scan, items are reconciled only if they are within this time window, otherwise they will remain unreconciled. Values are in time duration format i.e. “5s” or “5ms” and negative values allow for an unlimited CV reconciliation window. 

- ProductLookupEndpoint - URL for the Product Lookup service 

- WebSocketPort - Port number for the WebSocket that the service write data. Useful for connecting UI to receive the reconciler data. 

- ScaleToScaleTolerance - Allowable difference in weight values from the scanner scale and the security (bagging) scale. Required when product quantity is a weight. Value is a fraction of LBS., I.e. “0.02” 

## Loss Detector

The following Loss Detector service settings can be configured. All these settings are contained in the service’s `ApplicationSettings` configuration section. All values are strings. 

- NotificationEmailAddresses - Comma separated list email addresses to send notifications 

- NotificationName - Unique identifier used to subscribe with the EdgeX Support Notifications Service to send notifications 

## Checkout Event Simulator

The checkout event simulator configuration contains RESTful endpoints and MQTT endpoint to send the simulated data and MQTT topic.  These settings are contained in the `config.json` file. 

``` json
{
    "pos_endpoint": "http://localhost:59986/api/v3/resource/pos-rest",
    "scale_endpoint": "http://localhost:59986/api/v3/resource/scale-rest",
    "cv_roi_endpoint": "http://localhost:59986/api/v3/resource/cv-roi-rest",
    "rfid_roi_endpoint": "http://localhost:59986/api/v3/resource/rfid-roi-rest"
}
```
