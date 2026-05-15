package tests

import (
	"authentication/business"
	"authentication/commons"
	"authentication/commons/constants"
	"authentication/handlers"
	"authentication/models"
	"authentication/repository"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	genericConstants "stock_broker_application/src/constants"
	"stock_broker_application/src/utils"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ValidatOtpTestSuite struct {
	suite.Suite
	mock            sqlmock.Sqlmock
	db              *gorm.DB
	redisClient     *redis.Client
	mockRedisClient redismock.ClientMock
}

var (
	ValidateOTPUsername       = "Nakul"
	ValidateOTPPassword       = "Admin@123"
	fixedOTPExp         int64 = 1778438400
)

var signUpReq = models.BFFValidateUserOtpRequest{
	Username:   ValidateOTPUsername,
	Otp:        "1008",
	DeviceType: "web",
}

var jsonValidateOTPReq, _ = json.Marshal(signUpReq)

func (s *ValidatOtpTestSuite) SetupTest() {
	sqlDB, mock, _ := sqlmock.New()
	s.mock = mock

	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	s.db, _ = gorm.Open(dialector, &gorm.Config{})

	s.redisClient, s.mockRedisClient, _ = utils.GetRedisClient(ctx, false)

	if err := utils.InitJWTConfig("../../../config"); err != nil {
		log.Fatalf(genericConstants.ErrJWTConfigReadFailed, err)
	}
}

func (s *ValidatOtpTestSuite) getValidateOTPRouter() *gin.Engine {

	gin.SetMode(gin.TestMode)
	router := gin.New()

	repo := repository.NewValidateUserOtpRepository(s.db)
	svc := business.NewValidateUserOtpService(repo, s.redisClient)
	handler := handlers.NewValidateUserOtpHandler(svc)

	router.POST(constants.ValidateOTPTest, func(ctx *gin.Context) {
		ctx.Set(genericConstants.Token, token)
		ctx.Set(commons.TokenExpiry, time.Unix(fixedOTPExp, 0))
		ctx.Next()
	}, handler.HandleValidateUserOtp)

	return router
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP200Successful() {

	otpExpiry := time.Now().Add(5 * time.Minute).Unix()
	columns := []string{"id", "username", "password", "panCard", "phoneNumber", "email", "otpSent", "otpExpiresAt"}
	rows := sqlmock.NewRows(columns).AddRow(1, ValidateOTPUsername, hashedPassword, "ABCDE1234F", 9876543210, "nakul@example.com", 1008, otpExpiry)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.ValidateUserOTPTestQuery)).
		WithArgs(ValidateOTPUsername, 1).
		WillReturnRows(rows)

	sessionKey := fmt.Sprintf("session:%s:%s", ValidateOTPUsername, "web")
	s.mockRedisClient.ExpectSet(sessionKey, nil, 24*time.Hour).SetVal("OK")

	w := s.executeRequest(http.MethodPost, constants.ValidateOTPTest, jsonValidateOTPReq)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "OTP validated successfully")
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP400UserNotFound() {

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.ValidateUserOTPTestQuery)).
		WithArgs(ValidateOTPUsername, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	w := s.executeRequest(http.MethodPost, constants.ValidateOTPTest, jsonValidateOTPReq)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), constants.ErrUserNotFound)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP400BindingError() {

	req, _ := http.NewRequest(http.MethodPost, constants.ValidateOTPTest, bytes.NewBufferString(`{"Username": 123, "otp": 1122, "DeviceType":"web"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := s.getValidateOTPRouter()
	router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	strings.Contains(w.Body.String(), constants.ErrInvalidPayload)
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP400ValidationError() {
	invalidReq := models.BFFValidateUserOtpRequest{
		Username:   "1234",
		Otp:        "1008",
		DeviceType: "web",
	}

	invalidJsonReq, _ := json.Marshal(invalidReq)

	// req, _ := http.NewRequest(http.MethodPost, constants.ValidateOTPTest, bytes.NewBuffer(invalidJsonReq))
	w := s.executeRequest(http.MethodPost, constants.ValidateOTPTest, invalidJsonReq)
	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "invalid value")
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP401IncorrectOTP() {
	columns := []string{"id", "username", "password", "panCard", "phoneNumber", "email", "otpSent", "otpExpiresAt"}
	rows := sqlmock.NewRows(columns).AddRow(1, ValidateOTPUsername, hashedPassword, "ABCDE1234F", 9876543210, "nakul@example.com", 1800, 0)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.ValidateUserOTPTestQuery)).WithArgs(ValidateOTPUsername, 1).WillReturnRows(rows)

	w := s.executeRequest(http.MethodPost, constants.ValidateOTPTest, jsonValidateOTPReq)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "entered OTP is not correct")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP401ExpiredOTP() {
	columns := []string{"id", "username", "password", "panCard", "phoneNumber", "email", "otpSent", "otpExpiresAt"}
	rows := sqlmock.NewRows(columns).AddRow(1, ValidateOTPUsername, hashedPassword, "ABCDE1234F", 9876543210, "nakul@example.com", 1008, 0)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.ValidateUserOTPTestQuery)).
		WithArgs(ValidateOTPUsername, 1).
		WillReturnRows(rows)

	w := s.executeRequest(http.MethodPost, constants.ValidateOTPTest, jsonValidateOTPReq)
	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), constants.ErrExpiredOtp)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *ValidatOtpTestSuite) TestMockValidateUserOTP500InternalServerError() {
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.ValidateUserOTPTestQuery)).
		WithArgs(ValidateOTPUsername, 1).
		WillReturnError(errors.New("db connection error"))

	w := s.executeRequest(http.MethodPost, constants.ValidateOTPTest, jsonValidateOTPReq)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), genericConstants.ErrInternalServer)
}

func (s *ValidatOtpTestSuite) executeRequest(httpMethod string, routeUrl string, reqBody []byte) *httptest.ResponseRecorder {
	// jsonReq, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(httpMethod, routeUrl, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := s.getValidateOTPRouter()
	router.ServeHTTP(w, req)
	return w
}

func TestValidateOTPTestSuite(t *testing.T) {
	suite.Run(t, new(ValidatOtpTestSuite))
}
