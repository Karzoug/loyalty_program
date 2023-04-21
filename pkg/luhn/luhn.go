package luhn

// Valid checks a number is valid or not based on Luhn algorithm.
func Valid[T int64 | int32 | int](number T) bool {
	quotient := number / 10
	remainder := number % 10

	var luhn T
	for i := 0; quotient > 0; i++ {
		cur := quotient % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		quotient = quotient / 10
	}
	checksum := luhn % 10

	return (remainder+checksum)%10 == 0
}
