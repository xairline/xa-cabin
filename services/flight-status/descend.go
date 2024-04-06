package flight_status

import (
	"math"
	"xa-cabin/models"
)

func (f flightStatusService) processDatarefDescend(datarefValues models.DatarefValues) {
	if datarefValues["agl"].GetFloat64() < 30 &&
		datarefValues["vs"].GetFloat64() < -200 &&
		datarefValues["gear_force"].GetFloat64() < 5 {
		event := f.AddFlightEvent("Landing", models.StateEvent)

		f.changeState(models.FlightStateLanding, -1)
		f.addLocation(datarefValues, -1, &event)
		return
	}

	// Climb again
	if datarefValues["vs"].GetFloat64() > 500 {
		*f.climbCounter += 1
	} else {
		*f.climbCounter = 0
	}

	// Cruise again
	if datarefValues["vs"].GetFloat64() > -500 &&
		datarefValues["vs"].GetFloat64() < 500 &&
		datarefValues["elevation"].GetFloat64() > 3000 {
		*f.cruiseCounter += 1
	} else {
		*f.cruiseCounter = 0
	}
	// 15s
	if *f.cruiseCounter >= int(15/f.FlightStatus.PollFrequency) {
		event := f.AddFlightEvent("Cruise", models.StateEvent)

		f.changeState(models.FlightStateCruise, 1)
		f.addLocation(datarefValues, -1, &event)
		return
	}
	if *f.climbCounter >= int(15/f.FlightStatus.PollFrequency) {
		event := f.AddFlightEvent("Climb", models.StateEvent)
		f.changeState(models.FlightStateClimb, 0.2)
		f.addLocation(datarefValues, -1, &event)
	}

	currentHeading := datarefValues["heading"].GetFloat64()
	lastHeading := f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].Heading
	if math.Abs(lastHeading-currentHeading) > 10 {
		f.addLocation(datarefValues, -1, nil)
	} else {
		f.addLocation(datarefValues, 10, nil)
	}
}
