### requirements

Develop a parking lot application where we should be able to perform the following operations:
1) Parking manager can create parking lots with desired parking spaces/slots in each parking lot.
2) The user (Vehicle owner) can choose any parking lot & can park his vehicle in the nearest parking slot available in that lot (e.g. if parking slots are numbered 1,2,3....n, then we still start from 1 and pick the available one)
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
