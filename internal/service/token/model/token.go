package model

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserUUID string `json:"uuid"`
	jwt.RegisteredClaims
}
