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
	config *config.Config

	client *octopusenergyapi.Client
	db     influxdb.Client

	logger *logrus.Logger
}

func New(config *config.Config, client *octopusenergyapi.Client, db influxdb.Client, logger *logrus.Logger) *Exporter {
	return &Exporter{
		config,

		client,
		db,

		logger,
	}
}

func (e *Exporter) Export() error {
	e.logger.Info("starting export")

	var wg sync.WaitGroup
	ch := make(chan *influxdb.Point)

	wg.Add(2)

	go func() {
		wg.Wait()
		close(ch)
	}()

	go e.exportElectricityConsumption(&wg, ch)
	go e.exportGasConsumption(&wg, ch)

	bp, _ := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  e.config.InfluxDB.Database,
		Precision: "s",
	})

	for pt := range ch {
		bp.AddPoint(pt)
	}

	e.logger.WithField("count", len(bp.Points())).Info("writing points to influxdb")

	err := e.db.Write(bp)
	if err != nil {
		return fmt.Errorf("failed to write to influxdb: %v", err)
	}

	e.logger.Info("export complete")

	return nil
}

func (e *Exporter) exportElectricityConsumption(wg *sync.WaitGroup, ch chan<- *influxdb.Point) {
	defer wg.Done()

	for _, meterPoint := range e.config.ElectricityMeterPoints {
		logFields := map[string]interface{}{
			"mpan":          meterPoint.MPAN,
			"serial_number": meterPoint.SerialNumber,
		}

		consumption, err := e.client.GetElecMeterConsumption(
			meterPoint.MPAN,
			meterPoint.SerialNumber,
			octopusenergyapi.ConsumptionOption{},
		)
		if err != nil {
			e.logger.
				WithFields(logFields).
				WithError(err).
				Error("failed to get electricity consumption")
			continue
		}

		tags := map[string]string{
			"mpan":          meterPoint.MPAN,
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
				"electricity_consumption",
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

func (e *Exporter) exportGasConsumption(wg *sync.WaitGroup, ch chan<- *influxdb.Point) {
	defer wg.Done()

	for _, meterPoint := range e.config.GasMeterPoints {
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
