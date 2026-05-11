package handlers

import (
	"authentication/business"
	"authentication/commons"
	"authentication/commons/constants"
	"fmt"
	"net/http"
	genericConstants "stock_broker_application/src/constants"
	"stock_broker_application/src/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LogoutHandler struct {
	service *business.LogoutService
}

func NewLogoutHandler(service *business.LogoutService) *LogoutHandler {
	return &LogoutHandler{
		service: service,
	}
}

// Logout Logs out the user.
// @Summary Logs out a loggedin user
// @Description Handles user logout by validating jwt token
// @Tags User
// @Produce json
// @Success 200 {string} string "User logged out successfully"
// @Failure 401 {object} models.ErrorAPIResponse "unauthorized"
// @Failure 500 {object} models.ErrorAPIResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/auth/logout [post]
func (controller LogoutHandler) Logout(ctx *gin.Context) {

	tokenString := ctx.GetString(genericConstants.Token)

	// tokenExpiry := ctx.GetInt64(commons.TokenExpiry)
	tokenExpiry := ctx.MustGet(commons.TokenExpiry).(time.Time)
	fmt.Println("token expiry: ", tokenExpiry)

	// ttl := time.Until(time.Unix(tokenExpiry, 0))
	ttl := time.Until(tokenExpiry)
	fmt.Println("TTL: ", ttl)

	logrus.SetLevel(logrus.WarnLevel)

	err := controller.service.Logout(ctx, tokenString, ttl)
	if err != nil {
		logrus.Error("ERROR: ", err)
		if strings.Contains(err.Error(), genericConstants.RedisClientNotInitialized) {
			ctx.IndentedJSON(http.StatusInternalServerError, models.ErrorAPIResponse{
				Message: models.ErrorMessage{
					Key:          genericConstants.Server,
					ErrorMessage: genericConstants.RedisClientNotInitialized,
				},
				Error: genericConstants.ErrInternalServer,
			})
			logrus.Error("redis client not initialized")
			return
		}

		ctx.IndentedJSON(http.StatusInternalServerError, models.ErrorAPIResponse{
			Message: models.ErrorMessage{
				Key:          genericConstants.Server,
				ErrorMessage: constants.LogoutFailedError,
			},
			Error: genericConstants.ErrInternalServer,
		})
		logrus.Info("redis error: ", err)
		return
	}

	ctx.IndentedJSON(http.StatusOK, constants.LogoutSuccessfulMsg)
}
