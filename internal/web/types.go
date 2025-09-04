package web

import (
	"github.com/golang-jwt/jwt/v5"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
)

type Role string

const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

type UserLoginRequest struct {
	Username string
	Password string
}

type PageData struct {
	Leases []dhcp.Lease
}

type User struct {
	Username string
	Role     Role
}

type UserClaims struct {
	User
	jwt.RegisteredClaims
}
