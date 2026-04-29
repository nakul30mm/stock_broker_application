package handlers_test

import (
	"authentication/models"
	"encoding/json"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type SignUpTestSuite struct {
	suite.Suite
	r    *gin.Engine
	mock *sqlmock.Sqlmock
	db   *gorm.DB
}

func (s *SignUpTestSuite) TestMockSignup200SignUpSuccess() {
	username := "Nakul7500"
	password := "Admin@123"
	panNumber := "HEQPP2233M"
	email := "nakul@gmail.com"
	phoneNumber := uint64(8103631555)

	signUpReq := models.BFFCreateUserRequest{
		Username:        username,
		Password:        password,
		ConfirmPassword: password,
		PanCard:         panNumber,
		PhoneNumber:     phoneNumber,
		Email:           email,
	}

	jsonReq, _ := json.Marshal(signUpReq)
	fmt.Println("json request", jsonReq)

}
