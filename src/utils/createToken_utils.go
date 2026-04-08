package utils

import (
	"errors"
	"stock_broker_application/src/constants"
	"stock_broker_application/src/models"
	"stock_broker_application/src/utils/configs"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var SecretKey *models.JWT

func GenerateToken(username string) (string, string, error) {

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 12).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(SecretKey.AccessSecretKey))
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(SecretKey.RefreshSecretKey))
	if err != nil {
		return "", "", err
	}
	return accessTokenString, refreshTokenString, nil
}

func InitJWTConfig(configPath string) error {
	var err error
	SecretKey, err = configs.LoadConfig[models.JWT](configPath, constants.JWT, constants.Yaml)
	if err != nil {
		return err
	}
	return nil
}

func ExtractExpiry(tokenString string) (time.Duration, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey.AccessSecretKey), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		exp := int64(claims["exp"].(float64))
		expTime := time.Unix(exp, 0)
		return time.Until(expTime), nil
	}

	return 0, errors.New("invalid token")
}
