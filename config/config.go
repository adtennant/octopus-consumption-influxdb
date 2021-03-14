package config

type ElectricityMeterPoint struct {
	MPAN         string `yaml:"mpan"`
	SerialNumber string `yaml:"serialNumber"`
}

type GasMeterPoint struct {
	MPRN         string `yaml:"mprn"`
	SerialNumber string `yaml:"serialNumber"`
}

type Config struct {
	ElectricityMeterPoints []ElectricityMeterPoint `yaml:"electricityMeterPoints"`
	GasMeterPoints         []GasMeterPoint         `yaml:"gasMeterPoints"`
}
