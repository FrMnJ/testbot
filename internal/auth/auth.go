package auth

import (
	"strconv"
	"strings"

	jwt "github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	FullName  string   `json:"full_name"`
	UserEmail string   `json:"user_email"`
	Code      string   `json:"code"`
	Roles     []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenerateTokenForScenario(role string, jwtSecret string) (string, error) {
	claims := CustomClaims{
		FullName:  strings.ToUpper(role),
		UserEmail: role + "@example.com",
		Code:      "123456",
		Roles:     []string{role},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.FormatInt(1, 10),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
