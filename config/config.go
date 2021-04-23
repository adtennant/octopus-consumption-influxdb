package config

type InfluxDB struct {
	Address  string `env:"INFLUXDB_ADDRESS,required"`
	Database string `env:"INFLUXDB_DATABASE,required"`
	Username string `env:"INFLUXDB_USERNAME,required"`
	Password string `env:"INFLUXDB_PASSWORD,required"`
}

type Octopus struct {
	APIKey string `env:"OCTOPUS_API_KEY,required"`
}

type Config struct {
	ConfigPath string `env:"CONFIG_PATH,required"`
	InfluxDB   InfluxDB
	Octopus    Octopus
}

type ElectricityMeterPoint struct {
	MPAN         string `yaml:"mpan"`
	SerialNumber string `yaml:"serialNumber"`
}

type GasMeterPoint struct {
	MPRN         string `yaml:"mprn"`
	SerialNumber string `yaml:"serialNumber"`
}

type MeterPoints struct {
	ElectricityMeterPoints []ElectricityMeterPoint `yaml:"electricityMeterPoints"`
	GasMeterPoints         []GasMeterPoint         `yaml:"gasMeterPoints"`
}
