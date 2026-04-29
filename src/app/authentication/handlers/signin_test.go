package handlers_test

import (
	"authentication/business"
	"authentication/commons/constants"
	"authentication/handlers"
	"authentication/models"
	"authentication/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"stock_broker_application/src/utils"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// embedding all the requirements in this struct
type SignInTestSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	db   *gorm.DB
	r    *gin.Engine
}

func (s *SignInTestSuite) SetupTest() {
	//initializing mock DB
	sqlDB, mock, _ := sqlmock.New()
	s.mock = mock
	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	s.db, _ = gorm.Open(dialector, &gorm.Config{})

	//initializing layers
	repo := repository.NewSignInRepository(s.db)
	svc := business.NewSignInService(repo)
	handler := handlers.NewSignInHandler(svc)

	//settingup the router
	gin.SetMode(gin.TestMode)
	s.r = gin.New()
	s.r.POST(constants.SigninTest, handler.HandleSignIn)
}

func (s *SignInTestSuite) TestMockHandleSignIn200UserLoggedInSuccessfully() {
	username := "Nakul7500"
	password := "Admin@123"
	hashedPassword, _ := utils.HashPassword(password)

	columns := []string{"id", "username", "password", "pan_card", "phone_number", "email", "otp_sent", "otp_expires_at"}
	rows := sqlmock.NewRows(columns).AddRow(1, username, hashedPassword, "ABCDE1234F", 9876543210, "test@example.com", 0, 0)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.QueryUserByEmail)).
		WithArgs(username, 1).
		WillReturnRows(rows)

	signInReq := models.BFFSignInRequest{
		Username:   username,
		Password:   password,
		DeviceType: "web",
	}
	jsonReq, _ := json.Marshal(signInReq) //marhalling because our handler expects a json body and has the binding validations,
	// if we dont marshall the request body into json, it won't be able to extract the json tags we've defined in our request body struct

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewBuffer(jsonReq)) //newbuffer because it accepts a byte slice and json.marshal returns a byte slice,
	// instead if we would've used newbufferstring, it acceps a string input so we would have needed to convert our json.marshal's byte slice response to string,
	// so newbuffer eliminates that step
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.NoError(s.mock.ExpectationsWereMet()) //check if the expectations and the assertions defined above are all correctly met and are not failed, if failed, it returns an error
}

func (s *SignInTestSuite) TestMockHandleSignIn400InvalidPayloadBindingErr() {
	// username := "Nakul7500"
	// password := "Admin@123"

	// signInReq := models.BFFSignInRequest{
	// 	Username: username,
	// 	Password: password,
	// }

	// jsonReq, _ := json.Marshal(signInReq)

	// req, _ := http.NewRequest(http.MethodPost, constants.SigninTest, bytes.NewBuffer(jsonReq))

	req, _ := http.NewRequest(http.MethodPost, constants.SigninTest, bytes.NewBufferString(`{"username": 123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "invalid required payload")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignInTestSuite) TestMockHandleSignIn400InvalidPayloadValidationErr() {
	signInReq := models.BFFSignInRequest{
		Username:   "Nakul",
		Password:   "123",
		DeviceType: "invalid",
	}
	jsonReq, _ := json.Marshal(signInReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(strings.ToLower(w.Body.String()), "password")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignInTestSuite) TestMockHandleSignIn401UnauthorizedWrongPassword() {
	username := "Nakul7500"
	correctPassword := "Admin@123"
	hashedPassword, _ := utils.HashPassword(correctPassword)
	wrongPassword := "WrongPass@123"

	rows := sqlmock.NewRows([]string{"id", "username", "password"}).
		AddRow(1, username, hashedPassword)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.QueryUserByEmail)).
		WithArgs(username, 1).
		WillReturnRows(rows)

	signInReq := models.BFFSignInRequest{
		Username:   username,
		Password:   wrongPassword,
		DeviceType: "web",
	}
	jsonReq, _ := json.Marshal(signInReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
	// s.Contains(w.Body.String(), "incorrect password")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignInTestSuite) TestMockHandleSignIn404UserNotFound() {
	username := "nakul1122"

	columns := []string{"id", "username", "password"}
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.QueryUserByEmail)).
		WithArgs(username, 1).
		WillReturnRows(sqlmock.NewRows(columns)) //correct columns but zero rows of data

	signInReq := models.BFFSignInRequest{
		Username:   username,
		Password:   "Admin@123",
		DeviceType: "web",
	}
	jsonReq, _ := json.Marshal(signInReq)

	req, _ := http.NewRequest(http.MethodPost, "/api/auth/signin", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "user not found")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignInTestSuite) TestMockHandleSignIn500InternalServerError() {
	username := "Nakul7500"
	password := "Admin@123"
	hashedPassword, _ := utils.HashPassword(password)

	columns := []string{"id", "username", "password", "pan_card", "phone_number", "email", "otp_sent", "otp_expires_at"}
	rows := sqlmock.NewRows(columns).AddRow(1, username, hashedPassword, "ABCDE1234F", 9876543210, "test@example.com", 0, 0)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.IncorrectQueryUserByEmail)).WithArgs(username, 1).WillReturnRows(rows)

	signInReq := models.BFFSignInRequest{
		Username:   username,
		Password:   password,
		DeviceType: "web",
	}

	jsonReq, _ := json.Marshal(signInReq)

	req, _ := http.NewRequest(http.MethodPost, constants.SigninTest, bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "internal server error")
}

func TestSignInTestSuite(t *testing.T) {
	suite.Run(t, new(SignInTestSuite))
}
