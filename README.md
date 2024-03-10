# go-park

### How To Run The App

###### Using Makefile and Docker-compose

To run the application using the Makefile:

1. This app tailored for Go 1.22 and later, Ensure you installed this specific version locally.
2. Open your terminal and navigate to the project's root directory.
3. (Optional) Adjust the environment variables in the Makefile as necessary to fit your setup, otherwise, defaults will be loaded.
4. Ensure the Docker Desktop application is running and Run `docker-compose up`  -> For Postgres dependency.
5. Execute the command: `make run`


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


2.Park Vehicle, POST /parking-lots/:id/park

Request
```
{
"registrationNumber": "AB-1234"
}
```

Response (Success)
```
{
    "id": "ab423c8e-2c57-410f-a6fc-6b0af3641baf",
    "registrationNumber": "AB-1234",
    "slotId": "fbb7afa1-8d64-4c34-9a2f-baef18601646",
    "parkedAt": "2024-03-09T20:36:09.570456Z",
    "unparkedAt": null
}
```

Possible Errors
* Bad Request (400): Missing or invalid registration_number.
* Not Found (404): Parking lot doesn't exist.
* Conflict (409): The parking lot is full.
* Internal Server Error (500): Database error.

3.Unpark Vehicle, POST /parking-lots/:id/unpark

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

4.Get Parking Lot Status, GET /parking-lots/:id/status
Parking manager can view his current parking lot status, which cars are parked in which slots

Request None (Parking lot ID is part of the URL path)

Response
```
{
    "parkingLotId": "e391fda0-4141-46a8-b805-983228ed20a2",
    "name": "Blue Road",
    "slots": [
        {
            "slotId": "f7be7f2e-5e15-4627-8e22-980f54b825a9",
            "registrationNumber": {
                "String": "AB",
                "Valid": true
            },
            "parkedAt": "2024-03-10T16:50:17.535042+06:00"
        },
        {
            "slotId": "9f21ab14-0dbc-4751-b4bf-3d713a05c7bb",
            "registrationNumber": {
                "String": "",
                "Valid": false
            }
        },
        {
            "slotId": "0a4b3b4a-83d8-42bc-b6b4-704ae4feba58",
            "registrationNumber": {
                "String": "",
                "Valid": false
            }
        },
    ]
}
```

Possible Errors
* Not Found (404): Parking lot with specified ID doesn't exist.
* Internal Server Error (500): Database query failure.

5.Daily Report, GET /parking-lots/:id/report/:date (e.g., /parking-lots/123/report/2023-11-22)

Request None (Parking lot ID and date are part of the URL path)
This endpoint expects the date in the YYYY-MM-DD format.

Response:
```
{
    "totalVehiclesParked": 6,
    "TotalParkingHours": 3,
    "totalFeeCollected": 30
}
```

Possible Errors
* Not Found (404): Parking lot doesn't exist.
* Bad Request (400): Invalid date format.
* Internal Server Error (500): Database query issues.


<p align="right"><a href="#go-park">↑ Top</a></p>
