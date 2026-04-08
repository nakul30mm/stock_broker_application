package middleware

import (
	"authentication/commons"
	"authentication/commons/constants"
	"net/http"
	"stock_broker_application/src/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := ValidateToken(ctx, rdb); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorAPIResponse{
				Message: models.ErrorMessage{
					Key:          commons.Token,
					ErrorMessage: err.Error(),
				},
				Error: constants.ErrInvalidToken,
			})
			return
		}
		ctx.Next()
	}
}
