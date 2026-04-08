package middleware

import (
	"authentication/commons"
	"errors"
	"fmt"
	"stock_broker_application/src/utils"

	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// extracts token claims
func ValidateToken(ctx *gin.Context, rdb *redis.Client) error {
	//extracting the header
	header := ctx.GetHeader("Authorization")
	if header == "" {
		return errors.New("missing authorization header")
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return errors.New("invalid authorization format")
	}
	tokenString := parts[1]

	key := fmt.Sprintf("BLACKLISTED_TOKEN_%s", tokenString)
	val, err := rdb.Get(ctx, key).Result()
	if err == nil && val == "1" {
		return errors.New("token is blacklisted")
	}
	if err != nil && err != redis.Nil {
		return err
	}
	//validatig the token signature
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(utils.SecretKey.AccessSecretKey), nil
	})
	if err != nil {
		return err
	}

	//extracting claims for verification
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {

		//checking expiry
		exp, ok := claims["exp"].(float64)
		if !ok {
			return errors.New("invalid exp claim")
		}
		if int64(exp) < time.Now().Unix() {
			return errors.New("token expired")
		}

		//extracting username and binding with the context
		username, ok := claims["sub"].(string)
		if !ok {
			return errors.New("invalid sub claims")
		}
		ctx.Set(commons.Username, username)
		return nil
	}

	return errors.New("invalid token, cannot extract claims")
}
