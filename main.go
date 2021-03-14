package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"adtennant.dev/octopus-consumption-influxdb/config"
	"adtennant.dev/octopus-consumption-influxdb/exporter"
	"github.com/FileGo/octopusenergyapi"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func loadConfig(filename string) (*config.Config, error) {
	var c config.Config

	configFile, err := ioutil.ReadFile(filename)
	err = yaml.Unmarshal(configFile, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func main() {
	logger := logrus.New()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		logger.Fatal("missing CONFIG_PATH")
	}

	config, err := loadConfig(configPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to load config")
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

	exporter := exporter.New(config, client, db, logger)

	err = exporter.Export()
	if err != nil {
		logger.WithError(err).Fatal("failed to export")
	}
}
