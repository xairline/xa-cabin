package flight_status

import (
	"fmt"
	"xa-cabin/models"
	"xa-cabin/services/dataref"
)

func (f flightStatusService) processDatarefParked(datarefValues models.DatarefValues) {
	if datarefValues["gs"].GetFloat64() > 1 {
		// get the nearest airport
		airportId, airportName := f.DatarefSvc.GetNearestAirport()
		// get weight information
		var weightPrecision int8 = 1
		fuelWeight := f.DatarefSvc.GetValueByDatarefName("sim/flightmodel/weight/m_fuel_total", "fuel_weight", &weightPrecision, false)
		totalWeight := f.DatarefSvc.GetValueByDatarefName("sim/flightmodel/weight/m_total", "total_weight", &weightPrecision, false)
		f.setDepartureFlightInfo(
			airportId,
			airportName,
			datarefValues["ts"].GetFloat64(),
			fuelWeight.GetFloat64(),
			totalWeight.GetFloat64(),
		)
		// get aircraft name
		aircraftICAO := f.DatarefSvc.GetValueByDatarefName("sim/aircraft/view/acf_ICAO", "icao", nil, true)
		aircraftDisplayName := f.DatarefSvc.GetValueByDatarefName("sim/aircraft/view/acf_ui_name", "acf_ui_name", nil, true)
		f.FlightStatus.AircraftICAO = aircraftICAO.Value.(string)
		f.FlightStatus.AircraftDisplayName = aircraftDisplayName.Value.(string)
		f.Logger.Infof("Departure information: %+v", f.FlightStatus)

		event := f.AddFlightEvent(fmt.Sprintf("Taxi out at %s", airportId), models.StateEvent)
		f.changeState(models.FlightStateTaxiOut, 0.2)
		f.addLocation(datarefValues, -1, &event)
		datarefExtList := dataref.InitializeDatarefList(f.Logger)
		f.DatarefSvc.SetDatarefExtList(&datarefExtList)
	}
}
