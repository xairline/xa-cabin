package main

import (
	"github.com/xairline/goplane/extra/logging"
	"github.com/xairline/goplane/xplm/plugins"
	"github.com/xairline/goplane/xplm/utilities"
	"path/filepath"
	"xa-cabin/services"
	"xa-cabin/services/dataref"
	"xa-cabin/services/flight-status"
	"xa-cabin/utils/logger"
)

// @BasePath  /apis

func main() {
}

func init() {
	logger := logger.NewXplaneLogger()
	plugins.EnableFeature("XPLM_USE_NATIVE_PATHS", true)
	logging.MinLevel = logging.Info_Level
	logging.PluginName = "XA Cabin"
	// get plugin path
	systemPath := utilities.GetSystemPath()
	pluginPath := filepath.Join(systemPath, "Resources", "plugins", "XA Cabin")
	logger.Infof("Plugin path: %s", pluginPath)

	datarefSvc := dataref.NewDatarefService(logger)
	flightStatusSvc := flight_status.NewFlightStatusService(
		datarefSvc,
		logger,
	)
	// entrypoint
	services.NewXplaneService(
		datarefSvc,
		flightStatusSvc,
		logger,
	)
}
