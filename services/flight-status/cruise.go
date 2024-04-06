package flight_status

import (
	"math"
	"xa-cabin/models"
)

func (f flightStatusService) processDatarefCruise(datarefValues models.DatarefValues) {
	if datarefValues["vs"].GetFloat64() > 500 {
		*f.climbCounter += 1
	} else {
		*f.climbCounter = 0
	}
	if datarefValues["vs"].GetFloat64() < -500 {
		*f.descendCounter += 1
	} else {
		*f.descendCounter = 0
	}
	// 30s
	if *f.climbCounter >= int(30/f.FlightStatus.PollFrequency) {
		event := f.AddFlightEvent("Climb", models.StateEvent)

		f.changeState(models.FlightStateClimb, 0.2)
		f.addLocation(datarefValues, -1, &event)
	}
	if *f.descendCounter >= int(30/f.FlightStatus.PollFrequency) {
		event := f.AddFlightEvent("Descend", models.StateEvent)

		f.changeState(models.FlightStateDescend, 0.2)
		f.addLocation(datarefValues, -1, &event)
	}

	currentHeading := datarefValues["heading"].GetFloat64()
	lastHeading := f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].Heading
	if math.Abs(lastHeading-currentHeading) > 10 {
		f.addLocation(datarefValues, -1, nil)
	} else {
		f.addLocation(datarefValues, 50, nil)
	}
}
