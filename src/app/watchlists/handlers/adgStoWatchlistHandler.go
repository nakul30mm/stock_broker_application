package handlers

import (
	"watchlists/commons/constants"

	"encoding/json"
	"net/http"
	"watchlists/business"
	"watchlists/commons"
	"watchlists/models"

	genericModels "stock_broker_application/src/models"

	"github.com/gin-gonic/gin"
)

type AdgStoWatchlistHandler struct {
	AdgStoWatchlistService *business.AdgStoWatchlistService
}

func NewAdgStoWatchlistHandler(adgStoWatchlistService *business.AdgStoWatchlistService) *AdgStoWatchlistHandler {
	return &AdgStoWatchlistHandler{
		AdgStoWatchlistService: adgStoWatchlistService,
	}
}

func (controller *AdgStoWatchlistHandler) HandleAdgStoWatchlist(ctx *gin.Context) {
	var bffAdgStoWatchlistRequest models.BffAdgStoWatchlistRequest
	// reqAction := strings.ToLower((bffAdgStoWatchlistRequest.Action))

	if err := ctx.ShouldBind(&bffAdgStoWatchlistRequest); err != nil {
		errorMsg := genericModels.ErrorMessage{
			Key:          err.(*json.UnmarshalTypeError).Field,
			ErrorMessage: constants.ErrUnexpectedValue,
		}
		ctx.IndentedJSON(http.StatusBadRequest, genericModels.ErrorAPIResponse{
			Message: errorMsg,
			Error:   constants.ErrInvalidPayload,
		})
		return
	}

	// if err := validations.GetBFFValidator().Struct(&bffAdgStoWatchlistRequest); err != nil {
	// 	validationError, _ := validations.FormatValidationErrors(err)
	// 	ctx.IndentedJSON(http.StatusBadRequest, validationError)
	// 	return
	// }

	username := ctx.GetString(commons.Username)

	warnings, respWatchlistsWithIds, err := controller.AdgStoWatchlistService.AdgStoWatchlist(ctx, username, bffAdgStoWatchlistRequest)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.IndentedJSON(http.StatusOK, models.BffAdgStoWatchlistResponse{
		Status:          "success",
		Action:          constants.Actiontype(bffAdgStoWatchlistRequest.Action),
		WatchlistWithId: respWatchlistsWithIds,
		Warnings:        warnings,
	})
}
