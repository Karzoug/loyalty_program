package user

import (
	"fmt"
	"unicode/utf8"
)

const (
	MaxRuneCountInLogin = 100
)

var (
	ErrInvalidLogin = fmt.Errorf("invalid user login: contains not valid UTF-8-encoded runes or rune count not in (0; %d]", MaxRuneCountInLogin)
)

type Login string

func (l Login) Valid() bool {
	s := string(l)
	c := utf8.RuneCountInString(s)
	if !utf8.ValidString(s) || c > MaxRuneCountInLogin || c == 0 {
		return false
	}
	return true
}
