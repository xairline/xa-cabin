package flight_status

import (
	"xa-cabin/models"
)

func (f flightStatusService) processDatarefLanding(datarefValues models.DatarefValues) {
	if datarefValues["gs"].GetFloat64() < 30/1.9438 {
		airportId, airportName := f.DatarefSvc.GetNearestAirport()
		f.setArrivalFlightInfo(
			airportId,
			airportName,
			datarefValues["ts"].GetFloat64(),
			0,
			0,
		)
		f.changeState(models.FlightStateTaxiIn, 0.2)
		f.addLocation(datarefValues, -1, nil)
		return
	} else {
		var myEvent *models.FlightStatusEvent
		if f.IsTouchdown() {
			event := f.AddFlightEvent("Touchdown", models.LocationEvent)
			myEvent = &event
		}
		f.addLocation(datarefValues, 0.1, myEvent)
	}
	// go-around
	//if datarefValues["vs"].GetFloat64() > 500 &&
	//	datarefValues["gear_force"].GetFloat64() < 1 {
	//	f.AddFlightEvent(datarefValues, "Climb")
	//	f.changeState(models.FlightStateClimb, 0.2)
	//	return
	//}
}
