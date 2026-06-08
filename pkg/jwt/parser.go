package jwt

import (
	"github.com/golang-jwt/jwt/v5"
)

// ParseTokenLenient parses a JWT token without validating expiry.
// Used for refresh token endpoints where the token may be near or past expiry.
func ParseTokenLenient(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
