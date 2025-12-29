package config

var AppConfig *Config

type Config struct {
	Port string
	Mode string
}

func Init() {
	AppConfig = &Config{
		Port: ":8080",
		Mode: "debug",
	}
}
