package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Atmosfr/user-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"user_role"`
	jwt.RegisteredClaims
}

var JwtSecret []byte

func InitJWT(secret string) error {
	if secret == "" {
		return errors.New("jwt secret is empty")
	}
	JwtSecret = []byte(secret)
	return nil
}


func GenerateToken(user *models.User, duration time.Duration, jwtSecret []byte) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("jwt secret is empty")
	}

	if user == nil {
		return "", errors.New("user is nil")
	}

	if user.ID < 1 {
		return "", errors.New("invalid user ID")
	}

	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	return tokenString, err
}

func ValidateToken(tokenString string, jwtSecret []byte) (*models.User, error) {
	if len(jwtSecret) == 0 {
		return nil, errors.New("jwt secret is empty")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.UserID == 0 {
		return nil, errors.New("missing or invalid user_id claim")
	}

	if claims.IssuedAt != nil && claims.IssuedAt.Unix() > time.Now().Unix() {
		return nil, jwt.ErrTokenNotValidYet
	}

	user := &models.User{
		ID:   claims.UserID,
		Role: claims.Role,
	}

	return user, nil
}
