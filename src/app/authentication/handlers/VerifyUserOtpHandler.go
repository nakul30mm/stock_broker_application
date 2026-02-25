package handlers

import (
	"authentication/business"
	"authentication/commons"
	"authentication/commons/constants"
	"authentication/models"
	"encoding/json"
	"errors"
	"net/http"
	genericModels "stock_broker_application/src/models"
	"stock_broker_application/src/utils/validations"

	"github.com/gin-gonic/gin"
)

type ValidateUserOtpHandler struct {
	service *business.ValidateUserOtpService
}

func NewValidateUserOtpHandler(service *business.ValidateUserOtpService) *ValidateUserOtpHandler {
	return &ValidateUserOtpHandler{
		service: service,
	}
}

func (controller *ValidateUserOtpHandler) HandleValidateUserOtp(ctx *gin.Context) {
	var bffValidateUserOtpRequest models.BFFValidateUserOtpRequest
	if errWhileBindingReq := ctx.ShouldBind(&bffValidateUserOtpRequest); errWhileBindingReq != nil {
		errorMessage := genericModels.ErrorMessage{
			Key:          errWhileBindingReq.(*json.UnmarshalTypeError).Field,
			ErrorMessage: constants.ErrUnexpectedValue,
		}

		ctx.IndentedJSON(http.StatusBadRequest, genericModels.ErrorAPIResponse{
			Message: errorMessage,
			Error:   constants.ErrInvalidPayload,
		})
		return
	}

	if errWhileValidations := validations.GetBFFValidator().Struct(&bffValidateUserOtpRequest); errWhileValidations != nil {
		validationErrors, _ := validations.FormatValidationErrors(errWhileValidations)
		ctx.IndentedJSON(http.StatusBadRequest, validationErrors)
		return
	}

	errWhileOtpValidation := controller.service.ValidateUserOtp(ctx, ctx.Request.Context(), bffValidateUserOtpRequest)
	if errWhileOtpValidation != nil {
		if errors.Is(errWhileOtpValidation, commons.UserNotFoundError) {
			errorUserNotFoundResponse := genericModels.ErrorAPIResponse{
				Message: genericModels.ErrorMessage{
					Key:          commons.Username,
					ErrorMessage: constants.ErrUserNotFound,
				},
				Error: constants.ErrAuthenticationFailed,
			}
			ctx.IndentedJSON(http.StatusBadRequest, errorUserNotFoundResponse)
			return
		}

		if errors.Is(errWhileOtpValidation, commons.IncorrectOTPError) {
			errorIncorrectOtpResponse := genericModels.ErrorAPIResponse{
				Message: genericModels.ErrorMessage{
					Key:          commons.Otp,
					ErrorMessage: constants.ErrIncorrectOtp,
				},
				Error: constants.ErrAuthenticationFailed,
			}
			ctx.IndentedJSON(http.StatusBadRequest, errorIncorrectOtpResponse)
			return
		}

		if errors.Is(errWhileOtpValidation, commons.OtpExpired) {
			errorExpiredOtpResponse := genericModels.ErrorAPIResponse{
				Message: genericModels.ErrorMessage{
					Key:          commons.Otp,
					ErrorMessage: constants.ErrExpiredOtp,
				},
				Error: constants.ErrAuthenticationFailed,
			}
			ctx.IndentedJSON(http.StatusBadRequest, errorExpiredOtpResponse)
			return
		}

		ctx.IndentedJSON(http.StatusUnauthorized, genericModels.ErrorAPIResponse{
			Error: constants.ErrSignInFailed,
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, constants.OtpValidatedSuccessMsg)
}
