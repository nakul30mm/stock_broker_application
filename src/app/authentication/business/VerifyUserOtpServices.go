package business

import (
	"authentication/commons"
	"authentication/models"
	"authentication/repository"
	"context"
	"stock_broker_application/src/utils"
	"time"
)

type ValidateUserOtpService struct {
	repository repository.ValidateUserOtpRepository
}

func NewValidateUserOtpService(repositry repository.ValidateUserOtpRepository) *ValidateUserOtpService {
	return &ValidateUserOtpService{
		repository: repositry,
	}
}

func (service *ValidateUserOtpService) ValidateUserOtp(ctx context.Context, spanCtx context.Context, bffValidateUserOtpRequest models.BFFValidateUserOtpRequest) error {
	postgresClinet := utils.GetPostgresClient().GormDB

	userFromDB, errGettingUserFromDB := service.repository.GetUserByUsername(spanCtx, postgresClinet, bffValidateUserOtpRequest.Username)
	if errGettingUserFromDB != nil {
		return commons.UserNotFoundError
	}

	if !utils.CheckOtpExpiry(userFromDB.OtpExpiresAt, time.Now()) {
		return commons.OtpExpired
	}
	if !utils.CompareUserRequestOTP(userFromDB.OtpSent, bffValidateUserOtpRequest.Otp) {
		return commons.IncorrectOTPError
	}
	return nil
}
