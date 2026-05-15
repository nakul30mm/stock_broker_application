package tests

import (
	"authentication/business"
	"authentication/commons"
	"authentication/commons/constants"
	"authentication/handlers"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	genericConstants "stock_broker_application/src/constants"
	"stock_broker_application/src/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

var (
	token           = "consider-this-is-a-token-123"
	key             = fmt.Sprintf("BLACKLISTED_TOKEN_%s", token)
	ctx             = context.Background()
	fixedUnix int64 = 1778438400
)

func getLogoutRouter(redisClient *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	svc := business.NewLogoutService(redisClient)
	handler := handlers.NewLogoutHandler(svc)

	router.POST(constants.LogoutTest, func(ctx *gin.Context) {
		ctx.Set(genericConstants.Token, token)
		ctx.Set(commons.TokenExpiry, time.Unix(fixedUnix, 0))
		ctx.Next()
	}, handler.Logout)
	return router
}

type LogoutTestSuite struct {
	suite.Suite
}

func (l *LogoutTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
}

// test case for - successful logout
func (s *LogoutTestSuite) TestMockLogout200SuccessfullyLoggedOut() {

	// expectedTTL := time.Until(time.Unix(fixedUnix, 0))

	redisClient, mockRedisClient, _ := utils.GetRedisClient(ctx, false)
	router := getLogoutRouter(redisClient)

	mockRedisClient.ExpectSet(key, 1, 0).SetVal("OK")

	req, _ := http.NewRequest(http.MethodPost, constants.LogoutTest, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "logged out successfully")
	s.NoError(mockRedisClient.ExpectationsWereMet())
}

// test case for - nil redis client error
func (s *LogoutTestSuite) TestMockLogout500RedisClientNotInitializedErr() {

	router := getLogoutRouter(nil)
	//since the service will return immediately because of nil redis client, no expectset

	req, _ := http.NewRequest(http.MethodPost, constants.LogoutTest, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), "redis client not initialized")
	//no expectations too, because no calls to redis
}

// test case for - internal server error
func (s *LogoutTestSuite) TestMockLogoutUser500InternalServerError() {

	expectedTTL := time.Until(time.Unix(fixedUnix, 0))

	redisClient, mockRedisClient, _ := utils.GetRedisClient(ctx, false)
	router := getLogoutRouter(redisClient)

	mockRedisClient.ExpectSet(key, 1, time.Duration(expectedTTL)).SetErr(errors.New(genericConstants.ErrInternalServer))

	req, _ := http.NewRequest(http.MethodPost, constants.LogoutTest, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	s.Equal(http.StatusInternalServerError, w.Code)
	s.Contains(w.Body.String(), genericConstants.ErrInternalServer)
	s.NoError(mockRedisClient.ExpectationsWereMet())
}

func TestLogoutSuite(t *testing.T) {
	suite.Run(t, new(LogoutTestSuite))
}
