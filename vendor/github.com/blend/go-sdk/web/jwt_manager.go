package web

import (
	"context"
	"time"

	"github.com/blend/go-sdk/exception"
	"github.com/blend/go-sdk/jwt"
)

const (
	// ErrJWTNonstandardClaims can be returned by the jwt manager keyfunc.
	ErrJWTNonstandardClaims = exception.Class("jwt; invalid claims object; should be standard claims")
)

// NewJWTManager returns a new jwt manager from a key.
func NewJWTManager(key []byte) *JWTManager {
	return &JWTManager{
		KeyProvider: func(_ *Session) ([]byte, error) {
			return key, nil
		},
	}
}

// JWTManager is a manager for JWTs.
type JWTManager struct {
	KeyProvider func(*Session) ([]byte, error)
}

// Claims returns the sesion as a JWT standard claims object.
func (jwtm JWTManager) Claims(session *Session) *jwt.StandardClaims {
	return &jwt.StandardClaims{
		ID:        session.SessionID,
		Audience:  session.BaseURL,
		Issuer:    "go-web",
		Subject:   session.UserID,
		IssuedAt:  session.CreatedUTC.Unix(),
		ExpiresAt: session.ExpiresUTC.Unix(),
	}
}

// FromClaims returns a session from a given claims set.
func (jwtm JWTManager) FromClaims(claims *jwt.StandardClaims) *Session {
	return &Session{
		SessionID:  claims.ID,
		BaseURL:    claims.Audience,
		UserID:     claims.Subject,
		CreatedUTC: time.Unix(claims.IssuedAt, 0).In(time.UTC),
		ExpiresUTC: time.Unix(claims.ExpiresAt, 0).In(time.UTC),
	}
}

// KeyFunc is a shim function to get the key for a given token.
func (jwtm JWTManager) KeyFunc(token *jwt.Token) (interface{}, error) {
	typed, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, ErrJWTNonstandardClaims
	}
	return jwtm.KeyProvider(jwtm.FromClaims(typed))
}

// SerializeSessionValueHandler is a shim to the auth manager.
func (jwtm JWTManager) SerializeSessionValueHandler(_ context.Context, session *Session, _ State) (output string, err error) {
	var key []byte
	key, err = jwtm.KeyProvider(session)
	if err != nil {
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHMAC512, jwtm.Claims(session))
	output, err = token.SignedString(key)
	return
}

// ParseSessionValueHandler is a shim to the auth manager.
func (jwtm JWTManager) ParseSessionValueHandler(_ context.Context, sessionValue string, _ State) (*Session, error) {
	var claims jwt.StandardClaims
	_, err := jwt.ParseWithClaims(sessionValue, &claims, jwtm.KeyFunc)
	if err != nil {
		return nil, err
	}

	// do we check if the token is valid ???
	return jwtm.FromClaims(&claims), nil
}
