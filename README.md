# go-park

### How To Run The App

###### Using Makefile and Docker-compose

To run the application using the Makefile:

1. This app tailored for Go 1.22 and later, Ensure you installed this specific version locally.
2. Open your terminal and navigate to the project's root directory.
3. (Optional) Adjust the environment variables in the Makefile as necessary to fit your setup, otherwise, defaults will be loaded.
4. Ensure the Docker Desktop application is running and Run `docker-compose up`  -> For Postgres dependency.
5. Execute the command: `make run`

#### Project Structure (Domain-driven Design)

```plaintext
├── .github 
│   └── workflows
│       └── go-ci.yaml                    ← GitHub Actions CI workflows (Build, Test, Lint).
├── internal
│   └── domain
│       ├── helpers.go                    ← Helper repository methods(eg: get id from uuid).
│       ├── parking_lot.go                ← Parking lot domain models (ParkingLot, Slot, DailyReport)
│       ├── parking_lot_repository.go     ← Parking interface and it's interactions to postgres database.
│       └── vehicle.go                    ← Vehicle domain model.
│       ├── vehicle_repository.go         ← Vehicle interface and it's interactions to postgres database..
│   └── transport
│       ├── helpers.go                    ← Helper repository methods for http handlers (eg: writeResponse/json binding)
│       ├── parking_lot_handlers.go       ← Parking lot http handlers for net/http.
│       ├── vehicle_handlers.go           ← Vehicle http handlers for net/http.
│   └── common
│       ├── app_errs.go                   ← Mnaging errors, hiding sensitive err from client and send correct http status codes.
│       ├── custom_err.go                 ← Custom error strings fdr the client. (eg: unexpected database error).
│       ├── slog_config.go                ← Structured log with slog config.
│   └── infra
│       └── postgres
│           ├── postgres_conn.go          ← Pgx driver for postgres and db connection string parsing.
│           ├── postgres_conn_test.go     ← Test for database connections.
│       └── docker
│         └── initdb
│             ├── 01.create-database.sql  ← Full gopark database schema in docker entrypoint.
│             ├── 02.generate-data.sql    ← Seed data gopark database schema in docker entrypoint.
├── .gitignore                            ← Specifies intentionally untracked files to ignore.
├── .golangci.yaml                        ← Configuration for golangci-lint.
├── docker-compose.yaml                   ← Docker service setup for development environments.
├── Dockerfile                            ← Dockerfile for building the application image.
├── go.mod                                ← Go module dependencies.
├── main.go                               ← Entry point to start the application services.
├── Makefile                              ← Make command alliases for building and running the application.
└── readme.md                             ← Project documentation and setup instructions.
```

### Request/response payload structures (JSON)

