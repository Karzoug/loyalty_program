package user

import (
	"testing"
)

func TestLogin_Valid(t *testing.T) {
	tests := []struct {
		name  string
		login Login
		want  bool
	}{
		{
			name:  "positive",
			login: "jezebel",
			want:  true,
		},
		{
			name:  "negative: too long",
			login: "oxHanGthuMNalTERMOZAKortIosembenapAReyERMideLtHatEgintaBLANtAterWOMbANDialmErMaLERTInAbLEcTiOngerMatCIoNITNIaNXIAlt",
			want:  false,
		},
		{
			name:  "negative: empty",
			login: "",
			want:  false,
		},
		{
			name:  "negative: not valid UTF-8-encoded runes",
			login: Login([]byte{0xff, 0xfe, 0xfd}),
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.login.Valid(); got != tt.want {
				t.Errorf("Login.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
