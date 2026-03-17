package constants

type Actiontype string

const (
	AddAction Actiontype = "ADD"
	DelAction Actiontype = "DEL"
	GetAction Actiontype = "GET"
)

const (
	UsersTableName = "users"
	Username       = "username = ?"
	UserId         = "user_id = ?"
)

const SwaggerTitle = "Stock Broker Application API"
const SwaggerRoute = "/swagger/*any"

const (
	ServiceName       = "adg"
	PortDefaultValude = 8080
)
