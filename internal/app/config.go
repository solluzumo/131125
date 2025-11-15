package app

type Config struct {
	JsonDir     string
	PdfDir      string
	WorkersNum  int
	TaskChanCap int
}

func NewConfig() *Config {
	return &Config{
		JsonDir:     "./data",
		PdfDir:      "./data/pdf",
		WorkersNum:  2,
		TaskChanCap: 5,
	}
}
