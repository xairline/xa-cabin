package services

//go:generate mockgen -destination=./__mocks__/xplane.go -package=mocks -source=xplane.go

import (
	"encoding/json"
	"fmt"
	"github.com/xairline/goplane/extra"
	"github.com/xairline/goplane/xplm/processing"
	"github.com/xairline/goplane/xplm/utilities"
	"gorm.io/gorm"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"xa-cabin/services/dataref"
	"xa-cabin/services/flight-status"
	"xa-cabin/utils/logger"
)

type XplaneService interface {
	// init
	onPluginStateChanged(state extra.PluginState, plugin *extra.XPlanePlugin)
	onPluginStart()
	onPluginStop()
	// flight loop
	flightLoop(elapsedSinceLastCall, elapsedTimeSinceLastFlightLoop float32, counter int, ref interface{}) float32
}

type xplaneService struct {
	Plugin              *extra.XPlanePlugin
	DatarefSvc          dataref.DatarefService
	FlightStatusService flight_status.FlightStatusService
	Logger              logger.Logger
	db                  *gorm.DB
}

var xplaneSvcLock = &sync.Mutex{}
var xplaneSvc XplaneService

var commands = []string{}

func NewXplaneService(
	datarefSvc dataref.DatarefService,
	flightStatusSvc flight_status.FlightStatusService,
	logger logger.Logger,
) XplaneService {
	if xplaneSvc != nil {
		logger.Info("Xplane SVC has been initialized already")
		return xplaneSvc
	} else {
		logger.Info("Xplane SVC: initializing")
		xplaneSvcLock.Lock()
		defer xplaneSvcLock.Unlock()
		xplaneSvc := xplaneService{
			Plugin:              extra.NewPlugin("X Web Stack", "com.github.xairline.xwebstack", "A plugin enables Frontend developer to contribute to xplane"),
			DatarefSvc:          datarefSvc,
			FlightStatusService: flightStatusSvc,
			Logger:              logger,
		}
		xplaneSvc.Plugin.SetPluginStateCallback(xplaneSvc.onPluginStateChanged)
		return xplaneSvc
	}
}

func (s xplaneService) onPluginStateChanged(state extra.PluginState, plugin *extra.XPlanePlugin) {
	switch state {
	case extra.PluginStart:
		s.onPluginStart()
	case extra.PluginStop:
		s.onPluginStop()
	case extra.PluginEnable:
		s.Logger.Infof("Plugin: %s enabled", plugin.GetName())
	case extra.PluginDisable:
		s.Logger.Infof("Plugin: %s disabled", plugin.GetName())
	}
}

func (s xplaneService) onPluginStart() {
	s.Logger.Info("Plugin started")
	processing.RegisterFlightLoopCallback(s.flightLoop, -1, nil)
}

func (s xplaneService) onPluginStop() {
	s.Logger.Info("Plugin stopped")
}

func (s xplaneService) flightLoop(elapsedSinceLastCall, elapsedTimeSinceLastFlightLoop float32, counter int, ref interface{}) float32 {
	if len(commands) != 0 {
		command := commands[len(commands)-1]
		commands = commands[:len(commands)-1]
		cmdRef := utilities.FindCommand(command)
		utilities.CommandOnce(cmdRef)
		s.Logger.Infof("Command: %+v executed", cmdRef)
	}
	datarefValues := s.DatarefSvc.GetCurrentValues()
	return s.FlightStatusService.ProcessDataref(datarefValues)
}

func (s xplaneService) getAirportInfoFromICAO(icao string) (map[string]interface{}, error) {
	data := url.Values{
		"icao":    {icao},
		"country": {"ALL"},
		"db":      {"airports"},
		"action":  {"search"},
	}

	resp, err := http.PostForm("https://openflights.org/php/apsearch.php", data)

	if err != nil {
		return map[string]interface{}{}, err
	}

	var res map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	if res["airports"] == nil {
		return map[string]interface{}{}, fmt.Errorf("%s", "Failed to find airport info")
	}

	airport := res["airports"].([]interface{})[0]
	s.Logger.Infof("%+v", airport)
	lat, err := strconv.ParseFloat(airport.(map[string]interface{})["y"].(string), 64)
	if err != nil {
		return map[string]interface{}{}, err
	}
	lng, err := strconv.ParseFloat(airport.(map[string]interface{})["x"].(string), 64)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return map[string]interface{}{
		"AirportName": airport.(map[string]interface{})["name"],
		"Lat":         lat,
		"Lng":         lng,
	}, nil
}
