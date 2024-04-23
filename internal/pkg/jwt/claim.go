package jwt

import "github.com/golang-jwt/jwt/v4"

type Claim struct {
	jwt.StandardClaims
	Name  string
	Email string
	Type  string
}
