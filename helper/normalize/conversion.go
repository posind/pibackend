package normalize

import "strings"

func digitToLetter(digit int) string {
	return string('A' + digit - 1)
}

func NumberToAlphabet(num int) string {
	if num <= 0 {
		return ""
	}

	var result []string
	for num > 0 {
		digit := num % 10
		result = append([]string{digitToLetter(digit)}, result...)
		num /= 10
	}

	return strings.Join(result, "")
}
