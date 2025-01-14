package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type JWTAuthMiddleware struct {
	JWTSigningKey string
}

func (m *JWTAuthMiddleware) JWTAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			t := strings.Split(authHeader, " ")
			if len(t) != 2 {
				return echo.ErrUnauthorized
			}

			authToken := t[1]
			authorized, err := IsAuthorized(authToken, m.JWTSigningKey)
			if err != nil || !authorized {
				return err
			}

			_, err = ExtractIDFromToken(authToken, m.JWTSigningKey)
			if err != nil {
				return err
			}

			return next(c)
		}
	}
}

func InitJWTAuthMiddleware(jwtkey string) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{JWTSigningKey: jwtkey}
}

type BasicAuthMiddleware struct {
	Credentials string
}

func (m *BasicAuthMiddleware) BasicAuthMiddleware() echo.MiddlewareFunc {
	return echoMiddleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		decodedCreds, err := base64.StdEncoding.DecodeString(m.Credentials)
		if err != nil {
			return false, err
		}

		creds := strings.SplitN(string(decodedCreds), ":", 2)
		if len(creds) != 2 {
			return false, nil
		}

		// Be careful to use constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(creds[0])) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(creds[1])) == 1 {
			return true, nil
		}
		return false, nil
	})
}

func InitBasicAuthMiddleware(credentials string) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{Credentials: credentials}
}

func ExtractIDFromToken(requestToken string, secret string) (string, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}

	return claims["id"].(string), nil
}

func IsAuthorized(requestToken string, secret string) (bool, error) {
	_, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
