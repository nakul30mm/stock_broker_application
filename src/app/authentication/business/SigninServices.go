package business

import (
	"authentication/commons"
	"authentication/models"
	"authentication/repository"
	"context"
	"fmt"
	genericErrors "stock_broker_application/src/constants"
	"stock_broker_application/src/utils"
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
	postgresClinet := utils.GetPostgresClient()
	tx := postgresClinet.GormDB.Begin()

	if tx.Error != nil {
		return fmt.Errorf(genericErrors.ErrBeginTx, tx.Error)
	}

	user, errGetUserFromDB := service.signinRepository.GetUserByUsername(spanCtx, tx, bffSignInRequest.Username)
	if errGetUserFromDB != nil {
		return commons.UserNotFoundError
	}

	if !utils.CompareHashPassword(user.Password, bffSignInRequest.Password) {
		return commons.IncorrectPasswordError
	}
	return nil
}
