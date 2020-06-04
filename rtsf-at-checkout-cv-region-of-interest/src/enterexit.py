import paho.mqtt.client as paho
import json
import time
import requests 
from threading import Timer
import os

# MQTT related constants
MQTT_BROKER_HOST = "mqtt"
MQTT_BROKER_PORT = 1883
MQTT_KEEPALIVE = 60
MQTT_INCOMING_TOPIC_NAME = "AnalyticsData"
MQTT_OUTBOUND_TOPIC_NAME = "edgex"
EDGEX_DEVICE_NAME = "device-cv-mqtt"
EDGEX_ROI_EVENT = "cv-roi-event"
EDGEX_ENTER_EVENT = 'ENTERED'
EDGEX_EXIT_EVENT = 'EXITED'

MQTT_BROKER_ADDRESS = MQTT_BROKER_HOST + ":" + str(MQTT_BROKER_PORT)

oldFrameDict = {}

def on_connect(client, userdata, message, rc):
    print("Connected to mqtt broker")
    client.subscribe(MQTT_INCOMING_TOPIC_NAME)


def on_subscribe(client, userdata, message, qos):
    print("Subscribed to topic")


def on_message(client, userdata, message):
    newFrameDict = {}

    python_obj = json.loads(message.payload)
    resolution = python_obj["resolution"]
    height = resolution["height"]
    width = resolution["width"]
    source = python_obj["source"]
    roi_name = python_obj["tags"]["roi_name"]
    timestamp = python_obj["timestamp"] # timestamp is milliseconds since start of stream
    
    # Calculate timestamp for reporting
    milliSinceEPOCH = int(round(time.time() * 1000))

    if 'objects' in python_obj:
        # Broken down
        for indv_object_detected in python_obj['objects']:
            detection = indv_object_detected["detection"]
            bounding_box = detection["bounding_box"]
            x_max = bounding_box["x_max"]
            x_min = bounding_box["x_min"]
            y_max = bounding_box["y_max"]
            y_min = bounding_box["y_min"]
            confidence = detection["confidence"]
            label = detection["label"]
            label_id = detection["label_id"]

            #For each frame, add the label or increment it in the dict if it is seen
            if label in newFrameDict:
                newFrameDict[label] = newFrameDict[label] + 1;
            else:
                newFrameDict[label] = 1

        # Enter Exit Logic to be used when tracking is not available
        # This is a simple algorithm that uses counter logic to detect enter exit events
        global oldFrameDict

        # Create a blank dict for comparison for brand new roi_name
        if roi_name not in oldFrameDict:
            oldFrameDict[roi_name] = {}

        for key in newFrameDict:
            # Check to see if this object type was detected in the previous frame
            # and if so, what was the count
            # if the count does not match up with the previous frame, report enters or exits
            if key in oldFrameDict[roi_name]:
                if (newFrameDict[key] > oldFrameDict[roi_name][key]):
                    for i in range(0, (newFrameDict[key] - oldFrameDict[roi_name][key])):                                                
                        newEnterExitElement = {}
                        newEnterExitElement["source"] = source
                        newEnterExitElement["event_time"] = milliSinceEPOCH
                        newEnterExitElement["product_name"] = key
                        newEnterExitElement["roi_action"] = EDGEX_ENTER_EVENT
                        newEnterExitElement["roi_name"] = roi_name
                        mqtt_msg = wrap_edgex_event(EDGEX_DEVICE_NAME, EDGEX_ROI_EVENT, json.dumps(newEnterExitElement))
                        client.publish(MQTT_OUTBOUND_TOPIC_NAME, mqtt_msg)
                elif (newFrameDict[key] < oldFrameDict[roi_name][key]):
                    for i in range(0, (oldFrameDict[roi_name][key] - newFrameDict[key])):
                        newEnterExitElement = {}
                        newEnterExitElement["source"] = source
                        newEnterExitElement["event_time"] = milliSinceEPOCH
                        newEnterExitElement["product_name"] = key
                        newEnterExitElement["roi_action"] = EDGEX_EXIT_EVENT
                        newEnterExitElement["roi_name"] = roi_name
                        mqtt_msg = wrap_edgex_event(EDGEX_DEVICE_NAME, EDGEX_ROI_EVENT, json.dumps(newEnterExitElement))
                        client.publish(MQTT_OUTBOUND_TOPIC_NAME, mqtt_msg)
                del oldFrameDict[roi_name][key]
            else:
                # Report everything in here as new enter since it was not in the prev frame
                for i in range(0, newFrameDict[key]):
                    newEnterExitElement = {}
                    newEnterExitElement["source"] = source
                    newEnterExitElement["event_time"] = milliSinceEPOCH
                    newEnterExitElement["product_name"] = key
                    newEnterExitElement["roi_action"] = EDGEX_ENTER_EVENT
                    newEnterExitElement["roi_name"] = roi_name
                    mqtt_msg = wrap_edgex_event(EDGEX_DEVICE_NAME, EDGEX_ROI_EVENT, json.dumps(newEnterExitElement))
                    client.publish(MQTT_OUTBOUND_TOPIC_NAME, mqtt_msg)

    # Lastly, in case of an object type is completely removed from frame,
    # iterate over the old frame for the remaining types to report them as exited
    for key in oldFrameDict[roi_name]:
        for i in range(0, oldFrameDict[roi_name][key]):
            newEnterExitElement = {}
            newEnterExitElement["source"] = source
            newEnterExitElement["event_time"] = milliSinceEPOCH
            newEnterExitElement["product_name"] = key
            newEnterExitElement["roi_action"] = EDGEX_EXIT_EVENT
            newEnterExitElement["roi_name"] = roi_name
            mqtt_msg = wrap_edgex_event(EDGEX_DEVICE_NAME, EDGEX_ROI_EVENT, json.dumps(newEnterExitElement))
            client.publish(MQTT_OUTBOUND_TOPIC_NAME, mqtt_msg)

    #Replace the old frame data with the new frame data
    oldFrameDict[roi_name] = newFrameDict.copy()

