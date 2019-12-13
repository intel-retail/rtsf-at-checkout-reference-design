## Event Simulation

#### Event Simulator

This reference design includes the **Event Simulator** utility. This simulator reads a JSON based script that defines the event data and wait times between sending each event, which it sends to the `Edgex REST or MQTT Device service`. 

Please note that before the simulator is run, ensure the EdgeX stack is running first via  `make run` from top folder. 

To supply a customized simulation script, use the -f flag like this example: 

​	`./checkout-lane-simulator –f tests/checkoutEvents.json `

Below is an example script for sending POS and Scale events.  Scripts are provided for various `RTSF at Checkout` scenarios.

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

If having problems or issues on running EdgeX stack, please refer to the latter section [Troubleshooting Guide](#troubleshooting-guide) on some of common issues. 

#### PostMan 

The [PostMan](https://www.getpostman.com/) tool can be used to send simulated events to the EdgeX REST Device service. See above [POSTing to EdgeX REST Device Service](#posting-to-edgex-rest-device-service) section for details on how to POST to the Edgex REST Device service.

#### MQTT.FX

The [MQTT.FX](https://mqttfx.jensd.de/) tool or tools like it can be used to send simulated events to the EdgeX MQTT Device service. See above [Publishing to EdgeX MQTT Device Service](Publishing to EdgeX MQTT Device Service) section for details on how to publish to the Edgex REST Device service.

