package collector

import (
	"adtennant.dev/octopus-consumption-exporter/config"
	"github.com/FileGo/octopusenergyapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type Collector struct {
	config *config.Config
	client *octopusenergyapi.Client

	logger *logrus.Logger

	electricityConsumption *prometheus.Desc
	gasConsumption         *prometheus.Desc
}

func New(config *config.Config, client *octopusenergyapi.Client, logger *logrus.Logger) *Collector {
	return &Collector{
		config: config,
		client: client,

		logger: logger,

		electricityConsumption: prometheus.NewDesc("electricity_consumption_kwh", "Electricity Consumption in kWh", []string{"mpan", "serial_number"}, nil),
		gasConsumption:         prometheus.NewDesc("gas_consumption_m3", "Gas Consumption in m^3", []string{"mprn", "serial_number"}, nil),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.electricityConsumption
	ch <- c.gasConsumption
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Info("running scrape")

	collectElectricityConsumption(c, ch)
	collectGasConsumption(c, ch)

	c.logger.Info("scrape complete")
}

func collectElectricityConsumption(c *Collector, ch chan<- prometheus.Metric) {
	for _, electricityMeterPoint := range c.config.ElectricityMeterPoints {
		consumption, err := c.client.GetElecMeterConsumption(
			electricityMeterPoint.MPAN,
			electricityMeterPoint.SerialNumber,
			octopusenergyapi.ConsumptionOption{},
		)
		if err != nil {
			c.logger.WithError(err).Error("error during electricity scrape")
			continue
		}

		latest := consumption[0]
		ch <- prometheus.NewMetricWithTimestamp(
			latest.IntervalEnd,
			prometheus.MustNewConstMetric(
				c.electricityConsumption,
				prometheus.GaugeValue,
				float64(latest.Value),
				electricityMeterPoint.MPAN,
				electricityMeterPoint.SerialNumber,
			),
		)
	}
}

func collectGasConsumption(c *Collector, ch chan<- prometheus.Metric) {
	for _, gasMeterPoint := range c.config.GasMeterPoints {
		consumption, err := c.client.GetElecMeterConsumption(
			gasMeterPoint.MPRN,
			gasMeterPoint.SerialNumber,
			octopusenergyapi.ConsumptionOption{},
		)
		if err != nil {
			c.logger.WithError(err).Error("error during gas scrape")
			continue
		}

		latest := consumption[0]
		ch <- prometheus.NewMetricWithTimestamp(
			latest.IntervalEnd,
			prometheus.MustNewConstMetric(
				c.gasConsumption,
				prometheus.GaugeValue,
				float64(latest.Value),
				gasMeterPoint.MPRN,
				gasMeterPoint.SerialNumber,
			),
		)
	}
}
