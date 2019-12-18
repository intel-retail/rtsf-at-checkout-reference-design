# Instructions for Integrating with uniCenta oPOS

uniCenta is an open source Point Of Sale application that runs on Windows, Linux, and Mac devices.  It uses a MySQL database to store product information; the MySQL database must be configured separately as it is not a part of the uniCenta installer.  In this guide you will learn how to download and install uniCenta, how to upload a product list to uniCenta, and how to install custom scripts to communicate to the REST device service. For any additional uniCenta support see the uniCenta [site](https://unicenta.com/ "uniCenta oPOS") where you can access additional resources like user guides, source code, and scripts.

## Install uniCenta and MySQL

### Prerequisites

- Java JRE 8
- MySQL Server

### Installation

Download the installer for your operating system from the link [here](https://sourceforge.net/projects/unicentaopos/ "uniCenta Installer packages").  Once uniCenta is installed you need to link it to your the MySQL.  

To link uniCenta to MySQL

- Create a uniCenta user for MySQL.
- Open uniCenta, navigate to the Databases tab on the configuration page
- Add the connection information for your uniCenta user on MySQL

uniCenta has a useful [video series on YouTube](https://www.youtube.com/watch?v=URglMDmxwS0&list=PLCCYL2bRmxw1tngbebCFY8fg1uaZc8jVE&index=3 "uniCenta oPOS Video series") on installing and using their product  

### Upload Products List

To upload a product inventory list to unicenta open uniCenta, navigate to 'Products Import' by opening the side navbar and clicking 'Tools' -> 'Products Import', then select the CSV file containing your products, for every field in the 'Products Import' page select the corresponding field in the dropdown menu.  An example products file is included below.

```csv
Reference,Barcode,Name,Buy Price,Sell Price,Tax,Category,Default,Supplier
ABE391,013000006408,Ketchup,$0.00,$1.99,000,ALL,1,Jonah Albert
PKL121,049000050158,Sprite 2L,$0.00,$1.99,000,ALL,2,Benjamin Wooten
DKK435,021200519598,Ocelo Sponges,$0.00,$4.99,000,ALL,3,Garrison Calhoun
SCO091,028400159609,Ruffles,$0.00,$2.99,000,ALL,4,Tanner Hampton
POU311,052000338775,Gatorade,$0.00,$1.99,000,ALL,5,Gil Mayer
KFS842,038000183713,Pringles,$0.00,$2.99,000,ALL,6,Scott Herman
IFG952,048001353565,Mayonnaise,$0.00,$5.99,000,ALL,7,Kermit Bean
OLP122,043000955437,Koolaid Fruit Punch,$0.00,$0.99,000,ALL,7,Ben Ben
IKA931,022000008916,Extra Peppermint Gum,$0.00,$1.99,000,ALL,7,Kermit Bean
QWE211,051700988235,Finish Dishwasher Tablet,$0.00,$12.99,000,ALL,7,Kermit Bean
IKA321,012000163173,Mountain Dew 6 Pack,$0.00,$3.49,000,ALL,7,Ben Ben
YKG351,024000566670,Canned Green Beans,$0.00,$0.89,000,ALL,7,Ben Ben
ABE392,00000000324588,Red Apples,$0.00,$0.99,000,ALL,1,Jonah Albert
PKL125,00000000571111,Trail Mix,$0.00,$5.99,000,ALL,2,Benjamin Wooten
DKK439,00000000884389,Red Wine,$0.00,$10.99,000,ALL,3,Garrison Calhoun
SCO098,00000000735797,Steak,$0.00,$8.99,000,ALL,4,Tanner Hampton
POU312,00000000388771,Cheez It,$0.00,$3.99,000,ALL,5,Gil Mayer
KFS843,00000000830881,Salsa,$0.00,$4.99,000,ALL,6,Scott Herman
IFG953,00000000941969,Quaker Oats,$0.00,$8.99,000,ALL,7,Kermit Bean
```

These products will work with the checkout simulator provided.

## Write Custom uniCenta Scripts

uniCenta provides a framework for running custom code as scripts, these scrips are triggered on uniCenta events.  To build a POS system that integrates with our reference design you can use the following uniCenta events to trigger your custom code to send POS transactions to EdgeX.

| Event             | Happens  | Explanation   | REST device service events                                                                |
|-------------------|----------|--------------------------------------------------------------------------------------------------|--------|
| ticket.addline    | BEFORE   | When a new line item is added Script receives details of the line item being added               | basket-open, scanned-item |
| ticket.removeline | BEFORE   | When a line item is removed from the ticket, the script receives index of the line               | remove-item |
| ticket.total      | AFTER    | The = (equals) button is touched Taxes are calculated and...,BEFORE the Payments dialog is shown | payment-start |
| ticket.close      | AFTER    | The ticket is stored in the database and... AFTER the receipt number is assigned                 | payment-success |

The scripts are written in Java and get triggered by subscribing to specific event.  The linking of the custom script code to uniCenta's events is accomplished by first, adding your java code snippet as a resource under 'navbar' -> 'Maintenance' -> 'Resources'.  Second, you link the new resource by adding this line to the 'ticket.buttons' from the 'Resources' page `<event key="ticket.<EVENT_NAME>" code="script.<SCRIPT_NAME>"/>` this links the ticket.<EVENT_NAME> event to a file named 'script.SCRIPT_NAME', make sure to remove or comment out any other event already linked to 'ticket.<EVENT_NAME>'.  This process can be applied to the other uniCenta events as well.

For example to add a script to send a 'basket-open' and 'scanned-item' events to the REST device service you would subscribe to the the 'ticket.addline' event.  And in the script code you would gather the necessary product information and send the appropriate REST requests for 'basket-open' and 'scanned-item'.

Depending on the event you are subscribing to uniCenta will pass parameters into your function.

| Object     | Type         | Description                                                                  |
|------------|--------------|------------------------------------------------------------------------------|
| ticket     | TICKETINFO   | Contains all the information related to the ticket                           |
| place      | STRING       | Contains the table in restaurant mode                                        |
| taxes      | TAXINFO      | Contains a list of all the taxes                                             |
| taxeslogic | TAXESLOGIC   | Contains a list of methods that calculates the tax lines of a current ticket |
| user       | APPUSER      | Contains the User currently logged in                                        |
| sales      | SCRIPTOBJECT | Contains utility methods of the Sales screen                                 |

The functions for working with a ticketline are the following:

| FUNCTION                           |  RESULT                                                    |
|------------------------------------|------------------------------------------------------------|
| setTicket(String ticket, int line) | Set a ticket’s ticketline                                  |
| getTicket()                        | Return the current ticketline                              |
| setMultiply(double dvalue)         | Set the ticketline’s number of multiplier units to dvalue  |
| getMultiply()                      | Return the ticketline’s number of multiplier units         |
| setPrice(double dvalue)            |  Set the ticketline’s unit SellPrice to dvalue             |
| getPrice()                         | Return the ticketline’s unit SellPrice                     |
|setPriceTax(double dvalue)          | Set the ticketline’s unit SellPrice plus tax to dvalue     |
| getPriceTax()                      | Return the ticketline’s unit SellPrice plus tax            |

Additional information can be found on uniCenta's website, specifically downloads -> user guides -> 'uniCenta oPOS Developer Scripting Guide.pdf' once you register.

In the java code you can import `org.json.simple.JSONObject` to create json objects, and `java.net.HttpURLConnection` to initiate the POST request to the device service.  You can even go as for as reading from the uniCenta database by using the `java.sql.*` namespace.
