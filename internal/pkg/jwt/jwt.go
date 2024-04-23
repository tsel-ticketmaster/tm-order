package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v4"
)

// Errors.
var (
	ErrInvalidToken      error = fmt.Errorf("invalid token")
	ErrExpiredOrNotReady error = fmt.Errorf("token is either expired or not ready to use")
)

// JSONWebToken is a concrete struct of json web token.
type JSONWebToken struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewJSONWebToken is a constructor.
func NewJSONWebToken(privateKey, publicKey []byte) *JSONWebToken {
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		log.Println(err)
	}

	rsaPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		log.Println(err)
	}
	return &JSONWebToken{rsaPrivateKey, rsaPublicKey}
}

// Sign will generate new jwt token.
func (a *JSONWebToken) Sign(ctx context.Context, claims jwt.Claims) (tokenString string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(a.privateKey)
}

// Parse will parse the token string to bearer claims.
func (a *JSONWebToken) Parse(ctx context.Context, tokenString string, claims jwt.Claims) (err error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, a.keyFunc)
	if err = a.checkError(err); err != nil {
		return
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return
}

func (a *JSONWebToken) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, ErrInvalidToken
	}
	return a.publicKey, nil
}

func (a *JSONWebToken) checkError(err error) error {
	if err == nil {
		return err
	}

	ve, ok := err.(*jwt.ValidationError)
	if !ok {
		return ErrInvalidToken
	}
	if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
		return ErrExpiredOrNotReady
	}

	return ErrInvalidToken
}