[POSTMAN WORKSPACE LINK](https://www.postman.com/altimetry-cosmonaut-1609324/workspace/go-park)

1.Create A Parking Lot: POST /parking-lots/:id/slots

Request:
```
{
    "name": "Parking Lot 1",
    "desiredSlots": 5
}
```
Response:

```
{
    "id": "6d1a1cd3-fa1d-4596-9fc9-1f4cdb4a739c",
    "name": "Parking Lot 1",
    "desiredSlots": 5,
    "slots": [
        {
            "id": "3f17f943-06d5-4502-9cf0-5e5fa950e04d",
            "slotNumber": 1,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "dde4378a-1408-46e6-875e-450e174d0394",
            "slotNumber": 2,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "c615b97b-74ef-45c7-b2f4-e6c52e8f8e50",
            "slotNumber": 3,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "29dc8edb-ff0d-4da6-bf91-3fed56a6cd7a",
            "slotNumber": 4,
            "isAvailable": true,
            "isMaintenance": false
        },
        {
            "id": "be9e5e81-b0a1-45a3-9621-85bd19054ff6",
            "slotNumber": 5,
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
"registrationNumber": "ABC-123"
}
```

Response (Success)
```
{
    "id": "25bd957a-14ad-40c5-9534-2d158909ef4a",
    "registrationNumber": "ABC-123",
    "slotId": "3f17f943-06d5-4502-9cf0-5e5fa950e04d",
    "parkedAt": "2024-03-12T10:47:27.076353Z"
}

```
In Logs
```
time=2024-03-13T17:46:24.887+06:00 level=INFO source=vehicle_repository.go:132 msg="Chosen nearest slot available" "slot number"=1 vehicle=ABC-123
time=2024-03-13T17:46:33.097+06:00 level=INFO source=vehicle_repository.go:132 msg="Chosen nearest slot available" "slot number"=2 vehicle=ABC-124

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
    "registrationNumber": "ABC-ee767"
}

```


Response (Success)
```
{
    "id": "905f92c9-a4ce-4e2e-a70c-28ac85a255ec",
    "registrationNumber": "ABC-ee767",
    "slotId": "3a2e6c01-a84c-44e3-928e-464370f426be",
    "parkedAt": "2024-03-12T12:18:54.619432+06:00",
    "unparkedAt": "2024-03-12T18:33:18.827961+06:00",
    "fee": 70
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
    "parkingLotId": "9a789f4d-314c-4a95-98c6-00330f9e7f0f",
    "name": "Parking Lot 1",
    "slots": [
        {
            "slotId": "de6aa47b-e797-46da-b2ec-cfbf7e40a107",
            "registrationNumber": "ABC-4fdbb",
            "parkedAt": "2024-03-12T09:58:54.619432+06:00",
            "unparkedAt": null
        },
        {
            "slotId": "446af2fd-05b0-4364-96a0-a18b48eba5cb",
            "registrationNumber": "ABC-71140",
            "parkedAt": "2024-03-12T09:38:54.619432+06:00",
            "unparkedAt": null
        },
        {
            "slotId": "dd8f305a-1bf2-4e4a-bf14-34e11dcdddd1",
            "registrationNumber": "ABC-a45db",
            "parkedAt": "2024-03-12T08:18:54.619432+06:00",
            "unparkedAt": null
        },
        {
            "slotId": "37965e92-5815-4899-8e1c-42c63b1852cb",
            "registrationNumber": "ABC-aa7bc",
            "parkedAt": "2024-03-12T06:58:54.619432+06:00",
            "unparkedAt": null
        },
        {
            "slotId": "d5888249-d67d-401c-a838-0e265be12a4f",
            "registrationNumber": "ABC-ecd50",
            "parkedAt": "2024-03-12T12:28:54.619432+06:00",
            "unparkedAt": "2024-03-12T18:28:54.620855+06:00"
        },
        {
            "slotId": "607fd861-65b7-409d-b064-b9446e28cca1",
            "registrationNumber": "ABC-6ce42",
            "parkedAt": "2024-03-12T11:08:54.619432+06:00",
            "unparkedAt": "2024-03-12T18:28:54.620855+06:00"
        },
        {
            "slotId": "a8c942ce-aa53-4831-8065-16931e33fb2a",
            "registrationNumber": "ABC-a8454",
            "parkedAt": "2024-03-12T09:48:54.619432+06:00",
            "unparkedAt": "2024-03-12T18:28:54.620855+06:00"
        },
        {
            "slotId": "c2e14cae-6d08-4d03-8684-6ef1a1766cad",
            "registrationNumber": "ABC-57770",
            "parkedAt": "2024-03-12T09:28:54.619432+06:00",
            "unparkedAt": "2024-03-12T18:28:54.620855+06:00"
        },
        {
            "slotId": "1cbe8b9b-2ee0-4ddd-a33c-c59cf4b45f47",
            "registrationNumber": "ABC-72312",
            "parkedAt": "2024-03-12T07:08:54.619432+06:00",
            "unparkedAt": "2024-03-12T18:28:54.620855+06:00"
        },
        {
            "slotId": "3a2e6c01-a84c-44e3-928e-464370f426be",
            "registrationNumber": "ABC-ee767",
            "parkedAt": "2024-03-12T12:18:54.619432+06:00",
            "unparkedAt": "2024-03-12T18:33:18.827961+06:00"
        }
    ]
}
```

Possible Errors
* Not Found (404): Parking lot with specified ID doesn't exist.
* Internal Server Error (500): Database query failure.

5.Daily Report, GET /parking-lots/:id/report/:date (e.g., /parking-lots/123/report/2023-11-22)

Request None (Parking lot ID and date are part of the URL path)
This endpoint expects the date in the YYYY-MM-DD, -> 4 digit year, 2 digit month, 2 digit day.

Response:
```
{
    "totalVehiclesParked": 10,
    "totalParkingHours": 53,
    "totalFeeCollected": 530
}

```

Possible Errors
* Not Found (404): Parking lot doesn't exist.
* Bad Request (400): Invalid date format.
* Internal Server Error (500): Database query issues.


<p align="right"><a href="#go-park">↑ Top</a></p>
