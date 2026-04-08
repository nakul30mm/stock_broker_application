package handlers

import (
	"authentication/business"
	"net/http"
	"stock_broker_application/src/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type LogoutHandler struct {
	service *business.LogoutService
}

func NewLogoutHandler(service *business.LogoutService) *LogoutHandler {
	return &LogoutHandler{
		service: service,
	}
}

func (controller LogoutHandler) Logout(ctx *gin.Context) {
	header := ctx.GetHeader("Authorization")
	if header == "" {
		ctx.IndentedJSON(http.StatusUnauthorized, "missing authorization header")
		return
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.IndentedJSON(http.StatusUnauthorized, "invalid authorization header")
		return
	}

	tokenString := parts[1]
	ttl, err := utils.ExtractExpiry(tokenString)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, "invalid token")
		return
	}
	err = controller.service.Logout(ctx, tokenString, ttl)
	if err != nil {
		if err.Error() == "token is already blacklisted" {
			ctx.IndentedJSON(http.StatusBadRequest, "user already logged out")
			return
		}

		ctx.IndentedJSON(http.StatusInternalServerError, "logout failed")
		return
	}

	ctx.IndentedJSON(http.StatusOK, "user logged out successfully")
}
