### requirements

Develop a parking lot application where we should be able to perform the following operations:
1) Parking manager can create parking lots with desired parking spaces/slots in each parking lot.
2) The user (Vehicle owner) can choose any parking lot & can park his vehicle in the nearest parking slot available in that lot 
(e.g. if parking slots are numbered 1,2,3....n, then we still start from 1 and pick the available one).
3) The user can unpark his vehicle.
4) When the user unparks, the response should be successful along with the parking fee that will be calculated as Rs. 10 * the Number of hours the vehicle has been parked. eg If parked for 1 hour 5 minutes, it will be 10 * 2 = 20
5) Parking manager can view his current parking lot status (eg which cars are parked in which slots)
6) The parking manager can put any parking space/slot into maintenance mode and back to working state at any time.
7) The Parking Manager should be able to get the total number of vehicles parked on any day, total parking time and the total fee collected on that day.

Focus: Close the solution is to the requirements, code quality, API design, database design & extensibility in terms of adding new features.


### clarifications

1) Assume only 1 parking manager. Focus on actual parking problem, Users entity is not required.
2) Only cars for the sake of simplicity.
3) Parking lot size is irrelevant.
4) Only on the spot parking is allowed.
5) For simplicity, Payment gateway is not needed. Assume all payments will be success.
6) No waiting queue. But error conditions like Full parking should be handled and appropriate response should be returned.
7) Historical parking data should be there. But we will request data of 1 day only for any particular day in the past.
8) Dashboard or Report is not needed. We are only focussing on backend api response.

### Request/response payload structures (JSON)

1.Create A Parking Lot: POST /parking-lots/:id/slots

Request: 
```
{
    "name": "Gladson Load Parking",
    "desiredSlots": 4
}
```
Response:

```
{
    "id": "bad082a1-1132-4eee-a68f-0469786dd5bb",
    "name": "Gladson Load Parking",
    "desiredSlots": 4,
    "slots": [
        {
            "id": "d8a2351b-0f0c-47cb-87a2-22fb7a1eec07",
            "slotNumber": 1,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "532a76df-eb88-4dc1-a2ea-d44dc70e2129",
            "slotNumber": 2,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "3b665cae-304d-4b16-9886-3608f41e3f30",
            "slotNumber": 3,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "75598cd7-0498-4131-9dc1-c234a206f4a7",
            "slotNumber": 4,
            "isAvailable": true,
            "isMaintenance": false
        }
    ]
}
```

Possible Errors
* Bad Request (400): Missing or invalid parking lot name.
* Internal Server Error (500): Database insertion failure.
* Conflict error (409) : Parking lot with same name already exists.

2.Get Parking Lot Status, GET /parking-lots/:id/slots

Request None (Parking lot ID is part of the URL path)

Response
```
[
{
"id": "8ef6ab22-19aa-45b9-81a8-7e760eeb617c",
"slot_number": 1,
"is_available": true,
"is_maintenance": false
},
{
"id": "302f634a-8171-4069-96b5-365b0b6063af",
"slot_number": 2,
"is_available": false, // Occupied by a vehicle
"is_maintenance": false
},
// ... more slots
]
```

Possible Errors
* Not Found (404): Parking lot with specified ID doesn't exist.
* Internal Server Error (500): Database query failure.

3.Park Vehicle, POST /parking-lots/:id/park

Request
```
{
"registration_number": "AB-1234"
}
```

Response (Success)
```
{
"slot_id": "8ef6ab22-19aa-45b9-81a8-7e760eeb617c",
"parked_at": "2023-11-22T14:51:20Z"
}
```

Possible Errors
* Bad Request (400): Missing or invalid registration_number.
* Not Found (404): Parking lot doesn't exist.
* Conflict (409): The parking lot is full.
* Internal Server Error (500): Database error.

4.Unpark Vehicle, POST /parking-lots/:id/unpark

Request
```
{
"registration_number": "AB-1234"
}
```


Response (Success)
```
{
"slot_id": "8ef6ab22-19aa-45b9-81a8-7e760eeb617c",
"parked_at": "2023-11-22T14:51:20Z",
"unparked_at": "2023-11-22T16:33:58Z",
"fee": 30 // (3 hours * 10)
}
```

Possible Errors
* Bad Request (400): Missing or invalid registration_number.
* Not Found (404): Parking lot or vehicle with the given registration number doesn't exist.
* Internal Server Error (500): Database error or error calculating parking duration.

5.Daily Report, GET /parking-lots/:id/report/:date (e.g., /parking-lots/123/report/2023-11-22)

Request None (Parking lot ID and date are part of the URL path)

Response:

```{
"total_vehicles_parked": 5,
"total_parking_time": 20, // Hours
"total_fee_collected": 200
}
```

Possible Errors
* Not Found (404): Parking lot doesn't exist.
* Bad Request (400): Invalid date format.
* Internal Server Error (500): Database query issues.
