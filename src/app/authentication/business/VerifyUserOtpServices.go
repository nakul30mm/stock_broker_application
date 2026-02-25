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

func NewValidateUserOtpService(repository repository.ValidateUserOtpRepository) *ValidateUserOtpService {
	return &ValidateUserOtpService{
		repository: repository,
	}
}

func (service *ValidateUserOtpService) ValidateUserOtp(ctx context.Context, spanCtx context.Context, bffValidateUserOtpRequest models.BFFValidateUserOtpRequest) error {
	postgresClinet := utils.GetPostgresClient().GormDB

	userFromDB, errGettingUserFromDB := service.repository.GetUserByUsername(spanCtx, postgresClinet, bffValidateUserOtpRequest.Username)
	if errGettingUserFromDB != nil {
		return commons.UserNotFoundError
	}

	if !utils.CompareUserRequestOTP(userFromDB.OtpSent, bffValidateUserOtpRequest.Otp) {
		return commons.IncorrectOTPError
	}

	if !utils.CheckOtpExpiry(userFromDB.OtpExpiresAt, time.Now()) {
		return commons.OtpExpired
	}
	return nil
}
