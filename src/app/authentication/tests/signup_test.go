package tests

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
	// router *gin.Engine
}

func (s *SignUpTestSuite) SetupTest() {
	//initializing mock DB
	sqlDB, mock, _ := sqlmock.New()
	s.mock = mock
	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	s.db, _ = gorm.Open(dialector, &gorm.Config{})

	// //initializing layers
	// repo := repository.NewCreateUserRepository(s.db)
	// svc := business.NewCreateUserService(repo, s.db)
	// handler := handlers.NewCreateUserHandler(svc)

	// //settingup the router
	// gin.SetMode(gin.TestMode)
	// s.router = gin.New()
	// s.router.POST(constants.SignupTest, handler.HandleCreaterUser)
}

var (
	signUpUsername = "Nakul7500"
	signUpPassword = "Admin@123"
	email          = "test@example.com"
	pan            = "ABCDE1234P"
	phone          = uint64(9876543210)
)

var SignupReq = models.BFFCreateUserRequest{
	Username:        signUpUsername,
	Password:        signUpPassword,
	ConfirmPassword: signUpPassword,
	PanCard:         pan,
	PhoneNumber:     phone,
	Email:           email,
}

var jsonSignUpReq, _ = json.Marshal(SignupReq)
var arguements = `username, sqlmock.AnyArg(), pan, phone, email`

func (s *SignUpTestSuite) getSignUpRouter() *gin.Engine {
	router := gin.New()
	//initializing layers
	repo := repository.NewCreateUserRepository(s.db)
	svc := business.NewCreateUserService(repo, s.db)
	handler := handlers.NewCreateUserHandler(svc)

	//settingup the router
	gin.SetMode(gin.TestMode)
	router = gin.New()
	router.POST(constants.SignupTest, handler.HandleCreaterUser)

	return router
}

// test case for - successful signup
func (s *SignUpTestSuite) TestMockHandleSignup201UserCreatedSuccessfully() {

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "otpSent", "otpExpiresAt"}).
			AddRow(1, nil, nil))

		//TODO: try sending an hashed password instead of sqlmock.AnyArg
		//! tried - but the test didn't pass,
		//* because when a plain password is sent, the hash generates a different pasword everytime,
		//* which will never match with the one in the db, so for this reason (when we don't know what will be the value),
		//* we send AnyArg() because it matches the arguement type
	s.mock.ExpectCommit()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusCreated, w.Code)
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - user already exists
func (s *SignUpTestSuite) TestMockHandleSignup409UserAlreadyExists() {

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("already exists"))
	s.mock.ExpectRollback()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusConflict, w.Code)
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - validation error for the request body
func (s *SignUpTestSuite) TestMockHandleSignUp400InvalidPayloadValidationError() {

	signupReq := models.BFFCreateUserRequest{
		Username:        signUpUsername,
		Password:        signUpPassword,
		ConfirmPassword: "DifferentP@$$w0rd",
		PanCard:         pan,
		PhoneNumber:     uint64(phone),
		Email:           email,
	}
	jsonRequest, _ := json.Marshal(signupReq)

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonRequest)

	s.Equal(http.StatusBadRequest, w.Code)
	strings.Contains(w.Body.String(), constants.ErrInvalidPayload)

}

// test case for - binding error for the request body
func (s *SignUpTestSuite) TestMockHandleSignUp400InvalidPayloadBindingError() {

	req, _ := http.NewRequest(http.MethodPost, constants.SignupTest, bytes.NewBufferString(`{"username": 123}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := s.getSignUpRouter()
	router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	strings.Contains(w.Body.String(), constants.ErrInvalidPayload)

}

// test case for - internal server error due to db connection error
func (s *SignUpTestSuite) TestMockHandleSignup500InternalServerError() {

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("db connection error"))
	s.mock.ExpectRollback()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusInternalServerError, w.Code)
	strings.Contains(w.Body.String(), "internal server error")
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - the pan number entered by the user in the request already exists
func (s *SignUpTestSuite) TestMockHandleSignUp400BadRequestDuplicatePanNo() {

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("duplicate key value violates unique constraint, idx_users_pan_card"))
	s.mock.ExpectRollback()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusConflict, w.Code)
	strings.Contains(w.Body.String(), constants.IndexUsersPanCard)
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - the email given by the user in the request already exists
func (s *SignUpTestSuite) TestMockHandleSignUp400BadRequestDuplicateEmail() {
	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("duplicate key value violates unique constraint, idx_users_email"))
	s.mock.ExpectRollback()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusConflict, w.Code)
	strings.Contains(w.Body.String(), constants.IndexUsersEmail)
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - the username user is trying to give in the request already exists
func (s *SignUpTestSuite) TestMockHandleSignUp400BadRequestDuplicateUsername() {

	s.mock.ExpectBegin()
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnError(errors.New("duplicate key value violates unique constraint, username already exists"))
	s.mock.ExpectRollback()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusConflict, w.Code)
	strings.Contains(w.Body.String(), constants.ErrUsernameExists)
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - error while transaction begin
func (s *SignUpTestSuite) TestMockHandleSignUp500TransactionBeginError() {

	s.mock.ExpectBegin().WillReturnError(errors.New("failed to begin database transaction"))

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.NoError(s.mock.ExpectationsWereMet())
}

// test case for - error while transaction commit
func (s *SignUpTestSuite) TestMockHandleSignUp500TransactionCommitError() {
	s.mock.ExpectBegin()

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.CreateUserTestQuery)).
		WithArgs(signUpUsername, sqlmock.AnyArg(), pan, phone, email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "otpSent", "otpExpiresAt"}).
			AddRow(0, nil, nil))

	s.mock.ExpectCommit().WillReturnError(errors.New("failed to commit database transaction"))
	// s.mock.ExpectRollback()

	w := s.executeRequest(http.MethodPost, constants.SignupTest, jsonSignUpReq)

	s.Equal(http.StatusInternalServerError, w.Code)
	strings.Contains(w.Body.String(), "failed to commit database transaction")
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
//*DONE

func (s *SignUpTestSuite) executeRequest(httpMethod string, routeUrl string, reqBody []byte) *httptest.ResponseRecorder {
	// jsonReq, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(httpMethod, routeUrl, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := s.getSignUpRouter()
	router.ServeHTTP(w, req)
	return w
}

// func (s *SignUpTestSuite) mockUserLookup(query string, arguements string, userRow *sqlmock.Rows) {

// 	s.mock.ExpectQuery(query).
// 		WithArgs(arguements).
// 		WillReturnRows(userRow)
// }
