package app

type Config struct {
	DataDir    string
	WorkersNum int
}

func NewConfig() *Config {
	return &Config{
		DataDir:    "./data",
		WorkersNum: 2,
	}
}
