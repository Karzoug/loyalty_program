package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUser_VerifyPassword(t *testing.T) {
	u, err := New("Ylönen", "dangerous things")
	assert.NoError(t, err)

	assert.True(t, u.VerifyPassword("dangerous things"))
	assert.False(t, u.VerifyPassword("ultimate survivor"))
}

func TestNew(t *testing.T) {
	_, err := New("Ylönen", "")
	assert.Error(t, err)

	tests := []struct {
		name     string
		login    Login
		password string
		wantErr  bool
	}{
		{
			name:     "negative: empty password",
			login:    "Ylönen",
			password: "",
			wantErr:  true,
		},
		{
			name:     "negative: empty password",
			login:    "",
			password: "HiGjm*FC5aGD0a9jY6B19*@imz&9ltf^",
			wantErr:  true,
		},
		{
			name:     "negative: too long password",
			login:    "Jezebel",
			password: "w69rNZe*k7V*3n^loTek*K5HiGjm*FC5aGD0a9jY6B19*@imz&9ltf^1fYwhA!Vn6!OdrMzISAHkQ!L41gb2&nXScz",
			wantErr:  true,
		},
		{
			name:     "positive",
			login:    "Jezebel",
			password: "ltf^1fYwhA!Vn6!OdrMzISAHkQ!L41",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.login, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
