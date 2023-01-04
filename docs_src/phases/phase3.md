# Phase 3 - Bring Your Own Hardware

## Overview

After [phase 2](./phase2.md) has been completed, the next step is to integrate physical hardware. This guide will assist you in understanding the pieces of hardware used in the reference solution to provide better fidelity for loss prevention.

## Getting Started


### Step 1: Integrate your Security Scale of choice 

The first piece of hardware that we suggest to integrate is a security weight scale. This scale typically is located in the bagging area or otherwise after the scanner where items are placed before going back into a customer's cart. We have provided a reference design for a USB based weight scale [here](https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/tree/master/rtsf-at-checkout-device-scale). Anytime an item is placed on the scale, or subsequently removed from the scale a `weight` event is sent into EdgeX to be reconciled in the reconciler. While the reference service we provided is USB based, you may use any protocol you prefer for integrating your own device service.

### Step 2: (optional) Build Video Analytics Pipeline Server(Intel's Computer Vision Reference Design)** 

!!! note
    To ensure you have all required components, ensure you review the documentation provided for the [Intel® Video Analytics Pipeline Server Documentation](../references.md#other) located in the references section under "Other"

To use the Intel Video Analytics Pipeline Server application, run the following command:

``` sh
make models
```

which will download the models needed from Intel ModelZoo that can be run in following docker image:

- `intel/dlstreamer-pipeline-server`

This make target can run the Intel video analytcs pipeline server.

### Environment variable section of docker-compose.vap.yml

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

- MQTT_DESTINATION_HOST - the destination of the MQTT broker, to use the MQTT broker running in the compose file simply just use `edgex-mqtt-broker:1883` 
- CAMERA{i}_ENDPOINT - the destination of the pipeline resource you wish to run the analytics on. Use [http://pipeline-server:8080](http://pipeline-server:8080/) to hit the container running the VAP software then just add the route to the pipeline you are targeting. 
- CAMERA{i}_SRC - the path to the camera either RTSP or the Linux path to the camera resource 
- CAMERA{i}_ROI_NAME - the name of the ROI location this camera represents, this is a configurable parameter of the reconciler also and these ROI locations will need to match what is defined there as well. 
- CAMERA{i}_CROP_TBLR - the camera’s crop number of pixels to cut from the top, bottom, left, right. Try and crop the camera to just include the ROI of interest.
- CAMERA{i}_FRAME_STORE - the storage for image frames

Each of the CAMERA{i}_ environment variables can be incremented up to the number of cameras you wish to include in your solution.  For example, if you have two cameras your environment variables section would have the following values: 

- MQTT_DESTINATION_HOST
- CAMERA0_ENDPOINT
- CAMERA0_SRC
- CAMERA0_ROI_NAME
- CAMERA0_CROP_TBLR
- CAMERA0_FRAME_STORE
- CAMERA1_ENDPOINT
- CAMERA1_SRC
- CAMERA1_ROI_NAME
- CAMERA1_CROP_TBLR
- CAMERA1_FRAME_STORE

### Camera Source
Before you run Video Analytics Pipeline Server you'll need to specify your camera source in the docker compose file `docker-compose.vap.yml`.

Possible camera sources include:

- rtsp
- usb camera
- local file (debugging only)

#### RTSP

For an RTSP feed edit the CAMERA0_SRC environment variable to point to the URL of the RTSP feed.

`CAMERA0_SRC=rtsp://{URL}:{PORT}/{STREAM-END-POINT}`

If the RTSP camera has username and password auth append those values to the beginning of the URL. i.e. `CAMERA0_SRC=rtsp://{USERNAME}:{PASSWORD}@{URL}:{PORT}/{STREAM-END-POINT}`.

#### USB camera
First, add the camera device to the `pipeline-server` service in the docker-compose file.

``` yaml
devices:
    - /dev/video0:/dev/video0
```

Then edit the CAMERA0_SRC environment variable to point to the video source
`CAMERA0_SRC=/dev/video0`

#### Local video file

To use a local video sample add the path to the video file in the container to the CAMERA0_SRC environment variable.  The keyword `file` lets VAP know the URI is a local file and not a RTSP URL.

`CAMERA0_SRC=file:///home/pipeline-server/video-samples/EnterExitEvent_2second.mp4`

### Pipelines

There are a few pipelines included with this build, you can find them in the `pipelines` directory.  The pipelines are Object Detection Pipeline, and Object detection pipeline with frame store support.  

- default is the default pipeline for object detection
- frame_store is Object detection pipeline with frame store support.

### Starting the Video Analytics Pipeline Server Container

Once the configuration changes are applied to the compose file run the following command to start the VAP services:

``` sh
make run-vap
```

Once VAP is up and running you can check the logs for `cv-region-of-interest` and `pipeline-server`.  These inference results are sent to EdgeX and handled by reconciler service to ensure products are entering and exiting the correct regions of interest.

## End Results

By the end of this guide, you should now have an understanding of the hardware involved in completing the RTSF at Checkout Reference Solution. You should now be confident enough in adding your own hardware devices and begin swapping out components that better suit your target deployment scenario.
