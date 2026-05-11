package router

import (
	"authentication/business"
	"authentication/commons/constants"
	"authentication/docs"
	"authentication/handlers"
	"authentication/middleware"
	"authentication/repository"

	genericConstants "stock_broker_application/src/constants"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	files "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func GetRouter(db *gorm.DB, redisClient *redis.Client) *gin.Engine {
	router := gin.New()
	router.Use(middleware.LoggerMiddleware())
	router.Use(gin.Recovery())

	docs.SwaggerInfo.Title = constants.SwaggerTitle

	router.GET(constants.SwaggerRoute, ginSwagger.WrapHandler(files.Handler))

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{genericConstants.AllowedOrigin},
		AllowMethods: []string{genericConstants.POST, genericConstants.GET},
		AllowHeaders: []string{genericConstants.Origin, genericConstants.ContentType, genericConstants.Authorization},
	}))

	createUserRepository := repository.NewCreateUserRepository(db)
	createUserService := business.NewCreateUserService(createUserRepository, db)
	createUserHandler := handlers.NewCreateUserHandler(createUserService)

	signInRepository := repository.NewSignInRepository(db)
	signInService := business.NewSignInService(signInRepository)
	signInHandler := handlers.NewSignInHandler(signInService)

	// postgresClient := utils.GetPostgresClient().GormDB
	verifyUserOtpRepository := repository.NewValidateUserOtpRepository(db, redisClient)
	verifyUserOtpService := business.NewValidateUserOtpService(verifyUserOtpRepository, db, redisClient)
	verifyUserOtpHandler := handlers.NewValidateUserOtpHandler(verifyUserOtpService)

	changePasswordRepository := repository.NewChangePasswordRepository()
	changePasswordService := business.NewChangePasswordService(changePasswordRepository, db)
	changePasswordHandler := handlers.NewChangePasswordHandler(changePasswordService)

	logoutService := business.NewLogoutService(redisClient)
	logoutHandler := handlers.NewLogoutHandler(logoutService)

	authGroup := router.Group(constants.AuthRoutePrefix)
	{
		authGroup.POST(constants.Signup, createUserHandler.HandleCreaterUser)
		authGroup.POST(constants.Signin, signInHandler.HandleSignIn)
		authGroup.POST(constants.Validateotp, verifyUserOtpHandler.HandleValidateUserOtp)
		authGroup.POST(constants.Changepassword, middleware.AuthMiddleware(redisClient), changePasswordHandler.HandleChangePassword)
		authGroup.POST(constants.Logout, middleware.AuthMiddleware(redisClient), logoutHandler.Logout)
	}

	return router
}
