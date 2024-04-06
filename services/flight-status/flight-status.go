package flight_status

//go:generate mockgen -destination=../__mocks__/flight-status/flight-status.go -package=mocks -source=flight-status.go

import (
	_ "embed"
	"fmt"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	ruleLogger "github.com/hyperjumptech/grule-rule-engine/logger"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"xa-cabin/models"
	"xa-cabin/services/dataref"
	"xa-cabin/utils/logger"
)

var flightStatusSvcLock = &sync.Mutex{}
var flightStatusSvc FlightStatusService

type FlightStatusService interface {
	ResetFlightStatus()
	GetFlightStatus() *models.FlightStatus
	ProcessDataref(datarefValues models.DatarefValues) float32
	processDatarefParked(datarefValues models.DatarefValues)
	processDatarefTaxiOut(datarefValues models.DatarefValues)
	processDatarefTakeoff(datarefValues models.DatarefValues)
	processDatarefClimb(datarefValues models.DatarefValues)
	processDatarefCruise(datarefValues models.DatarefValues)
	processDatarefDescend(datarefValues models.DatarefValues)
	processDatarefLanding(datarefValues models.DatarefValues)
	processDatarefTaxiIn(datarefValues models.DatarefValues)
	changeState(newState models.FlightState, newPollFrequency float32)
	AddFlightEvent(description string, eventType models.FlightStatusEventType) models.FlightStatusEvent
	setDepartureFlightInfo(airportId, airportName string, timestamp, fuelWeight, totalWeight float64)
	setArrivalFlightInfo(airportId, airportName string, timestamp, fuelWeight, totalWeight float64)
	addLocation(datarefValues models.DatarefValues, distance_threshold float64, event *models.FlightStatusEvent)
	GetLocation() models.FlightStatusLocation
	AddViolationEvent(description string, details interface{})
	EventExists(description string) bool
	IsTouchdown() bool
}

type flightStatusService struct {
	FlightStatus    *models.FlightStatus
	DatarefSvc      dataref.DatarefService
	cruiseCounter   *int
	climbCounter    *int
	descendCounter  *int
	Logger          logger.Logger
	CurrentLocation *models.FlightStatusLocation
	KnowledgeBase   *ast.KnowledgeBase
	Engine          *engine.GruleEngine
	DataCtx         ast.IDataContext
}

func (f flightStatusService) EventExists(description string) bool {
	for _, v := range f.FlightStatus.Events {
		if v.Description == description {
			return true
		}
	}
	return false
}

func (f flightStatusService) IsTouchdown() bool {
	return f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].GearForce < 1 && f.CurrentLocation.GearForce > 1
}

func (f flightStatusService) GetLocation() models.FlightStatusLocation {
	var datarefValues = f.DatarefSvc.GetCurrentValues()
	return models.FlightStatusLocation{
		ID:        0,
		FlightId:  int(f.FlightStatus.ID),
		Timestamp: datarefValues["ts"].GetFloat64(),
		Vs:        datarefValues["vs"].GetFloat64(),
		Ias:       datarefValues["ias"].GetFloat64(),
		Lat:       datarefValues["lat"].GetFloat64(),
		Lng:       datarefValues["lng"].GetFloat64(),
		Altitude:  datarefValues["elevation"].GetFloat64(),
		Agl:       datarefValues["agl"].GetFloat64(),
		GearForce: datarefValues["gear_force"].GetFloat64(),
		GForce:    datarefValues["g_force"].GetFloat64(),
		Fuel:      datarefValues["fuel_weight"].GetFloat64(),
		Heading:   datarefValues["heading"].GetFloat64(),
		State:     f.FlightStatus.CurrentState,
	}
}

func (f flightStatusService) addLocation(datarefValues models.DatarefValues, distance_threshold float64, event *models.FlightStatusEvent) {
	newLocation := models.FlightStatusLocation{
		ID:        0,
		FlightId:  int(f.FlightStatus.ID),
		Timestamp: datarefValues["ts"].GetFloat64(),
		Vs:        datarefValues["vs"].GetFloat64(),
		Ias:       datarefValues["ias"].GetFloat64(),
		Lat:       datarefValues["lat"].GetFloat64(),
		Lng:       datarefValues["lng"].GetFloat64(),
		Altitude:  datarefValues["elevation"].GetFloat64(),
		Agl:       datarefValues["agl"].GetFloat64(),
		GearForce: datarefValues["gear_force"].GetFloat64(),
		GForce:    datarefValues["g_force"].GetFloat64(),
		Fuel:      datarefValues["fuel_weight"].GetFloat64(),
		GS:        datarefValues["gs"].GetFloat64(),
		Pitch:     datarefValues["pitch"].GetFloat64(),
		Heading:   datarefValues["heading"].GetFloat64(),
		State:     f.FlightStatus.CurrentState,
	}
	if event != nil {
		myEvent := *event
		flightStatusSvcLock.Lock()
		f.FlightStatus.Events = append(f.FlightStatus.Events, myEvent)
		flightStatusSvcLock.Unlock()
	}
	if len(f.FlightStatus.Locations) == 0 || (f.FlightStatus.CurrentState == models.FlightStateLanding && newLocation.Agl < 5) {
		f.FlightStatus.Locations = append(
			f.FlightStatus.Locations,
			newLocation,
		)
	} else {
		lastLat := f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].Lat
		lastLng := f.FlightStatus.Locations[len(f.FlightStatus.Locations)-1].Lng
		curLat := datarefValues["lat"].GetFloat64()
		curLng := datarefValues["lng"].GetFloat64()
		delta := distance(lastLat, lastLng, curLat, curLng)
		if delta > distance_threshold {
			f.FlightStatus.Locations = append(
				f.FlightStatus.Locations,
				newLocation,
			)
		}
	}
}

