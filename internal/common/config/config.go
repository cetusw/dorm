package config

import (
	"encoding/json"
	"os"

	"dorm/internal/dorm/application/model"
)

func LoadConfig(fileName string) (*model.Config, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config model.Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
