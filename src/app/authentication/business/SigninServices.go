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

	user, err := service.signinRepository.GetUserByUsername(spanCtx, tx, bffSignInRequest.Username)
	if err != nil {
		return commons.ErrUserNotFound
	}

	if !utils.CompareHashPassword(user.Password, bffSignInRequest.Password) {
		return commons.ErrIncorrectPassword
	}
	return nil
}
