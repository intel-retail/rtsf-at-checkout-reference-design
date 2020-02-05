There are three different ways to simulate data in this reference design.

- [Event Simulator](#event-simulator)
- [Postman](#postman)
- [MQTT.fx](#mqtt.fx)


### Event Simulator

This reference design includes the **Event Simulator** utility which reads a JSON-based script that defines the event data and wait times between sending each event. Events are sent to the `EdgeX REST or MQTT Device service`. 


#### Troubleshooting
If you have problems running the EdgeX stack, see [Troubleshooting Guide](#troubleshooting-guide).

#### Getting Started

- Open the docker logs in a terminal window. This lets you make sure checkout events are processed correctly. To open the docker logs:
```
docker logs -f event-reconciler
```
- Make sure the EdgeX stack is running. To do so, from the top folder, run:
```
make run-base
``` 

Optional: To supply a customized simulation script, use the -f flag as in this example: 
```
./event-simulator –f tests/checkoutEvents.json
``` 
   
#### Example Script

The script below provides and example to send POS and Scale events. Scripts are provided for various `RTSF at Checkout` scenarios.

> Note: The simulator adds the `event_time` field to each event so that the values are dynamic.

``` json
{
    "checkout_events": [{
            "device": "device-pos",
            "resource": "basket-open",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "2s"
        },
        {
            "device": "Scale",
            "resource": "scale-item",
            "data": {
                "lane_id" : "1",
                "scale_id" : "abc123",
                "total": 1,
                "units": "lbs"
            },
            "wait_time": "1s"
        },
        {
            "device": "device-pos",
            "resource": "scanned-item",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "product_id": "00000000571111",
                "product_id_type": "UPC",
                "product_name": "Trail Mix",
                "quantity": 1.0,
                "quantity_unit": "EA",
                "unit_price": 5.99,
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "2s"
        },
        {
            "device": "Scale",
            "resource": "scale-item",
            "data": {
                "lane_id" : "1",
                "scale_id" : "abc123",
                "total": 3.0,
                "units": "lbs"
            },
            "wait_time": "1s"
        },
        {
            "device": "device-pos",
            "resource": "scanned-item",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "product_id": "00000000884389",
                "product_id_type": "UPC",
                "product_name": "Red Wine",
                "quantity": 1.0,
                "quantity_unit": "EA",
                "unit_price": 10.99,
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "1s"
        },
        {
            "device": "Scale",
            "resource": "scale-item",
            "data": {
                "lane_id" : "1",
                "scale_id" : "abc123",
                "total": 6.0,
                "units": "lbs"
            },
            "wait_time": "1s"
        },
        {
            "device": "device-pos",
            "resource": "scanned-item",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "product_id": "00000000735797",
                "product_id_type": "UPC",
                "product_name": "Steak",
                "quantity": 1.0,
                "quantity_unit": "EA",
                "unit_price": 8.99,
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "2s"
        },
        {
            "device": "Scale",
            "resource": "scale-item",
            "data": {
                "lane_id" : "1",
                "scale_id" : "abc123",
                "total": 8.11,
                "units": "lbs"
            },
            "wait_time": "1s"
        },
        {
            "device": "device-pos",
            "resource": "payment-start",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "4s"
        },
        {
            "device": "device-pos",
            "resource": "payment-success",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "4s"
        },
        {
            "device": "device-pos",
            "resource": "basket-close",
            "data": {
                "lane_id" : "1",
                "basket_id": "abc-012345-def",
                "customer_id": "joe5",
                "employee_id": "mary1"
            },
            "wait_time": "2s"
        }
    ]
}
```

### Postman 

You can use the [Postman](https://www.getpostman.com/) tool to send simulated events to the EdgeX REST Device Service. See [POSTing to EdgeX REST Device Service](../device_services.md#posting-to-edgex-rest-device-service) for information.

### MQTT.FX

You can use the [MQTT.FX](https://mqttfx.jensd.de/) tool or a similar tool to send simulated events to the EdgeX MQTT Device Service. See [Publishing to EdgeX MQTT Device Service](../device_services.md#publishing-to-edgex-mqtt-device-service) for information.



