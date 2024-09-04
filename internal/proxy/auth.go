package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (server *ProxyServer) validateKey(req *http.Request) error {
	if server.verificationCert == nil {
		return nil
	}

	authorizationHeader := req.Header.Get("Authorization")
	if authorizationHeader == "" {
		return fmt.Errorf("no auth header found")
	}
	tokenString := strings.Split(authorizationHeader, " ")[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return server.verificationCert, nil
	})
	if err != nil {
		return err
	}

	if token.Valid {
		return nil
	}

	return fmt.Errorf("invalud token")
}
