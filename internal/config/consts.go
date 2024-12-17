package config

import "time"

const (
	DefaultStorageFolderPath = "./storage"
	LogsDir                  = "logs"
)

const (
	RSAPublicKeyDefaultTTL = time.Hour * 1
)

const (
	PostgresDefaultHost    = "localhost"
	PostgresDefaultPort    = "5432"
	PostgresDefaultSSLMode = "disable"
)

const DefaultUsersFoldername = "/home"

const DefaultContextTimeout = time.Duration(time.Second * 5)
