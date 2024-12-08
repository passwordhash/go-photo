package config

import "time"

const (
	PhotosDir = "photos"
	LogsDir   = "logs"
)

const (
	PostgresDefaultHost    = "localhost"
	PostgresDefaultPort    = "5432"
	PostgresDefaultSSLMode = "disable"
)

// TEMP
const DefaultUsersFoldername = "/home"

const DefaultContextTimeout = time.Duration(time.Second * 5)
