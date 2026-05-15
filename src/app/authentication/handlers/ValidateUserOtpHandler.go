package handlers

import (
	"authentication/business"
	"authentication/commons"
	"authentication/commons/constants"
	"authentication/models"
	"encoding/json"
	"fmt"
	"net/http"
	genericConstants "stock_broker_application/src/constants"
	genericModels "stock_broker_application/src/models"
	"stock_broker_application/src/utils/validations"
	"strings"

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

// this fucntion handles user requests and responses by
// Handles user OTP validation
// @Summary Validates user OTP
// @Description Validates user OTP and return clear success/ failure message
// @Tags User
// @Accept json
// @Produce json
// @Param request body models.BFFValidateUserOtpRequest true "User OTP Validation Request"
// @Success 200 {object} models.BFFValidateUserOtpResponse "OTP validation successful"
// @Failure 400 {object} models.ErrorAPIResponse "Invalid input payload"
// @Failure 401 {object} models.ErrorAPIResponse "Case A: Incorrect OTP / Case B: Expired OTP"
// @Failure 404 {object} models.ErrorAPIResponse "User does not exist"
// @Failure 500 {object} models.ErrorAPIResponse "Internal Server Error"
// @Router /api/auth/validate-otp [post]
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
		fmt.Println("VALIDATION ERROR: ", errWhileValidations)
		validationErrors, _ := validations.FormatValidationErrors(errWhileValidations)
		ctx.IndentedJSON(http.StatusBadRequest, validationErrors)
		return
	}

	accessToken, errWhileOtpValidation := controller.service.ValidateUserOtp(ctx.Request.Context(), bffValidateUserOtpRequest)
	if errWhileOtpValidation != nil {
		fmt.Printf("************%s****************\n", errWhileOtpValidation)
		// if errors.Is(errWhileOtpValidation, commons.UserNotFoundError) { //errors.New(constants.ErrUserNotFound)
		if strings.Contains(errWhileOtpValidation.Error(), constants.ErrUserNotFound) {
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

		// if errors.Is(errWhileOtpValidation, commons.IncorrectOTPError) { //errors.New(constants.ErrIncorrectOtp)
		if strings.Contains(errWhileOtpValidation.Error(), constants.ErrOtpsMismatch) {
			errorIncorrectOtpResponse := genericModels.ErrorAPIResponse{
				Message: genericModels.ErrorMessage{
					Key:          commons.Otp,
					ErrorMessage: constants.ErrIncorrectOtp,
				},
				Error: constants.ErrAuthenticationFailed,
			}
			ctx.IndentedJSON(http.StatusUnauthorized, errorIncorrectOtpResponse)
			return
		}

		// if errors.Is(errWhileOtpValidation, commons.OtpExpiredError) { //errors.New(constants.ErrExpiredOtp)
		if strings.Contains(errWhileOtpValidation.Error(), constants.ErrExpiredOtp) {
			errorExpiredOtpResponse := genericModels.ErrorAPIResponse{
				Message: genericModels.ErrorMessage{
					Key:          commons.Otp,
					ErrorMessage: constants.ErrExpiredOtp,
				},
				Error: constants.ErrAuthenticationFailed,
			}
			ctx.IndentedJSON(http.StatusUnauthorized, errorExpiredOtpResponse)
			return
		}

		ctx.IndentedJSON(http.StatusInternalServerError, genericModels.ErrorAPIResponse{
			Error: genericConstants.ErrInternalServer,
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, models.BFFValidateUserOtpResponse{
		Message:     constants.OtpValidatedSuccessMsg,
		AccessToken: accessToken,
	})
}
