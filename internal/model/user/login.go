package user

import (
	"fmt"
	"unicode/utf8"
)

const (
	MaxRuneCountInLogin = 100
)

var (
	ErrInvalidLogin = fmt.Errorf("invalid user login: contains not valid UTF-8-encoded runes or rune count exceeds %d", MaxRuneCountInLogin)
)

type Login string

func (l Login) Valid() bool {
	s := string(l)
	if !utf8.ValidString(s) || utf8.RuneCountInString(s) > MaxRuneCountInLogin {
		return false
	}
	return true
}
