package luhn

import "testing"

func TestValid(t *testing.T) {
	testsInt64 := []struct {
		name   string
		number int64
		want   bool
	}{
		{
			name:   "invalid #1 (int64)",
			number: 1234567812345678,
			want:   false,
		},
		{
			name:   "invalid #2 (int64)",
			number: 676196000000551043,
			want:   false,
		},
		{
			name:   "valid #1 (int64)",
			number: 676196000029070555,
			want:   true,
		},
		{
			name:   "valid #2 (int64)",
			number: 676196000000551045,
			want:   true,
		},
	}
	for _, tt := range testsInt64 {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.number); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
	testsInt32 := []struct {
		name   string
		number int32
		want   bool
	}{
		{
			name:   "invalid #1 (int32)",
			number: 1234567816,
			want:   false,
		},
		{
			name:   "invalid #2 (int32)",
			number: 1234557890,
			want:   false,
		},
		{
			name:   "valid #1 (int32)",
			number: 1884567890,
			want:   true,
		},
		{
			name:   "valid #2 (int32)",
			number: 1234567814,
			want:   true,
		},
	}
	for _, tt := range testsInt32 {
		t.Run(tt.name, func(t *testing.T) {
			if got := Valid(tt.number); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
