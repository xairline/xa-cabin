package flight_status

import "xa-cabin/models"

func (f flightStatusService) processDatarefTakeoff(datarefValues models.DatarefValues) {
	if datarefValues["vs"].GetFloat64() > 200 &&
		datarefValues["gear_force"].GetFloat64() < 1 {
		event := f.AddFlightEvent("Climb", models.StateEvent)
		f.changeState(models.FlightStateClimb, 0.2)
		f.addLocation(datarefValues, -1, &event)
	} else {
		f.addLocation(datarefValues, 0.01, nil)
	}
}
