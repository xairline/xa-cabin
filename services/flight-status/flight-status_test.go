package flight_status

import (
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"testing"
	"xa-cabin/models"
	"xa-cabin/services/dataref"
	"xa-cabin/utils/logger"
)

func TestFlightStatusService_Rule(t *testing.T) {
	t.Run("simple rule test", func(tt *testing.T) {
		// rules
		knowledgeLibrary := ast.NewKnowledgeLibrary()
		ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)
		bundle := pkg.NewGITResourceBundle("https://github.com/xairline/xairline-v2.git", "/**/*.grl")
		bundle.RefName = "refs/heads/dev"
		resources := bundle.MustLoad()
		for _, res := range resources {
			err := ruleBuilder.BuildRuleFromResource("TutorialRules", "0.0.1", res)
			if err != nil {
				panic(err)
			}
		}
		myEngine := engine.NewGruleEngine()
		dataCtx := ast.NewDataContext()
		flightStatus := models.FlightStatus{}
		flightStatusSvc := &flightStatusService{
			FlightStatus:    &flightStatus,
			CurrentLocation: new(models.FlightStatusLocation),
			DatarefSvc:      dataref.NewDatarefService(logger.NewGenericLogger()),
			cruiseCounter:   new(int),
			climbCounter:    new(int),
			descendCounter:  new(int),
			Logger:          logger.NewGenericLogger(),
			db:              nil,
			KnowledgeBase:   knowledgeLibrary.NewKnowledgeBaseInstance("TutorialRules", "0.0.1"),
			Engine:          myEngine,
			DataCtx:         dataCtx,
		}
		flightStatusSvc.CurrentLocation = &models.FlightStatusLocation{
			ID:        0,
			FlightId:  1,
			Timestamp: 0.0,
			Vs:        -1000,
			Ias:       0,
			Lat:       0,
			Lng:       0,
			Altitude:  0,
			Agl:       0,
			GearForce: 0,
			GForce:    0,
			Fuel:      0,
			GS:        0,
			Pitch:     0,
			Heading:   0,
			FlapRatio: 0,
			State:     models.FlightStateLanding,
		}
		flightStatusSvc.FlightStatus.AircraftICAO = "C172"
		err := flightStatusSvc.DataCtx.Add("FACT", flightStatusSvc)
		if err != nil {
			flightStatusSvc.Logger.Errorf("rule: %v", err)
		}
		res, err := flightStatusSvc.Engine.FetchMatchingRules(flightStatusSvc.DataCtx, flightStatusSvc.KnowledgeBase)
		if err != nil {
			flightStatusSvc.Logger.Errorf("rule: %+v", err)
		}
		flightStatusSvc.Logger.Infof("%v", res)
	})
}