func (f flightStatusService) setDepartureFlightInfo(airportId, airportName string, timestamp, fuelWeight, totalWeight float64) {
	f.FlightStatus.DepartureFlightInfo = models.FlightInfo{
		AirportId:   airportId,
		AirportName: airportName,
		FuelWeight:  fuelWeight,
		TotalWeight: totalWeight,
		Time:        timestamp,
	}
}

func (f flightStatusService) setArrivalFlightInfo(airportId, airportName string, timestamp, fuelWeight, totalWeight float64) {
	f.FlightStatus.ArrivalFlightInfo = models.FlightInfo{
		AirportId:   airportId,
		AirportName: airportName,
		FuelWeight:  fuelWeight,
		TotalWeight: totalWeight,
		Time:        timestamp,
	}
}

func (f flightStatusService) AddFlightEvent(description string, eventType models.FlightStatusEventType) models.FlightStatusEvent {
	event := models.FlightStatusEvent{
		ID:          0,
		FlightId:    int(f.FlightStatus.ID),
		Timestamp:   f.CurrentLocation.Timestamp,
		Description: description,
		EventType:   eventType,
	}
	f.Logger.Infof(
		"NEW Event: %+v",
		event.Description,
	)
	return event
}

func (f flightStatusService) AddViolationEvent(description string, details interface{}) {
	event := models.FlightStatusEvent{
		ID:          0,
		FlightId:    int(f.FlightStatus.ID),
		Timestamp:   f.CurrentLocation.Timestamp,
		Description: description,
		EventType:   models.ViolationEvent,
		Details:     fmt.Sprintf("%s: %v", description, details),
	}
	f.Logger.Infof(
		"NEW Event: %+v",
		event.Description,
	)
	flightStatusSvcLock.Lock()
	f.FlightStatus.Events = append(f.FlightStatus.Events, event)
	flightStatusSvcLock.Unlock()
}

func (f flightStatusService) GetFlightStatus() *models.FlightStatus {
	return f.FlightStatus
}

func (f flightStatusService) ProcessDataref(datarefValues models.DatarefValues) float32 {

	f.CurrentLocation = &models.FlightStatusLocation{
		ID:        0,
		FlightId:  int(f.FlightStatus.ID),
		Timestamp: datarefValues["ts"].GetFloat64(),
		Vs:        datarefValues["vs"].GetFloat64(),
		Ias:       datarefValues["ias"].GetFloat64(),
		Lat:       datarefValues["lat"].GetFloat64(),
		Lng:       datarefValues["lng"].GetFloat64(),
		Altitude:  datarefValues["elevation"].GetFloat64(),
		Agl:       datarefValues["agl"].GetFloat64(),
		GearForce: datarefValues["gear_force"].GetFloat64(),
		GForce:    datarefValues["g_force"].GetFloat64(),
		Fuel:      datarefValues["fuel_weight"].GetFloat64(),
		GS:        datarefValues["gs"].GetFloat64(),
		Pitch:     datarefValues["pitch"].GetFloat64(),
		Heading:   datarefValues["heading"].GetFloat64(),
		FlapRatio: datarefValues["flap_ratio"].GetFloat64(),
		State:     f.FlightStatus.CurrentState,
	}
	err := f.DataCtx.Add("FACT", f)
	if err != nil {
		f.Logger.Errorf("add rule: %v", err)
	}
	err = f.Engine.Execute(f.DataCtx, f.KnowledgeBase)
	if err != nil {
		f.Logger.Errorf("rule: %v", err)
	}
	switch f.FlightStatus.CurrentState {
	case models.FlightStateParked:
		f.processDatarefParked(datarefValues)
	case models.FlightStateTaxiOut:
		f.processDatarefTaxiOut(datarefValues)
	case models.FlightStateTakeoff:
		f.processDatarefTakeoff(datarefValues)
	case models.FlightStateClimb:
		f.processDatarefClimb(datarefValues)
	case models.FlightStateCruise:
		f.processDatarefCruise(datarefValues)
	case models.FlightStateDescend:
		f.processDatarefDescend(datarefValues)
	case models.FlightStateLanding:
		f.processDatarefLanding(datarefValues)
	case models.FlightStateTaxiIn:
		f.processDatarefTaxiIn(datarefValues)
	}
	return f.FlightStatus.PollFrequency
}

