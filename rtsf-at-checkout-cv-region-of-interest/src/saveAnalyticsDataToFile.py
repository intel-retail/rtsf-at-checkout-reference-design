import paho.mqtt.client as paho
import json

# MQTT related constants
MQTT_BROKER_HOST = "localhost"
MQTT_BROKER_PORT = 1883
MQTT_KEEPALIVE = 60
MQTT_TOPIC_NAME = "AnalyticsData"
MQTT_TOPIC_ENTER_NAME = "enter_event"
MQTT_TOPIC_EXIT_NAME = "exit_event"
MQTT_BROKER_ADDRESS = MQTT_BROKER_HOST + ":" + str(MQTT_BROKER_PORT)

enterExitCount = 0

def on_connect(client, userdata, message, rc):
    print("Connected to mqtt broker")
    f = open("AnalyticsData.txt", "w")
    f.close()
    client.subscribe(MQTT_TOPIC_NAME)


def on_subscribe(client, userdata, message, qos):
    print("Subscribed to topic")


def on_message(client, userdata, message):
    if message.topic == MQTT_TOPIC_NAME:
        f = open("AnalyticsData.txt", "a")
        f.write(str(message.payload.decode("utf-8")) + "\n")
        f.close()

mqttClient = paho.Client()
mqttClient.on_message = on_message
mqttClient.on_connect = on_connect
mqttClient.on_subscribe = on_subscribe
try:
    mqttClient.connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT, MQTT_KEEPALIVE)
    mqttClient.loop_forever()
except:
    print("WARNING: Enter Exit Service could not connect to mqtt broker, no enter exit messages will be produced")
