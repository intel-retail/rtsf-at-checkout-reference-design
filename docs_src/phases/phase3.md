# Phase 3 - Bring Your Own Hardware

## Overview

After [phase 2](./phase2.md) has been completed, the next step is to integrate physical hardware. This guide will assist you in understanding the pieces of hardware used in the reference solution to provide better fidelity for loss prevention.

## Getting Started


### Step 1: Integrate your Security Scale of choice 

The first piece of hardware that we suggest to integrate is a security weight scale. This scale typically is located in the bagging area or otherwise after the scanner where items are placed before going back into a customer's cart. We have provided a reference design for a USB based weight scale [here](https://github.com/intel-iot-devkit/rtsf-at-checkout-reference-design/tree/master/rtsf-at-checkout-device-scale). Anytime an item is placed on the scale, or subsequently removed from the scale a `scale-item` event is sent into EdgeX to be reconciled in the reconciler. While the reference service we provided is USB based, you may use any protocol you prefer for integrating your own device service.

### Step 2: (optional) Build the Intel® Retail Sensor Platform (Intel® RSP) Controller 

!!! note 
    You must have the Intel® RSP RFID Solution hardware which can be purchased by following the links in the references section provided [here](../references.md#intel-rsp-rfid-solution). 
    
If you want to use the Intel® RSP application with your RFID sensors, run the following command:

``` sh
make rsp
```

which will create these docker images:

- `rsp/mqtt-device-service`
- `rsp/sw-toolkit-gw`
- `rsp/avahi`
- `rsp/ntp`

This make target will clone the RSP repositories and run the appropriate scripts. Once complete you will need to stop the that docker stack that the scripts started so you can use the compose file provided by this reference design. This is required because the `rsp-sw-toolkit-installer` build script runs the RSP docker-compose file. To take down the docker stack:

``` sh
docker-compose -p rsp -f $PROJECTS_DIR/rsp-sw-toolkit-installer/docker/compose/docker-compose.yml down
```

Run the following command to start the RSP services:

``` sh
make run-rsp
```



### Step 3: (optional) Build Video Analytics Serving (Intel's Computer Vision Reference Design)** 

!!! note
    To ensure you have all required components, ensure you review the documentation provided for the [Intel® Video Analytics Serving Documentation](../references.md#other) located in the references section under "Other"

To use the Intel Video Analytics Serving application, run the following command:

``` sh
make vas
```

which will create the following docker image:

- `video_analytics_serving_gstreamer`

This make target will clone the VAS repo and run the appropriate scripts to create the docker image.

### Environment variable section of docker-compose.vap.yml

``` yaml
environment:
    - MQTT_DESTINATION_HOST=mqtt:1883
    - CAMERA0_ENDPOINT=http://video-analytic:8080/pipelines/object_detection_cpu_render/1
    # Fill in camera source info below
    # for RTSP cameras
    # - CAMERA0_SRC=rtsp://{URL}:{PORT}/{STREAM-END-POINT}
    # or for USB cameras
    # - CAMERA0_SRC=/dev/video0
    - CAMERA0_ROI_NAME=Staging
    - CAMERA0_CROP_TBLR=0,0,0,0
```

- MQTT_DESTINATION_HOST - the destination of the MQTT broker, to use the MQTT broker running in the compose file simply just use `mqtt:1883` 
- CAMERA{i}_ENDPOINT - the destination of the pipeline resource you wish to run the analytics on. Use [http://video-analytic:8080](http://video-analytic:8080/) to hit the container running the VAS software then just add the route to the pipeline you are targeting. 
- CAMERA{i}_SRC - the path to the camera either RTSP or the Linux path to the camera resource 
- CAMERA{i}_ROI_NAME - the name of the ROI location this camera represents, this is a configurable parameter of the reconciler also and these ROI locations will need to match what is defined there as well. 
- CAMERA{i}_CROP_TBLR - the camera’s crop number of pixels to cut from the top, bottom, left, right. Try and crop the camera to just include the ROI of interest.

Each of the CAMERA{i}_ environment variables can be incremented up to the number of cameras you wish to include in your solution.  For example, if you have two cameras your environment variables section would have the following values: 

- MQTT_DESTINATION_HOST
- CAMERA0_ENDPOINT
- CAMERA0_SRC
- CAMERA0_ROI_NAME
- CAMERA0_CROP_TBLR
- CAMERA1_ENDPOINT
- CAMERA1_SRC
- CAMERA1_ROI_NAME
- CAMERA1_CROP_TBLR

### Camera Source
Before you run Video Analytics Serving you'll need to specify your camera source in the docker compose file `docker-compose.vap.yml`.

Possible camera sources include:

- rtsp
- usb camera
- local file (debugging only)

#### RTSP

For an RTSP feed edit the CAMERA0_SRC environment variable to point to the URL of the RTSP feed.

`CAMERA0_SRC=rtsp://{URL}:{PORT}/{STREAM-END-POINT}`

If the RTSP camera has username and password auth append those values to the beginning of the URL. i.e. `CAMERA0_SRC=rtsp://{USERNAME}:{PASSWORD}@{URL}:{PORT}/{STREAM-END-POINT}`.

#### USB camera
First, add the camera device to the `video-analytic` service in the docker-compose file.

``` yaml
devices:
    - /dev/video0:/dev/video0
```

Then edit the CAMERA0_SRC environment variable to point to the video source
`CAMERA0_SRC=/dev/video0`

#### Local video file

To use a local video sample add the path to the video file in the container to the CAMERA0_SRC environment variable.  The keyword `file` lets VAS know the URI is a local file and not a RTSP URL.

`CAMERA0_SRC=file:///home/video-analytics-serving/video-samples/grocery-test.mp4`

### Pipelines

There are a few pipelines included with this build, you can find them in the `pipelines` directory.  The pipelines are object_detection_cpu, object_detection_cpu_render, and object_detection_cpu_restream.  

- object_detection_cpu_render is the default pipeline which opens a x11 window to show the inference results in realtime.  (when using this default pipeline in order to let docker's x11 client create a window be sure to run `xhost +` at the command prompt to disable x11 access control temporally)
- object_detection_cpu_restream is similar but instead of opening an x11 window the pipeline starts an RTSP server where you can view the realtime inference results.
- object_detection_cpu does the same inferencing but doesn't display any results.

For all of the pipelines version 1 is rtsp or local file and version 2 is USB camera.

### Starting the Video Analytics Container

Once the configuration changes are applied to the compose file run the following command to start the VAP services:

``` sh
xhost +
make run-vap
```

Once VAS is up and running you can check the logs for `cv-region-of-interest` and `video-analytic`.  And you should see a x11 window showing the inference results.  These inference results are sent to EdgeX and handled by reconciler service to ensure products are entering and exiting the correct regions of interest.

## End Results

By the end of this guide, you should now have an understanding of the hardware involved in completing the RTSF at Checkout Reference Solution. You should now be confident enough in adding your own hardware devices and begin swapping out components that better suit your target deployment scenario.
