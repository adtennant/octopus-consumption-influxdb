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
	ch := make(chan influxdb.Point)

	wg.Add(1)

	go func() {
		wg.Wait()
		close(ch)
	}()

	go e.exportElectricityConsumption(&wg, ch)
	//e.exportGasConsumption()

	bp, _ := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  e.config.InfluxDB.Database,
		Precision: "s",
	})

	for pt := range ch {
		bp.AddPoint(&pt)
	}

	e.logger.WithField("count", len(bp.Points())).Info("writing points")

	err := e.db.Write(bp)
	if err != nil {
		return fmt.Errorf("failed to write to influxdb: %v", err)
	}

	e.logger.Info("export complete")

	return nil
}

func (e *Exporter) exportElectricityConsumption(wg *sync.WaitGroup, ch chan<- influxdb.Point) {
	defer wg.Done()

	for _, electricityMeterPoint := range e.config.ElectricityMeterPoints {
		logFields := map[string]interface{}{
			"mpan":          electricityMeterPoint.MPAN,
			"serial_number": electricityMeterPoint.SerialNumber,
		}

		consumption, err := e.client.GetElecMeterConsumption(
			electricityMeterPoint.MPAN,
			electricityMeterPoint.SerialNumber,
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
			"mpan":          electricityMeterPoint.MPAN,
			"serial_number": electricityMeterPoint.SerialNumber,
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

			ch <- *pt
		}
	}
}

/*func (e *Exporter) exportGasConsumption() {
for _, gasMeterPoint := range e.config.GasMeterPoints {
	e.logger.
		WithField("mprn", gasMeterPoint.MPRN).
		WithField("serial number", gasMeterPoint.SerialNumber).
		Info("exporting gas consumption")

	consumption, err := e.client.GetElecMeterConsumption(
		gasMeterPoint.MPRN,
		gasMeterPoint.SerialNumber,
		octopusenergyapi.ConsumptionOption{},
	)
	if err != nil {
		e.logger.WithError(err).Error("failed exporting gas consumption")
		continue
	}

	/*latest := consumption[0]
	c.logger.
		WithField("timestamp", latest.IntervalEnd).
		Info("found latest gas consumption")

	ch <- prometheus.NewMetricWithTimestamp(
		latest.IntervalEnd,
		prometheus.MustNewConstMetric(
			c.gasConsumption,
			prometheus.GaugeValue,
			float64(latest.Value),
			gasMeterPoint.MPRN,
			gasMeterPoint.SerialNumber,
		),
	)*/
/*}
}*/
