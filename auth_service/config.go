package main

import (
	"auth_service/database"
	"os"
	"strconv"
)

type Config struct {
	Port   uint64            `json:"port"`
	DBConf database.DBConfig `json:"db"`
}

func ReadConfig() (*Config, error) {
	config := Config{}
	config.DBConf.DBHost = os.Getenv("DB_HOST")
	config.DBConf.DBUser = os.Getenv("DB_USER")
	config.DBConf.DBPassword = os.Getenv("DB_PASSWORD")
	config.DBConf.DBName = os.Getenv("DB_NAME")
	var err error
	config.DBConf.DBPort, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}

	config.Port, err = strconv.ParseUint(os.Getenv("APP_PORT"), 10, 0)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
