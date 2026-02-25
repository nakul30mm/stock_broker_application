package business

import (
	"authentication/commons"
	"authentication/models"
	"authentication/repository"
	"context"
	"fmt"
	"stock_broker_application/src/utils"
	"time"
)

type SignInService struct {
	signinRepository repository.SignInRepository
}

func NewSignInService(signinRepository repository.SignInRepository) *SignInService {
	return &SignInService{
		signinRepository: signinRepository,
	}
}

func (service *SignInService) SignIn(ctx context.Context, spanCtx context.Context, bffSignInRequest models.BFFSignInRequest) error {
	postgresClinet := utils.GetPostgresClient().GormDB

	userFromDB, errGetUserFromDB := service.signinRepository.GetUserByUsername(spanCtx, postgresClinet, bffSignInRequest.Username)
	if errGetUserFromDB != nil {
		return commons.UserNotFoundError
	}

	//mock otp testing by updating the otpSent and otpExpiresAt fields in db table when signed in for validation task
	otp := uint64(1234)
	expiry := time.Now().Add(2 * time.Minute)

	postgresClinet.Model(&userFromDB).Updates(map[string]interface{}{
		"otpSent":      otp,
		"otpExpiresAt": expiry,
	})
	fmt.Println("mock otp generated: ", otp)

	if !utils.CompareHashPassword(userFromDB.Password, bffSignInRequest.Password) {
		return commons.IncorrectPasswordError
	}
	return nil
}