def wrap_edgex_event(device_name, cmd_name, data):
    edgexMQTTWrapper = {}
    edgexMQTTWrapper["name"] = device_name
    edgexMQTTWrapper["cmd"] = cmd_name
    edgexMQTTWrapper[cmd_name] = data
    return json.dumps(edgexMQTTWrapper)

def create_pipelines():
    print("creating video analytics pipelines")    
    
    cameraConfiguration = []
    mqttDestHost = os.environ.get('MQTT_DESTINATION_HOST')
    
    if cameraConfiguration == None:
        print("WARNING: Enter Exit Service could not create video pipeline(s), environment variable MQTT_DESTINATION_HOST not set correctly")
        return
      
    i = 0 
    while True:
        # read env vars to find camera topic and source
        # expecting env vars to be in the form CAMERA0_SRC and CAMERA0_MQTTTOPIC
        camSrc = os.environ.get('CAMERA' + str(i) + '_SRC')
        roiName = os.environ.get('CAMERA' + str(i) + '_ROI_NAME')
        camEndpoint = os.environ.get('CAMERA'+ str(i) +'_ENDPOINT')
        camCropTBLR = str(os.environ.get('CAMERA'+ str(i) +'_CROP_TBLR'))
        camStreamPort = os.environ.get('CAMERA' + str(i) + '_PORT')    
        camCrops = dict(zip(["top", "bottom", "left", "right"], [x for x in camCropTBLR.split(",")]))
        if len(camCrops) < 4:
            camCrops = dict(zip(["top", "bottom", "left", "right"], [0] * 4))
        
        if camStreamPort == None: 
            camStreamPort = 0

        if camSrc == None or roiName == None:
            break # should break out of the loop when no more CAMERA env vars are found

        srcPath, srcType = ('uri', 'uri') if ('rtsp' in camSrc) else ('path', 'string')       
        jsonConfig = {
            'source': {
                srcPath: camSrc,
                'type': srcType
            },
            'destination': {
                "type": "mqtt",
                "address": mqttDestHost,
                "topic": "AnalyticsData",
                "timeout": 1000
            }, 
            'tags': {'roi_name':roiName},
            'parameters' :{
                "top":int(camCrops["top"]),
                "left":int(camCrops["left"]),
                "right":int(camCrops["right"]),
                "bottom":int(camCrops["bottom"]),
                "port":int(camStreamPort)
            },
            'camEndpoint': camEndpoint
        }
        
        cameraConfiguration.append(jsonConfig)
        i += 1 
        
    if len(cameraConfiguration) < 1:
        print("WARNING: Enter Exit Service could not create video pipeline(s), environment variable(s) not set correctly")
        return
    
    for camConfig in cameraConfiguration:
        data = {}
        data['source'] = camConfig['source']
        data['destination'] =  camConfig['destination']   
        data['tags'] =  camConfig['tags']  
        data['parameters'] = camConfig['parameters']       
        jsonData = json.dumps(data)            
        endpoint = camConfig['camEndpoint']  
        print(jsonData)
        headers = {'Content-type': 'application/json'}
        r = requests.post(url = endpoint, data = jsonData, headers = headers) 
        if r.status_code == 200:
            print("Created new pipeline with id: %s"%r.text)
        else:
            print("Error creating pipeline: %s"%r)

# TODO fix this for cam endpoints
# def delete_pipeline(instance):
#     endpoint = os.environ.get('VAS_ENDPOINT')
#     url = endpoint + '/' + instance
#     r = requests.delete(url = url)
#     print("Deleted pipeline: %s"%r.text)


wait_time = 5.0
t = Timer(wait_time, create_pipelines)
t.start()

mqttClient = paho.Client()
mqttClient.on_message = on_message
mqttClient.on_connect = on_connect
mqttClient.on_subscribe = on_subscribe
try:
    mqttClient.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT, MQTT_KEEPALIVE)
    mqttClient.loop_forever()
except:
    print("WARNING: Enter Exit Service could not connect to mqtt broker, no enter exit messages will be produced")

