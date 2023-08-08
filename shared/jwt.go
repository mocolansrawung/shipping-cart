package shared

import (
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	jwt.StandardClaims
}
