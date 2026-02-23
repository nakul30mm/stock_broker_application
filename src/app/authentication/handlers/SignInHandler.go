package handlers

import (
	"authentication/business"
	"authentication/commons/constants"
	"authentication/models"
	"encoding/json"
	"net/http"

	genericModels "stock_broker_application/src/models"
	"stock_broker_application/src/utils/validations"

	"github.com/gin-gonic/gin"
)

type SigninUserHandler struct {
	service *business.SigninUserService
}

func NewSigninUserHandler(service *business.SigninUserService) *SigninUserHandler {
	return &SigninUserHandler{
		service: service,
	}
}

// HandleSigninUser handles the user signin request.
// @Summary Sign in user
// @Description Authenticate user and return JWT tokens
// @Tags User
// @Accept json
// @Produce json
// @Param request body models.BFFSigninUserRequest true "Signin Request"
// @Success 200 {object} models.BFFSigninUserResponse
// @Failure 400 {object} models.ErrorAPIResponse
// @Failure 401 {object} models.ErrorAPIResponse
// @Router /api/auth/signin [post]
func (controller *SigninUserHandler) HandleSigninUser(ctx *gin.Context) {

	var bffSigninUserRequet models.BFFSigninUserRequest

	if err := ctx.ShouldBind(&bffSigninUserRequet); err != nil {
		errorMsgs := genericModels.ErrorMessage{
			Key:          err.(*json.UnmarshalTypeError).Field,
			ErrorMessage: constants.ErrUnexpectedValue,
		}
		ctx.JSON(http.StatusBadRequest, genericModels.ErrorAPIResponse{
			Message: errorMsgs,
			Error:   constants.ErrInvalidPayload,
		})
		return
	}

	if err := validations.GetBFFValidator().Struct(&bffSigninUserRequet); err != nil {
		validationErrors, _ := validations.FormatValidationErrors(err)
		ctx.JSON(http.StatusBadRequest, validationErrors)
		return
	}

	err := controller.service.SigninUser(ctx, ctx.Request.Context(), bffSigninUserRequet)
	if err != nil {

		if err.Error() == constants.ErrInvalidEmailorPassword {
			errorResponse := genericModels.ErrorAPIResponse{
				Message: genericModels.ErrorMessage{
					Key:          "Authentication Error",
					ErrorMessage: constants.ErrInvalidEmailorPassword,
				},
				Error: constants.ErrConflict,
			}
			ctx.IndentedJSON(http.StatusUnauthorized, errorResponse)
			return

		}
		errorResponse := genericModels.ErrorAPIResponse{
        Message: genericModels.ErrorMessage{
            Key:          "INTERNAL_ERROR",
            ErrorMessage: "Something went wrong",
        },
        Error: constants.ErrInternalServer,
    	}

		ctx.IndentedJSON(http.StatusInternalServerError, errorResponse)
		return
	}

	ctx.IndentedJSON(http.StatusOK, constants.UserLoggedInSuccessMsg)
}
