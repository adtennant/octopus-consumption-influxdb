package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"adtennant.dev/octopus-consumption-exporter/collector"
	"adtennant.dev/octopus-consumption-exporter/config"
	"github.com/FileGo/octopusenergyapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	apiKey := os.Getenv("OCTOPUS_API_KEY")
	if apiKey == "" {
		logger.Fatal("missing OCTOPUS_API_KEY")
	}

	configPath := os.Getenv("OCTOPUS_CONFIG_PATH")
	if configPath == "" {
		logger.Fatal("missing OCTOPUS_CONFIG_PATH")
	}

	config, err := loadConfig(configPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to load config")
	}

	client, err := octopusenergyapi.NewClient(apiKey, http.DefaultClient)
	if err != nil {
		logger.WithError(err).Fatal("failed to create client")
	}

	collector := collector.New(config, client, logger)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
