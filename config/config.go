package config

type ElectricityMeterPoint struct {
	MPAN         string `yaml:"mpan"`
	SerialNumber string `yaml:"serialNumber"`
}

type GasMeterPoint struct {
	MPRN         string `yaml:"mprn"`
	SerialNumber string `yaml:"serialNumber"`
}

type InfluxDB struct {
	Address  string `yaml:"address"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Octopus struct {
	APIKey string `yaml:"apikey"`
}

type Config struct {
	ElectricityMeterPoints []ElectricityMeterPoint `yaml:"electricityMeterPoints"`
	GasMeterPoints         []GasMeterPoint         `yaml:"gasMeterPoints"`
	InfluxDB               InfluxDB                `yaml:"influxdb"`
	Octopus                Octopus                 `yaml:"octopus"`
}
