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

Run the following command to start the VAP services:

``` sh
make run-vap
```



## End Results

By the end of this guide, you should now have an understanding of the hardware involved in completing the RTSF at Checkout Reference Solution. You should now be confident enough in adding your own hardware devices and begin swapping out components that better suit your target deployment scenario.