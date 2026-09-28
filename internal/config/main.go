package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_url 			  string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func Read() (*Config, error){
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return &Config{}, err
	}
	fileData, err := os.ReadFile(configFilePath)

	var configData Config
	if err := json.Unmarshal(fileData, &configData); err != nil {
		return &Config{}, fmt.Errorf("Error while unmarshaling json data: %v", err)
	}

	return &configData, nil
}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error while getting home dir: %v", err)
	}

	configFilePath := fmt.Sprintf("%s/%s", homeDir, configFileName)

	return configFilePath, nil
}

func (cfg *Config) SetUser(userName string) error {
	cfg.Current_user_name = userName

	return write(*cfg)
}

func write(cfg Config) error {
	encodedConfigData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Error while encoding json: %v", err)
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("Error while getting home dir: %v", err)
	}
	
	err = os.WriteFile(configFilePath, encodedConfigData, 0666)
	if err != nil {
		return fmt.Errorf("Error while writing data into config: %v", err)
	}

	return nil
}
