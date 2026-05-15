package constants

//Authentications API URL Keys
const (
	ServiceName       = "authentication"
	PortDefaultValude = 8081
)

// Database table name & field names for users
const (
	UsersTableName = "users"
	Fieldemail     = "email"
	Username       = "username = ?"
)

// Success message for user
const (
	UserCreationSuccessMsg    = "User created successfully"
	UserLoggedInSuccessMsg    = "User logged in successfully"
	OtpValidatedSuccessMsg    = "OTP validated successfully"
	PasswordChangedSuccessMsg = "Password changed succefully"
	LogoutSuccessfulMsg       = "User logged out successfully"
)

//Swagger Titile
const SwaggerTitle = "Stock Broker Application API"

const EmailorPasswordField = "email_or_password"

//Cookies
const (
	Name     = "refresh_token"
	Time     = 30 * 24 * 60 * 60
	Path     = "/"
	Domain   = ""
	Secure   = true
	HttpOnly = true
)

//test queries
const (
	UserByEmailTestQuery          = `SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`
	UserByEmailIncorrectTestQuery = `SELECT * FROM "users" WHERE username = $1 ORDER "users"."id" LIMIT $2` //doesn't include "BY"

	CreateUserTestQuery          = `INSERT INTO "users"`
	CreateUserIncorrectTestQuery = `INSERT "users"`

	// ValidateUserOTPTestQuery = `SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`
	ValidateUserOTPTestQuery = `SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`
)
