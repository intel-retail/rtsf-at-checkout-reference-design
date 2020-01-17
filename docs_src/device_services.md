The EdgeX REST and MQTT device services allow for an easy point of integration with this reference design. The above events can be sent to the underlying EdgeX framework using either of these device services. Note that the **Intel Retail Sensor Platform (RSP) **RFID solution has it's own custom EdgeX Device Service. 

## EdgeX REST Device service

This reference design has configured the EdgeX REST device service to have the following devices which accept the above events.  

`device-pos-rest` - Accepts the above POS events and defines the following device resources for each POS Event:

- `basket-open`
- `scanned-item`
- `payment-start`
- `payment-success`
- `basket-close`

`device-scale-rest` - Accepts the above Scale events and defines the following device resource for the single Scale Event:

- `scale-item`

`device-cv-roi-rest` - Accepts the above CV ROI events and defines the following device resource for the single CV ROI Event:

- `cv-roi-event`

`device-rfid-roi-rest`- Accepts the above RFID ROI events and defines the following device resource for the single RFID ROI Event:

- `rfid-roi-event`

### POSTing to EdgeX REST Device Service

This EdgeX REST Device service defines a parametrized endpoint for POSTing the JSON data for the event. This endpoint has the form:
``` text
/resource/{device name}/{resource name}
```

where `{device name}` is one of the above defined devices 
and `{resource name}` is one of the defined resource for that device.

Example URL for POSTing `basket-open` JSON data:

[https://localhost:59990/device-pos-rest/basket-open](https://localhost:59990/device-pos-rest/basket-open)
where the JSON body is:

``` json
{
	"lane_id" : "1",
	"basket_id": "abc-012345-def",
	"customer_id": "joe5",
	"employee_id": "mary1",
    "event_time" : 15736013010000
}
```

## EdgeX MQTT Device service

This reference design has configured the EdgeX MQTT device service to have the following devices which accept the above events.

`device-pos-mqtt` - Accepts the above POS events and defines the following device commands for each POS Event:

- `basket-open`
- `scanned-item`
- `payment-start`
- `payment-success`
- `basket-close`

`device-scale-mqtt` - Accepts the above Scale events and defines the following device command for the single Scale Event:

- `scale-item`

`device-cv-roi-mqtt` - Accepts the above ROI events and defines the following device command for the single CV ROI Event:

- `cv-roi-event`

`device-rfid-roi-mqtt` Accepts the above RFID events and defines the following device command for the single RFID ROI Event:

- `rfid-roi-event`

### Publishing to EdgeX MQTT Device Service

This EdgeX MQTT Device service accepts events published to the `edgex/#` topic. Events published to this topic must conform to the follow JSON schema:

``` json
{
	"name" : "<device name>",
    "cmd" : "<command name>",
    "<command name>" : "<event data json>"
}
```

where `<device name>` is one of the above defined devices 
and `<command name>` is one of the defined commands for that device
and `<event data json>` is a string containing the event data JSON, i.e not embedded JSON.

Example JSON for publishing `scale-item` data:

``` json
{
	"name" : "device-scale-mqtt",
    "cmd" : "scale-item",
    "scale-item" : "{\"lane_id\":\"1\",\"ScaleId\":\"abc123\",\"total\":3.25, 		                          \"delta\":1.15,\"units\":\"lbs\",\"event_time\":15736013940000}"
}
```
