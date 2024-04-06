package flight_status

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"xa-cabin/models"
	"xa-cabin/utils"
	"xa-cabin/utils/logger"
)

func TestFlightStatusService_ProcessDatarefClimb(t *testing.T) {
	tmpDbDir := path.Join("/tmp", uuid.New().String())
	os.Mkdir(tmpDbDir, 0700)
	db, _ := utils.CreateDatabase(logger.NewGenericLogger(), tmpDbDir, false)
	var tests = []struct {
		name          string
		initSvc       FlightStatusService
		datarefValues []models.DatarefValues
		expect        models.FlightState
	}{
		{
			name: "it should go from climb to cruise",
			initSvc: flightStatusService{
				FlightStatus: &models.FlightStatus{
					CurrentState:  models.FlightStateTakeoff,
					PollFrequency: 3,
				},
				cruiseCounter:  new(int),
				climbCounter:   new(int),
				descendCounter: new(int),
				Logger:         logger.NewGenericLogger(),
				db:             db,
			},
			datarefValues: []models.DatarefValues{
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 700.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
				map[string]models.DatarefValue{
					"vs": {
						Value: 400.0,
					},
					"ts": {
						Value: 1.0,
					},
				},
			},
			expect: models.FlightStateCruise,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, datarefValues := range tt.datarefValues {
				tt.initSvc.processDatarefClimb(datarefValues)
			}
			assert.Equal(t, tt.initSvc.GetFlightStatus().CurrentState, tt.expect)
		})
	}
}
