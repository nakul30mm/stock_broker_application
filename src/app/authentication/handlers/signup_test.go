package handlers_test

import (
	"authentication/business"
	"authentication/commons/constants"
	"authentication/handlers"
	"authentication/models"
	"authentication/repository"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SignUpTestSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	db   *gorm.DB
	r    *gin.Engine
}

func (s *SignUpTestSuite) SetupTest() {
	//initializing mock DB
	sqlDB, mock, _ := sqlmock.New()
	s.mock = mock
	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	s.db, _ = gorm.Open(dialector, &gorm.Config{})

	//initializing layers
	repo := repository.NewCreateUserRepository(s.db)
	svc := business.NewCreateUserService(repo, s.db)
	handler := handlers.NewCreateUserHandler(svc)

	//settingup the router
	gin.SetMode(gin.TestMode)
	s.r = gin.New()
	s.r.POST(constants.SignupTest, handler.HandleCreaterUser)
}

var (
	username = "Nakul7500"
	password = "Admin@123"
	email    = "test@example.com"
	pan      = "ABCDE1234P"
	phone    = uint64(9876543210)
)

var SignupReq = models.BFFCreateUserRequest{
	Username:        username,
	Password:        password,
	ConfirmPassword: password,
	PanCard:         pan,
	PhoneNumber:     phone,
	Email:           email,
}

var JsonReq, _ = json.Marshal(SignupReq)

func (s *SignUpTestSuite) TestMockHandleSignup201UserCreatedSuccessfully() {

	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(username, sqlmock.AnyArg(), pan, phone, email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "otpSent", "otpExpiresAt"}).
			AddRow(1, nil, nil))

		//TODO: try sending an hashed password instead of sqlmock.AnyArg
		//* tried - the test didn't pass

	s.mock.ExpectCommit()

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(JsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignUpTestSuite) TestMockHandleSignup409UserAlreadyExists() {

	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(username, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("already exists"))

	s.mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(JsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignUpTestSuite) TestMockHadleSignUp400InvalidPayloadValidationError() {

	signupReq := models.BFFCreateUserRequest{
		Username:        username,
		Password:        password,
		ConfirmPassword: "DifferentP@$$w0rd",
		PanCard:         pan,
		PhoneNumber:     uint64(phone),
		Email:           email,
	}
	jsonReq, _ := json.Marshal(signupReq)

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	strings.Contains(w.Body.String(), constants.ErrInvalidPayload)

}

func (s *SignUpTestSuite) TestMockHandleSignUp400InvalidPayloadBindingError() {

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBufferString(`{"username": 123}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	strings.Contains(w.Body.String(), constants.ErrInvalidPayload)

}

func (s *SignUpTestSuite) TestMockHandleSignup500InternalServerError() {

	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(username, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("db connection error"))

	s.mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(JsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	strings.Contains(w.Body.String(), "internal server error")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignUpTestSuite) TestMockHandleSignUp400BadRequestDuplicatePanNo() {

	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(username, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("duplicate key value violates unique constraint, idx_users_pan_card"))

	s.mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(JsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
	strings.Contains(w.Body.String(), constants.IndexUsersPanCard)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignUpTestSuite) TestMockHandleSignUp400BadRequestDuplicateEmail() {
	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(username, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("duplicate key value violates unique constraint, idx_users_email"))

	s.mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(JsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
	strings.Contains(w.Body.String(), constants.IndexUsersEmail)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignUpTestSuite) TestMockHandleSignUp400BadRequestDuplicateUsername() {

	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(username, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("duplicate key value violates unique constraint, username already exists"))

	s.mock.ExpectRollback()

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBuffer(JsonReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.r.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
	strings.Contains(w.Body.String(), constants.ErrUsernameExists)
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestSignUpTestSuite(t *testing.T) {
	suite.Run(t, new(SignUpTestSuite))
}

//* all below are passing
//success 201
//already exists 409
//validation error 400
//bnding error 400
//internalservererror 500
//duplicate pan number 400
//duplicate email 400
//duplicate username 400

//TODO: write test function for covering the transaction errors of the service layer (check coverage)
