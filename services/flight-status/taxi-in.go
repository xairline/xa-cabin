package flight_status

import (
	"github.com/gorilla/websocket"
	"math"
	"os"
	"time"
	"xa-cabin/models"
)

func (f flightStatusService) processDatarefTaxiIn(datarefValues models.DatarefValues) {
	//todo: fix below
	var n1 float64
	for _, value := range datarefValues["n1"].Value.([]float64) {
		n1 = value
		break
	}
	if datarefValues["gs"].GetFloat64() < 1/1.9438 &&
		n1 < 3 {
		var weightPrecision int8 = 1
		fuelWeight := f.DatarefSvc.GetValueByDatarefName("sim/flightmodel/weight/m_fuel_total", "fuel_weight", &weightPrecision, false)
		totalWeight := f.DatarefSvc.GetValueByDatarefName("sim/flightmodel/weight/m_total", "total_weight", &weightPrecision, false)
		f.FlightStatus.ArrivalFlightInfo.FuelWeight = fuelWeight.GetFloat64()
		f.FlightStatus.ArrivalFlightInfo.TotalWeight = totalWeight.GetFloat64()
		f.addLocation(datarefValues, -1, nil)
		f.ResetFlightStatus()
	} else {
		currentHeading := datarefValues["heading"].GetFloat64()
		lastHeading := f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].Heading
		if math.Abs(lastHeading-currentHeading) > 10 {
			f.addLocation(datarefValues, -1, nil)
		}
	}
}

func (f flightStatusService) hack_sync_flight() {
	token := os.Getenv("CLIENT_TOKEN")
	f.Logger.Infof("hack - CLIENT_TOKEN: %s", token)
	if token == "" {
		f.Logger.Errorf("CLIENT_TOKEN is empty")
		return
	}

	for {
		// Use the token to connect to the WebSocket endpoint
		wsUrl := "wss://app.xairline.org/apis/ws?auth=" + token
		_, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
		if err != nil {
			f.Logger.Errorf("Could not open a ws connection on %s %v", wsUrl, err)
			time.Sleep(time.Second * 5) // Wait for 5 seconds before trying to reconnect
			continue
		}
		break
	}
}
