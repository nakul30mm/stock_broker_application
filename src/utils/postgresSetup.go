package utils

import (
	"fmt"
	"stock_broker_application/src/constants"
	"stock_broker_application/src/models"
	"stock_broker_application/src/utils/configs"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var postgresClient *models.DatabaseConfiguration
var postgresConfig *models.PostgresConfig

func InitPostgresConfg(configPath string) error {

	if err := initPostgresConfig(configPath); err != nil {
		return fmt.Errorf(constants.ErrLoadConfigFailed, err)
	}
	dns := fmt.Sprintf(constants.DSNString, postgresConfig.Host, postgresConfig.Port, postgresConfig.Username, postgresConfig.Password, postgresConfig.DBName, postgresConfig.SSLMode, postgresConfig.Timezone)

	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return fmt.Errorf(constants.ErrPostgresConnectionFailed, err)
	}

	setDBInstance(db)
	return nil
}

func setDBInstance(db *gorm.DB) {
	postgresClient = &models.DatabaseConfiguration{GormDB: db}
}

func GetPostgresClient() *models.DatabaseConfiguration {
	return postgresClient
}

func initPostgresConfig(configPath string) error {
	var err error
	postgresConfig, err = configs.LoadConfig[models.PostgresConfig](configPath, constants.Postgres, constants.Yaml)
	if err != nil {
		return err
	}
	return nil
}
