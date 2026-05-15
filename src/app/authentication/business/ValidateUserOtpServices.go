package business

import (
	"authentication/commons"
	"authentication/models"
	"authentication/repository"
	"context"
	"fmt"
	"stock_broker_application/src/utils"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ValidateUserOtpService struct {
	repository  repository.ValidateUserOtpRepository
	redisClient *redis.Client
}

func NewValidateUserOtpService(repository repository.ValidateUserOtpRepository, redisClient *redis.Client) *ValidateUserOtpService {
	return &ValidateUserOtpService{
		repository:  repository,
		redisClient: redisClient,
	}
}

// this function takes userRequest, fetches the user from db(via repository), performs all otp validations and returns error/ nil
func (service *ValidateUserOtpService) ValidateUserOtp(ctx context.Context, bffValidateUserOtpRequest models.BFFValidateUserOtpRequest) (string, error) {
	userFromDB, err := service.repository.GetUserByUsername(ctx, bffValidateUserOtpRequest.Username)
	if err != nil {
		// if errors.Is(err, gorm.ErrRecordNotFound) {
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return "", commons.UserNotFoundError //errors.New(constants.ErrUserNotFound)
		}
		return "", err
	}

	if !utils.CompareUserRequestOTP(userFromDB.OtpSent, bffValidateUserOtpRequest.Otp) {
		return "", commons.IncorrectOTPError //errors.New(constants.ErrIncorrectOtp)
	}

	if !utils.CheckOtpExpiry(userFromDB.OtpExpiresAt, time.Now()) {
		return "", commons.OtpExpiredError //errors.New(constants.ErrExpiredOtp)
	}

	accessToken, _, jti, err := utils.GenerateToken(bffValidateUserOtpRequest.Username, bffValidateUserOtpRequest.DeviceType)
	if err != nil {
		return "", err
	}

	//created a key for storing the session and the JTI for tht session
	sessionKey := fmt.Sprintf("session:%s:%s", bffValidateUserOtpRequest.Username, bffValidateUserOtpRequest.DeviceType)

	//saving/ updating in redis
	err = service.redisClient.Set(ctx, sessionKey, jti, 24*time.Hour).Err()
	if err != nil {
		return "", fmt.Errorf("failed to register session: %v", err)
	}
	return accessToken, nil
}
