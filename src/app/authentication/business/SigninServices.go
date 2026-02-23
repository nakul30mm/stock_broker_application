package business

import (
	"authentication/commons/constants"
	"authentication/models"
	"authentication/repository"
	"context"
	"errors"
	"fmt"
	"stock_broker_application/src/utils"
)

type SigninUserService struct {
	signinUserRepository repository.SigninUserRepository
}

func NewSigninUserService(signinUserRepository repository.SigninUserRepository) *SigninUserService {
	return &SigninUserService{
		signinUserRepository: signinUserRepository,
	}
}

func (service *SigninUserService) SigninUser(ctx context.Context, spanCtx context.Context, bffSigninUserRequest models.BFFSigninUserRequest) error {
	postgresClinet := utils.GetPostgresClient()
	client := postgresClinet.GormDB
	user, err := service.signinUserRepository.SigninUser(spanCtx, client, bffSigninUserRequest)
	if err != nil {
		if err.Error() == constants.ErrUserNotFound {
			return errors.New(constants.ErrUserNotFound)
		}
		return err
	}

	passwordMatch := utils.CompareHashPassword(user.Password, bffSigninUserRequest.Password)
	if !passwordMatch {
		return fmt.Errorf(constants.ErrPasswordMismatch, errors.New(constants.ErrInvalidUsernamePassword))
	}
	return nil

}
