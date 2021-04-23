package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"adtennant.dev/octopus-consumption-influxdb/config"
	"adtennant.dev/octopus-consumption-influxdb/exporter"
	"github.com/FileGo/octopusenergyapi"
	"github.com/caarlos0/env/v6"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func loadConfig(filename string) (*config.MeterPoints, error) {
	var mp config.MeterPoints

	configFile, err := ioutil.ReadFile(filename)
	err = yaml.Unmarshal(configFile, &mp)
	if err != nil {
		return nil, err
	}

	return &mp, nil
}

func main() {
	logger := logrus.New()

	config := config.Config{}
	if err := env.Parse(&config); err != nil {
		fmt.Printf("%+v\n", err)
	}

	meterPoints, err := loadConfig(config.ConfigPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to load meter points")
	}

	client, err := octopusenergyapi.NewClient(config.Octopus.APIKey, http.DefaultClient)
	if err != nil {
		logger.WithError(err).Fatal("failed to create octopus client")
	}

	db, err := influxdb.NewHTTPClient(influxdb.HTTPConfig{
		Addr:     config.InfluxDB.Address,
		Username: config.InfluxDB.Username,
		Password: config.InfluxDB.Password,
	})
	if err != nil {
		logger.WithError(err).Fatal("failed to create influxdb client")
	}

	q := influxdb.NewQuery(fmt.Sprintf("CREATE DATABASE %s", config.InfluxDB.Database), "", "")

	_, err = db.Query(q)
	if err != nil {
		logger.WithError(err).Fatal("failed to create database")
	}

	exporter := exporter.New(meterPoints, client, db, logger)

	err = exporter.Export(config.InfluxDB.Database)
	if err != nil {
		logger.WithError(err).Fatal("failed to export")
	}
}
