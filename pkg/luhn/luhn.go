package luhn

// Valid checks a number is valid or not based on Luhn algorithm.
func Valid[T int64 | int32 | int](number T) bool {
	remainder := number % 10
	checksum := Checksum(number / 10)

	return (remainder+checksum)%10 == 0
}

func Checksum[T int64 | int32 | int](number T) T {
	var luhn T
	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
