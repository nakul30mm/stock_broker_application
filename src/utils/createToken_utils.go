package utils

import (
	"stock_broker_application/src/constants"
	"stock_broker_application/src/models"
	"stock_broker_application/src/utils/configs"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var SecretKey *models.JWT

func GenerateToken(username string, deviceType string) (string, string, string, error) {

	jti := uuid.New().String()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         username,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
		"jti":         jti,
		"device_type": deviceType,
	})

	accessTokenString, err := accessToken.SignedString([]byte(SecretKey.AccessSecretKey))
	if err != nil {
		return "", "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":         username,
		"iat":         time.Now().Unix(),
		"exp":         time.Now().Add(time.Hour * 24 * 30).Unix(),
		"jti":         jti,
		"device_type": deviceType,
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(SecretKey.RefreshSecretKey))
	if err != nil {
		return "", "", "", err
	}
	return accessTokenString, refreshTokenString, jti, nil
}

func InitJWTConfig(configPath string) error {
	var err error
	SecretKey, err = configs.LoadConfig[models.JWT](configPath, constants.JWT, constants.Yaml)
	if err != nil {
		return err
	}
	return nil
}
