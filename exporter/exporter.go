package exporter

import (
	"fmt"
	"sync"

	"adtennant.dev/octopus-consumption-influxdb/config"
	"github.com/FileGo/octopusenergyapi"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
)

type Exporter struct {
	meterPoints *config.MeterPoints

	client *octopusenergyapi.Client
	db     influxdb.Client

	logger *logrus.Logger
}

func New(meterPoints *config.MeterPoints, client *octopusenergyapi.Client, db influxdb.Client, logger *logrus.Logger) *Exporter {
	return &Exporter{
		meterPoints,

		client,
		db,

		logger,
	}
}

func (e *Exporter) Export(database string) error {
	bp, _ := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})

	points, err := e.getElectricityConsumption()
	if err != nil {
		return fmt.Errorf("failed to get electricity consumption: %v", err)
	}

	for _, pt := range points {
		bp.AddPoint(pt)
	}

	/*points, errs = e.getElectricityConsumption()
	if len(errs) > 0 {
		errors = append(errors, errs...)
	}

	for _, pt := range points {
		bp.AddPoint(pt)
	}*/

	err = e.db.Write(bp)
	if err != nil {
		return fmt.Errorf("failed to write to influxdb: %v", err)
	}

	return nil
}

func (e *Exporter) getElectricityConsumption() ([]*influxdb.Point, error) {
	var points []*influxdb.Point

	for _, mp := range e.meterPoints.ElectricityMeterPoints {
		consumption, err := e.client.GetElecMeterConsumption(
			mp.MPAN,
			mp.SerialNumber,
			octopusenergyapi.ConsumptionOption{},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get electricity consumption for %s/%s: %v", mp.MPAN, mp.SerialNumber, err)
		}

		tags := map[string]string{
			"mpan":          mp.MPAN,
			"serial_number": mp.SerialNumber,
		}

		for _, interval := range consumption {
			fields := map[string]interface{}{
				"value": interval.Value,
			}

			pt, err := influxdb.NewPoint(
				"electricity_consumption",
				tags,
				fields,
				interval.IntervalEnd,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create point at %s for %s/%s: %v", interval.IntervalEnd, mp.MPAN, mp.SerialNumber, err)
			}

			points = append(points, pt)
		}
	}

	return points, nil
}

func (e *Exporter) exportGasConsumption(wg *sync.WaitGroup, ch chan<- *influxdb.Point) {
	defer wg.Done()

	for _, meterPoint := range e.meterPoints.GasMeterPoints {
		logFields := map[string]interface{}{
			"mprn":          meterPoint.MPRN,
			"serial_number": meterPoint.SerialNumber,
		}

		consumption, err := e.client.GetGasMeterConsumption(
			meterPoint.MPRN,
			meterPoint.SerialNumber,
			octopusenergyapi.ConsumptionOption{},
		)
		if err != nil {
			e.logger.
				WithFields(logFields).
				WithError(err).
				Error("failed to get gas consumption")
			continue
		}

		tags := map[string]string{
			"mprn":          meterPoint.MPRN,
			"serial_number": meterPoint.SerialNumber,
		}

		for _, interval := range consumption {
			fields := map[string]interface{}{
				"value": interval.Value,
			}

			e.logger.
				WithFields(logFields).
				WithField("interval_end", interval.IntervalEnd).
				WithField("value", interval.Value).
				Info("adding point")

			pt, err := influxdb.NewPoint(
				"gas_consumption",
				tags,
				fields,
				interval.IntervalEnd,
			)
			if err != nil {
				e.logger.
					WithFields(logFields).
					WithField("interval_end", interval.IntervalEnd).
					WithError(err).
					Error("failed to create point")
				continue
			}

			ch <- pt
		}
	}
}
