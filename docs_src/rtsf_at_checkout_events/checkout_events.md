The following are the details for the four different RTSF at Checkout event types that provide the information for a RTSF at Checkout solution. See the [Data Dictionary](#data-dictionary) section below for details about each of the fields in these events.

### POS Events

POS events are what drive the RTSF at Checkout solution. They are the ones that can not be omitted. There are five different POS events required for this reference design, which are:

#### Basket Open
`basket-open` Occurs when a session has started at the self checkout.

Example event:

``` json
   {
		"lane_id" : "1",
		"basket_id": "abc-012345-def",
		"customer_id": "joe5",
		"employee_id": "mary1",
		"event_time" : 15736013010000
   }
```
   
#### Scanned Item 
`scanned-item` Also know as the Real Time Transaction Log (RTTL)
Occurs when an item has been scanned at the self checkout

Example event:

``` json
   {
		"lane_id" : "1",
		"basket_id" : "abc-012345-def",
		"product_id" : "00000000324588",
		"product_id_type" : "UPC",
		"product_name" : "Red Apples",
		"quantity" : 3.0,
		"quantity_unit" : "EA",
		"unit_price" : 0.99,
		"customer_id" : "joe5",
		"employee_id" : "mary1",
		"event_time" : 15736013170000
   }
```

#### Payment Start 
`payment-start` occurs when the payment has started at the self checkout.

Example event:

``` json
   {
		"lane_id" : "1",
		"basket_id" : "abc-012345-def",
		"customer_id" : "joe5",
		"employee_id" : "mary1",
		"event_time" : 15736013660000    
   }
```

   

#### Payment Success 
`payment-success` occurs when the payment has successfully completed at the self checkout.

Example event:

``` json
   {
		"lane_id" : "1",
		"basket_id" : "abc-012345-def",
		"customer_id" : "joe5",
		"employee_id" : "mary1",
		"event_time" : 15736013780000    
   }
```

   

#### Basket Close

`basket-close` occurs when the session has ended at the self checkout.

Example event:

``` json
   {
		"lane_id" : "1",
		"basket_id" : "abc-012345-def",
		"customer_id" : "joe5",
		"employee_id" : "mary1",
		"event_time" : 15736013940000    
   }
```

### Scale Events

Scale events track items on the scale. There is only one scale event type required for this reference design, which is:

#### Scale Weight Reading
`Scale Weight Reading` occurs when an item has been placed or removed from the security scale.

Example event:

``` json
   {
		"lane_id" : "1",
		"scale_id" : "abc123", 
		"total" : 3.25,
		"units" : "lbs",
		"event_time" : 15736013940000    
   }
```

### CV ROI Events

CV ROI events track when objects enter or exit specific ROI . There is only one CV ROI event type required for this reference design, which is:

#### CV ROI Event 
`cv-roi-event` occurs when an object has entered or exited a ROI

Example event:

``` json
   {
		"lane_id" : "1",
		"product_name": "abcde11",
		"roi_action": "ENTERED",
		"roi_name": "Staging",
		"event_time" : 15736014560000    
   }
```

`roi_action` can be either `ENTERED` or `EXITED`.

### RFID ROI Events

RFID events track the RFID tagged products entering and exiting specific ROI. There is only one RFID ROI event type required for this reference design, which is:

#### RFID ROI Event 
`rfid-roi-event` occurs when a RFID tagged object has entered or exited a ROI. 

Example event:
   
``` json
    {
		"lane_id" : "1",
		"epc":"30143639F8419145BEEF0009",
		"roi_name": "Staging",
		"roi_action": "EXITED",
		"event_time" : 15736014790000            
    } 
```
   
   `roi_action` can be either `ENTERED` or `EXITED`.

