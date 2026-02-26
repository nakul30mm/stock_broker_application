package business

import (
	"authentication/commons"
	"authentication/models"
	"authentication/repository"
	"context"
	"stock_broker_application/src/utils"
	"time"

	"gorm.io/gorm"
)

type ValidateUserOtpService struct {
	repository repository.ValidateUserOtpRepository
	db         *gorm.DB
}

func NewValidateUserOtpService(repository repository.ValidateUserOtpRepository, db *gorm.DB) *ValidateUserOtpService {
	return &ValidateUserOtpService{
		repository: repository,
		db:         db,
	}
}

func NewValidateUserOtpServiceForTest(mockRepo repository.ValidateUserOtpRepository, db *gorm.DB) *ValidateUserOtpService {
	return &ValidateUserOtpService{
		repository: mockRepo,
		db:         db,
	}
}

// this function takes userRequest, fetches the user from db(via repository), performs all otp validations and returns error/ nil
func (service *ValidateUserOtpService) ValidateUserOtp(ctx context.Context, spanCtx context.Context, bffValidateUserOtpRequest models.BFFValidateUserOtpRequest) error {
	// postgresClinet := utils.GetPostgresClient().GormDB
	userFromDB, errGettingUserFromDB := service.repository.GetUserByUsername(spanCtx, service.db, bffValidateUserOtpRequest.Username)
	if errGettingUserFromDB != nil {
		return commons.UserNotFoundError
	}

	if !utils.CompareUserRequestOTP(userFromDB.OtpSent, bffValidateUserOtpRequest.Otp) {
		return commons.IncorrectOTPError
	}

	if !utils.CheckOtpExpiry(userFromDB.OtpExpiresAt, time.Now()) {
		return commons.OtpExpiredError
	}
	return nil
}
