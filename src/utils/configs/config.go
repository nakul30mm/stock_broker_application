package configs

import (
	"fmt"
	"os"
	"stock_broker_application/src/constants"

	"github.com/spf13/viper"
)

func LoadConfig[T any](configPath, configFile, configType string) (*T, error) {
	var config T

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configFile)
	viper.SetConfigType(configType)
	fmt.Println(os.Getwd())
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf(constants.ErrReadConfigFailed, err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf(constants.ErrUnmarshallConfigFailed, err)
	}
	fmt.Print(config)
	return &config, nil
}
