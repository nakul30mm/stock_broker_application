package models

type PostgresConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	Timezone string
	SSLMode  string
}
