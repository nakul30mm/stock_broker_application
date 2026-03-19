package handlers

import (
	"watchlists/commons/constants"

	"encoding/json"
	"net/http"
	"stock_broker_application/src/utils/validations"
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

// this fucntion handles user requests for add, delete and get scrips to/ from a list of watchlists
// Handles ADG scrip to watchlist functionality
// @Summary Perform ADG for watchlist
// @Description verifies the JWT and adds/ deletes/ gets the scrip in the request to/ from the list of watchlists
// @Tags Watchlist
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.BffAdgStoWatchlistRequest true "ADG Watchlist Request"
// @Success 200 {object} models.BffAdgStoWatchlistResponse "ADG Performed successfully"
// @Failure 400 {object} models.ErrorAPIResponse "Invalid input payload"
// @Failure 401 {object} models.ErrorAPIResponse "Case A: JWT expired / Case B: JWT invalid / Case C: Unauthorized request"
// @Failure 404 {object} models.ErrorAPIResponse "User does not exist"
// @Failure 500 {object} models.ErrorAPIResponse "Internal Server Error"
// @Router /adg/scrip [post]
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

	if err := validations.GetBFFValidator().Struct(&bffAdgStoWatchlistRequest); err != nil {
		validationError, _ := validations.FormatValidationErrors(err)
		ctx.IndentedJSON(http.StatusBadRequest, validationError)
		return
	}

	username := ctx.GetString(commons.Username)

	warnings, respWatchlistsWithIds, err := controller.AdgStoWatchlistService.AdgStoWatchlist(ctx, username, bffAdgStoWatchlistRequest)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	ctx.IndentedJSON(http.StatusOK, models.BffAdgStoWatchlistResponse{
		Status:          "success",
		Action:          models.Actiontype(bffAdgStoWatchlistRequest.Action),
		WatchlistWithId: respWatchlistsWithIds,
		Warnings:        warnings,
	})
}