func (f flightStatusService) ResetFlightStatus() {
	f.Logger.Warning("====== RESET Flight status ======")

	// cleanup data points
	f.cleanupDataPointsAndStore()

	// reset
	f.FlightStatus.Locations = []models.FlightStatusLocation{}
	f.FlightStatus.Events = []models.FlightStatusEvent{}
	f.FlightStatus.ArrivalFlightInfo = models.FlightInfo{}
	f.FlightStatus.DepartureFlightInfo = models.FlightInfo{}
	f.CurrentLocation = new(models.FlightStatusLocation)
	f.FlightStatus.ID = 0
	f.changeState(models.FlightStateParked, 5)
}

func NewFlightStatusService(datarefSvc dataref.DatarefService, logger logger.Logger) FlightStatusService {
	if flightStatusSvc != nil {
		logger.Info("FlightStatus SVC has been initialized already")
		return flightStatusSvc
	} else {
		logger.Info("FlightStatus SVC: initializing")
		flightStatusSvcLock.Lock()
		defer flightStatusSvcLock.Unlock()

		// rules
		knowledgeLibrary := ast.NewKnowledgeLibrary()
		ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)
		bundle := pkg.NewGITResourceBundle("https://github.com/xairline/xairline-v2.git", "/**/*.grl")
		bundle.RefName = "refs/heads/dev"
		resources := bundle.MustLoad()
		for _, res := range resources {
			logger.Infof("load rules: %v", res.String())
			err := ruleBuilder.BuildRuleFromResource("TutorialRules", "0.0.1", res)
			if err != nil {
				panic(err)
			}
		}
		gruleEngine := engine.NewGruleEngine()
		dataCtx := ast.NewDataContext()
		flightStatus := models.FlightStatus{}
		flightStatusSvc = &flightStatusService{
			FlightStatus:    &flightStatus,
			CurrentLocation: new(models.FlightStatusLocation),
			DatarefSvc:      datarefSvc,
			cruiseCounter:   new(int),
			climbCounter:    new(int),
			descendCounter:  new(int),
			Logger:          logger,
			KnowledgeBase:   knowledgeLibrary.NewKnowledgeBaseInstance("TutorialRules", "0.0.1"),
			Engine:          gruleEngine,
			DataCtx:         dataCtx,
		}
		flightStatusSvc.ResetFlightStatus()
		ruleLogger.SetLogLevel(ruleLogger.Level(logrus.TraceLevel))
		return flightStatusSvc
	}
}

func (f flightStatusService) changeState(newState models.FlightState, newPollFrequency float32) {
	f.FlightStatus.CurrentState = newState
	f.FlightStatus.PollFrequency = newPollFrequency
	*f.cruiseCounter = 0
	*f.climbCounter = 0
	*f.descendCounter = 0
}

func (f flightStatusService) cleanupDataPointsAndStore() {
	if len(f.FlightStatus.Locations) > 0 {
		var indexOfLastLanding, indexOfFirstTaxiIn int
		for index, location := range f.FlightStatus.Locations {
			location := location
			if location.State == models.FlightStateLanding {
				if f.FlightStatus.Locations[index].GearForce <= 1 &&
					f.FlightStatus.Locations[index+1].GearForce > 1 {
					indexOfLastLanding = index
				}
			}
			if location.State == models.FlightStateTaxiIn {
				indexOfFirstTaxiIn = index
				f.CurrentLocation = &location
				event := f.AddFlightEvent(fmt.Sprintf("Taxi in at %s", f.FlightStatus.ArrivalFlightInfo.AirportId), models.StateEvent)
				flightStatusSvcLock.Lock()
				f.FlightStatus.Events = append(f.FlightStatus.Events, event)
				flightStatusSvcLock.Unlock()
				break
			}
		}
		numOfDataPointsBetweenLandingAndTaxi := indexOfFirstTaxiIn - indexOfLastLanding - 100
		numOfDataPointsToKeep := int(float64(numOfDataPointsBetweenLandingAndTaxi) * 0.2)
		gap := numOfDataPointsBetweenLandingAndTaxi / numOfDataPointsToKeep
		locations := f.FlightStatus.Locations[:indexOfLastLanding+1+100]
		for i := 1; i < numOfDataPointsToKeep; i++ {
			if i*gap+indexOfLastLanding+1 < indexOfFirstTaxiIn {
				locations = append(locations, f.FlightStatus.Locations[i*gap+indexOfLastLanding+1+100])
			}
		}
		locations = append(locations, f.FlightStatus.Locations[indexOfFirstTaxiIn:]...) // flush locations to db

		f.Logger.Infof("Original: %v, now: %v", len(f.FlightStatus.Locations), len(locations))

	}
}

func distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515

	// convert to km
	dist = dist * 1.609344

	return dist
}
