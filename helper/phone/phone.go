package phone

import "regexp"

func MaskPhoneNumber(phone string) string {
	// Regular expression untuk menangkap tiga bagian: enam digit pertama, tiga digit di tengah, dan sisanya.
	re := regexp.MustCompile(`(\d{6})(\d{3})(\d+)`)
	return re.ReplaceAllString(phone, `$1xxx$3`)
}
