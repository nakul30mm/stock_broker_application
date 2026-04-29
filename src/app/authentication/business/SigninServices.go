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

type SignInService struct {
	signinRepository repository.SignInRepository
}

func NewSignInService(signinRepository repository.SignInRepository) *SignInService {
	return &SignInService{
		signinRepository: signinRepository,
	}
}

func (service *SignInService) SignIn(spanCtx context.Context, bffSignInRequest models.BFFSignInRequest) error {

	userFromDB, err := service.signinRepository.GetUserByUsername(spanCtx, bffSignInRequest.Username)
	if err != nil {
		fmt.Println("error from service:", err)
		return err
	}

	if !utils.CompareHashPassword(userFromDB.Password, bffSignInRequest.Password) {
		fmt.Println("incorrect password")
		return errors.New(constants.ErrIncorrectPassword)
	}
	return nil
}
