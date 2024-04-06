package flight_status

import (
	"math"
	"xa-cabin/models"
)

func (f flightStatusService) processDatarefTaxiOut(datarefValues models.DatarefValues) {
	//todo: fix below
	var n1 float64
	for _, value := range datarefValues["n1"].Value.([]float64) {
		n1 = value
		break
	}
	if (n1 > 75 || f.CurrentLocation.GearForce < 10) &&
		datarefValues["gs"].GetFloat64() > 30/1.9438 {
		event := f.AddFlightEvent("Takeoff", models.StateEvent)

		f.changeState(models.FlightStateTakeoff, 0.1)
		f.addLocation(datarefValues, -1, &event)
	} else {
		currentHeading := datarefValues["heading"].GetFloat64()
		lastHeading := f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].Heading
		if math.Abs(lastHeading-currentHeading) > 10 {
			f.addLocation(datarefValues, -1, nil)
		}
	}
}
