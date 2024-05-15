package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	VmDB     VmDBConfig
	DataBase DataBaseConfig
}

type DataBaseConfig struct {
	File string
}

type VmDBConfig struct {
	Url string
}

type ServerConfig struct {
	HttpPort string
}

var Cfg Config

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.SetConfigType("yaml")

	viper.SetDefault("server.httpPort", "8080")
	viper.SetDefault("vmDB.url", "http://localhost:8428")
	viper.SetDefault("database.file", "./config/database.db")

	// ENV
	viper.BindEnv("server.httpPort", "HTTP_PORT")
	viper.BindEnv("vmDB.url", "VM_DB_URL")
	viper.BindEnv("database.file", "DATABASE_FILE")

	if err := os.MkdirAll("./config", 0755); err != nil {
		panic(err)
	}
	if _, err := os.Stat("./config/config.yaml"); os.IsNotExist(err) {
		logrus.Info("Config file not found, creating a new one")
		if err := os.WriteFile("./config/config.yaml", []byte(""), 0644); err != nil {
			panic(err)
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&Cfg); err != nil {
		panic(err)
	}
}

func GetConfig() *Config {
	return &Cfg
}

func GetVmDBConfig() *VmDBConfig {
	return &Cfg.VmDB
}

func GetServerConfig() *ServerConfig {
	return &Cfg.Server
}

func GetDataBaseConfig() *DataBaseConfig {
	return &Cfg.DataBase
}
