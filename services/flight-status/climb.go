package flight_status

import "xa-cabin/models"

func (f flightStatusService) processDatarefClimb(datarefValues models.DatarefValues) {
	if datarefValues["vs"].GetFloat64() > -500 &&
		datarefValues["vs"].GetFloat64() < 500 {
		*f.cruiseCounter += 1
	} else {
		*f.cruiseCounter = 0
	}
	if datarefValues["vs"].GetFloat64() < -500 {
		*f.descendCounter += 1
	} else {
		*f.descendCounter = 0
	}
	// 15s
	if *f.cruiseCounter >= int(15/f.FlightStatus.PollFrequency) {
		event := f.AddFlightEvent("Cruise", models.StateEvent)
		f.changeState(models.FlightStateCruise, 1)
		f.addLocation(datarefValues, -1, &event)
		return
	}
	if *f.descendCounter >= int(15/f.FlightStatus.PollFrequency) {
		event := f.AddFlightEvent("Descend", models.StateEvent)
		f.changeState(models.FlightStateDescend, 0.2)
		f.addLocation(datarefValues, -1, &event)
	}
	f.addLocation(datarefValues, 10, nil)
}
