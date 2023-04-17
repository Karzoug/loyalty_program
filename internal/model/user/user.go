package user

import (
	"errors"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrBadPassword = errors.New("bad password: too long (> 72 bytes)")
)

type User struct {
	Login             Login
	EncryptedPassword string
	Balance           decimal.Decimal
}

func New(login Login, password string) (*User, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrBadPassword
	}

	if !login.Valid() {
		return nil, ErrInvalidLogin
	}

	return &User{
		Login:             login,
		EncryptedPassword: string(encpw),
		Balance:           decimal.Decimal{},
	}, nil
}

func (u User) VerifyPassword(psw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(psw)) == nil
}
