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
	// router *gin.Engine
}

func (s *SignInTestSuite) SetupTest() {
	//initializing mock DB
	sqlDB, mock, _ := sqlmock.New()
	s.mock = mock
	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	s.db, _ = gorm.Open(dialector, &gorm.Config{})

	// //initializing layers
	// repo := repository.NewSignInRepository(s.db)
	// svc := business.NewSignInService(repo)
	// handler := handlers.NewSignInHandler(svc)

	// //settingup the router
	// gin.SetMode(gin.TestMode)
	// s.router = gin.New()
	// s.router.POST(constants.SigninTest, handler.HandleSignIn)
	// gin.SetMode(gin.TestMode)
}

func (s *SignInTestSuite) getSignInRouter() *gin.Engine {

	gin.SetMode(gin.TestMode)
	router := gin.New()

	//initializing layers
	repo := repository.NewSignInRepository(s.db)
	svc := business.NewSignInService(repo)
	handler := handlers.NewSignInHandler(svc)

	//settingup the router

	router.POST(constants.SigninTest, handler.HandleSignIn)
	return router
}

var (
	signInUsername    = "nakul7500"
	signInPassword    = "Admin@123"
	hashedPassword, _ = utils.HashPassword(signInPassword)
	DeviceType        = "web"
)

var signInReq = models.BFFSignInRequest{
	Username:   signInUsername,
	Password:   signInPassword,
	DeviceType: DeviceType,
}

var signInjsonReq, _ = json.Marshal(signInReq) //marhalling because our handler expects a json body and has the binding validations,
// if we dont marshal the request body into json, it won't be able to extract the json tags we've defined in our request body struct

func (s *SignInTestSuite) TestMockHandleSignIn200UserLoggedInSuccessfully() {

	columns := []string{"id", "username", "password", "pan_card", "phone_number", "email", "otp_sent", "otp_expires_at"}
	rows := sqlmock.NewRows(columns).AddRow(1, signInUsername, hashedPassword, "ABCDE1234F", 9876543210, "test@example.com", 0, 0)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.UserByEmailTestQuery)).
		WithArgs(signInUsername, 1).
		WillReturnRows(rows)

	w := s.executeRequest(http.MethodPost, constants.SigninTest, signInjsonReq)

	s.Equal(http.StatusOK, w.Code)
	s.NoError(s.mock.ExpectationsWereMet()) //check if the expectations(related to db, i.e. begin, commit, call, etc) defined above are all correctly met and are not failed, if failed, it returns an error
}

func (s *SignInTestSuite) TestMockHandleSignIn400InvalidPayloadBindingErr() {

	req, _ := http.NewRequest(http.MethodPost, constants.SigninTest, bytes.NewBufferString(`{"username": 123}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := s.getSignInRouter()
	router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(w.Body.String(), "invalid required payload")
}

func (s *SignInTestSuite) TestMockHandleSignIn400InvalidPayloadValidationErr() {

	signInReq := models.BFFSignInRequest{
		Username:   "Nakul",
		Password:   "123",
		DeviceType: "invalid",
	}
	jsonReq, _ := json.Marshal(signInReq)

	w := s.executeRequest(http.MethodPost, constants.SigninTest, jsonReq)

	s.Equal(http.StatusBadRequest, w.Code)
	s.Contains(strings.ToLower(w.Body.String()), "password")
}

func (s *SignInTestSuite) TestMockHandleSignIn401UnauthorizedWrongPassword() {

	wrongPassword := "WrongPass@123"

	rows := sqlmock.NewRows([]string{"id", "username", "password"}).
		AddRow(1, signInUsername, hashedPassword)

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.UserByEmailTestQuery)).
		WithArgs(signInUsername, 1).
		WillReturnRows(rows)

	signInReqIncorrectPassword := models.BFFSignInRequest{
		Username:   signInUsername,
		Password:   wrongPassword,
		DeviceType: DeviceType,
	}

	jsonReq, _ := json.Marshal(signInReqIncorrectPassword)

	w := s.executeRequest(http.MethodPost, constants.SigninTest, jsonReq)

	s.Equal(http.StatusUnauthorized, w.Code)
	s.Contains(w.Body.String(), "entered password is not correct")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignInTestSuite) TestMockHandleSignIn404UserNotFound() {

	// username := "nakul1122"

	columns := []string{"id", "username", "password"}
	s.mock.ExpectQuery(regexp.QuoteMeta(constants.UserByEmailTestQuery)).
		WithArgs(signInUsername, 1).
		WillReturnRows(sqlmock.NewRows(columns)) //correct columns but zero rows of data or we can also return error directly

	w := s.executeRequest(http.MethodPost, constants.SigninTest, signInjsonReq)

	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "user not found")
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *SignInTestSuite) TestMockHandleSignIn500InternalServerError() {

	s.mock.ExpectQuery(regexp.QuoteMeta(constants.UserByEmailTestQuery)).
		WithArgs(signInUsername, 1).
		WillReturnError(errors.New("db connection error"))

	w := s.executeRequest(http.MethodPost, constants.SigninTest, signInjsonReq)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "internal server error")
	s.NoError(s.mock.ExpectationsWereMet())
}

func TestSignInTestSuite(t *testing.T) {
	suite.Run(t, new(SignInTestSuite))
}

// func getMockSqlDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
// 	sqlDB, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	t.Cleanup(func() { _ = sqlDB.Close() })

// 	dialector := postgres.New(postgres.Config{Conn: sqlDB})
// 	gdb, err := gorm.Open(dialector, &gorm.Config{})
// 	require.NoError(t, err)

// 	return gdb, mock
// }

// func newSigninRouter(gdb *gorm.DB) *gin.Engine {
// 	gin.SetMode(gin.TestMode)
// 	router := gin.New()

// 	repo := repository.NewSignInRepository(gdb)
// 	svc := business.NewSignInService(repo)
// 	handler := handlers.NewSignInHandler(svc)

// 	router.POST(constants.SigninTest, handler.HandleSignIn)
// 	return router
// }

// type SignInTestSuite struct {
// 	suite.Suite
// }

// func (suite *SignInTestSuite) SetupTest() {
// 	gin.SetMode(gin.TestMode)
// }

func (s *SignInTestSuite) executeRequest(httpMethod string, routeUrl string, reqBody []byte) *httptest.ResponseRecorder {
	// jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest(httpMethod, routeUrl, bytes.NewBuffer(reqBody)) //newbuffer because it accepts a byte slice and json.marshal returns a byte slice,
	// instead if we would've used newbufferstring, it acceps a string input so we would have needed to convert our json.marshal's byte slice response to string,
	// so newbuffer eliminates that step

	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router := s.getSignInRouter()
	router.ServeHTTP(w, req)
	return w
}

func (s *SignInTestSuite) mockUserLookup(username string, userRow *sqlmock.Rows) {
	query := regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)

	s.mock.ExpectQuery(query).
		WithArgs(username, 1).
		WillReturnRows(userRow)
}
